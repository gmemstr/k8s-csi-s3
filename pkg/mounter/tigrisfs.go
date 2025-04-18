package mounter

import (
	"cmp"
	"fmt"
	"os"
	"strings"
	"time"

	"git.gmem.ca/arch/k8s-csi-s3/pkg/s3"
	systemd "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

// Implements Mounter
type tigrisfsMounter struct {
	meta            *s3.FSMeta
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
	binary          string
}

func newTigrisFSMounter(meta *s3.FSMeta, cfg *s3.Config, binary string) (Mounter, error) {
	return &tigrisfsMounter{
		meta:            meta,
		endpoint:        cfg.Endpoint,
		region:          cfg.Region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
		binary:          binary,
	}, nil
}

func (tigrisfs *tigrisfsMounter) CopyBinary(from, to string) error {
	st, err := os.Stat(from)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %v", from, err)
	}
	st2, err := os.Stat(to)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat %s: %v", to, err)
	}
	if err != nil || st2.Size() != st.Size() || st2.ModTime() != st.ModTime() {
		if err == nil {
			// remove the file first to not hit "text file busy" errors
			err = os.Remove(to)
			if err != nil {
				return fmt.Errorf("error removing %s to update it: %v", to, err)
			}
		}
		bin, err := os.ReadFile(from)
		if err != nil {
			return fmt.Errorf("error copying %s to %s: %v", from, to, err)
		}
		err = os.WriteFile(to, bin, 0755)
		if err != nil {
			return fmt.Errorf("error copying %s to %s: %v", from, to, err)
		}
		err = os.Chtimes(to, st.ModTime(), st.ModTime())
		if err != nil {
			return fmt.Errorf("error copying %s to %s: %v", from, to, err)
		}
	}
	return nil
}

func (tigrisfs *tigrisfsMounter) MountDirect(target string, args []string) error {
	args = append([]string{
		"--endpoint", tigrisfs.endpoint,
		"-o", "allow_other",
		"--log-file", "/dev/stderr",
	}, args...)
	envs := []string{
		"AWS_ACCESS_KEY_ID=" + tigrisfs.accessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + tigrisfs.secretAccessKey,
	}
	return fuseMount(target, tigrisfs.binary, args, envs)
}

func (tigrisfs *tigrisfsMounter) Mount(target, volumeID string) error {
	ctx := context.Background()
	fullPath := fmt.Sprintf("%s:%s", tigrisfs.meta.BucketName, tigrisfs.meta.Prefix)
	var args []string
	if tigrisfs.region != "" {
		args = append(args, "--region", tigrisfs.region)
	}
	args = append(
		args,
		"--setuid", "65534", // nobody. drop root privileges
		"--setgid", "65534", // nogroup
	)
	useSystemd := true
	for i := 0; i < len(tigrisfs.meta.MountOptions); i++ {
		opt := tigrisfs.meta.MountOptions[i]
		if len(opt) == 0 {
			continue
		}
		if opt == "--no-systemd" {
			useSystemd = false
			continue
		}
		if len(opt) > 0 && opt[0] != '-' {
			args = append(args, opt)
			continue
		}
		// Remove unsafe options
		s := 1
		if len(opt) > 1 && opt[1] == '-' {
			s++
		}
		key := opt[s:]
		e := strings.Index(opt, "=")
		if e >= 0 {
			key = opt[s:e]
		}
		// If we aren't using systemd, we consider these flags "safe".
		if (key == "log-file" || key == "shared-config" || key == "cache") && useSystemd {
			args = append(args, opt)
		} else if key != "" {
			args = append(args, opt)
		}
	}
	args = append(args, fullPath, target)
	if !useSystemd {
		return tigrisfs.MountDirect(target, args)
	}
	return tigrisfs.setupSystemdMount(ctx, volumeID, target, args)
}

func (tigrisfs *tigrisfsMounter) setupSystemdMount(ctx context.Context, volumeID, target string, args []string) error {
	conn, err := systemd.NewWithContext(ctx)
	if err != nil {
		glog.Errorf("failed to connect to systemd dbus service: %v, starting tigrisfs directly", err)
		return tigrisfs.MountDirect(target, args)
	}
	defer conn.Close()
	// systemd is present
	if err = tigrisfs.CopyBinary(
		fmt.Sprintf("/usr/bin/%s", tigrisfs.binary),
		fmt.Sprintf("/csi/%s", tigrisfs.binary)); err != nil {
		return err
	}
	pluginDir := cmp.Or(os.Getenv("PLUGIN_DIR"), "/var/lib/kubelet/plugins/ca.gmem.s3.csi")
	args = append([]string{pluginDir + "/tigrisfs", "-f", "-o", "allow_other", "--endpoint", tigrisfs.endpoint}, args...)
	glog.Info("starting s3 mount using systemd: " + strings.Join(args, " "))
	unitName := fmt.Sprintf("%s-%s.service", tigrisfs.binary, systemd.PathBusEscape(volumeID))
	newProps := []systemd.Property{
		{
			Name:  "Description",
			Value: dbus.MakeVariant("TigrisFS mount for Kubernetes volume " + volumeID),
		},
		systemd.PropExecStart(args, false),
		{
			Name: "Environment",
			Value: dbus.MakeVariant([]string{
				fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", tigrisfs.accessKeyID),
				fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", tigrisfs.secretAccessKey),
			}),
		},
		{
			Name:  "CollectMode",
			Value: dbus.MakeVariant("inactive-or-failed"),
		},
	}
	unitProps, err := conn.GetAllPropertiesContext(ctx, unitName)
	if err == nil {
		// Unit already exists
		if s, ok := unitProps["ActiveState"].(string); ok && (s == "active" || s == "activating" || s == "reloading") {
			// Unit is already active
			curPath := ""
			prevExec, ok := unitProps["ExecStart"].([][]interface{})
			if ok && len(prevExec) > 0 && len(prevExec[0]) >= 2 {
				execArgs, ok := prevExec[0][1].([]string)
				if ok && len(execArgs) >= 2 {
					curPath = execArgs[len(execArgs)-1]
				}
			}
			if curPath != target {
				// FIXME This may mean that the same bucket&path are used for multiple PVs. Support it somehow
				return fmt.Errorf(
					"tigrisFS for volume %v is already mounted on host, but"+
						" in a different directory. We want %v, but it's in %v",
					volumeID, target, curPath,
				)
			}
			// Already mounted at right location, wait for mount
			return waitForMount(target, 30*time.Second)
		} else {
			// Stop and garbage collect the unit if automatic collection didn't work for some reason
			_, err := conn.StopUnitContext(ctx, unitName, "replace", nil)
			if err != nil {
				return err
			}
			err = conn.ResetFailedUnitContext(ctx, unitName)
			if err != nil {
				return err
			}
		}
	}
	unitPath := "/run/systemd/system/" + unitName + ".d"
	err = os.MkdirAll(unitPath, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %v", unitPath, err)
	}
	// force & lazy unmount to cleanup possibly dead mountpoints
	err = os.WriteFile(
		unitPath+"/50-StopProps.conf",
		[]byte("[Service]\nExecStopPost=/bin/umount -f -l "+target+"\nTimeoutStopSec=20\n"),
		0600,
	)
	if err != nil {
		return fmt.Errorf("error writing %v/50-ExecStopPost.conf: %v", unitPath, err)
	}
	_, err = conn.StartTransientUnitContext(ctx, unitName, "replace", newProps, nil)
	if err != nil {
		return fmt.Errorf("error starting systemd unit %s on host: %v", unitName, err)
	}
	return waitForMount(target, 30*time.Second)
}
