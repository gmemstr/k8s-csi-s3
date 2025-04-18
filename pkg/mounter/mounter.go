package mounter

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"git.gmem.ca/arch/k8s-csi-s3/pkg/s3"
	systemd "github.com/coreos/go-systemd/v22/dbus"
	"github.com/golang/glog"
	"github.com/mitchellh/go-ps"
	"k8s.io/mount-utils"
)

// Mounter interface which can be implemented
// by the different mounter types
type Mounter interface {
	Mount(target, volumeID string) error
}

const (
	s3fsMounterType     = "s3fs"
	geesefsMounterType  = "geesefs"
	tigrisfsMounterType = "tigrisfs"
	rcloneMounterType   = "rclone"
	TypeKey             = "mounter"
	BucketKey           = "bucket"
	OptionsKey          = "options"
)

// New returns a new mounter depending on the mounterType parameter.
// TigrisFS is a fork of GeeseFS, so we can use the same methods and just swap
// the binaries.
func New(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	mounter := meta.Mounter
	// Fall back to mounterType in cfg
	if len(meta.Mounter) == 0 {
		mounter = cfg.Mounter
	}
	switch mounter {
	case geesefsMounterType:
		return newTigrisFSMounter(meta, cfg, geesefsMounterType)

	case tigrisfsMounterType:
		return newTigrisFSMounter(meta, cfg, tigrisfsMounterType)

	case s3fsMounterType:
		return newS3fsMounter(meta, cfg)

	case rcloneMounterType:
		return newRcloneMounter(meta, cfg)

	default:
		// default to GeeseFS
		return newTigrisFSMounter(meta, cfg, tigrisfsMounterType)
	}
}

func fuseMount(path string, command string, args []string, envs []string) error {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	// cmd.Environ() returns envs inherited from the current process
	cmd.Env = append(cmd.Environ(), envs...)
	glog.V(3).Infof("mounting fuse with command: %s and args: %s", command, args)

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error fuseMount command: %s\nargs: %s\noutput: %s", command, args, out)
	}

	return waitForMount(path, 10*time.Second)
}

func Unmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	return nil
}

func SystemdUnmount(volumeID string) (bool, error) {
	ctx := context.Background()
	conn, err := systemd.NewWithContext(ctx)
	if err != nil {
		glog.Errorf("failed to connect to systemd dbus service: %v", err)
		return false, err
	}
	defer conn.Close()
	unitName := "geesefs-" + systemd.PathBusEscape(volumeID) + ".service"
	units, err := conn.ListUnitsByNamesContext(ctx, []string{unitName})
	glog.Errorf("got %v", units)
	if err != nil {
		glog.Errorf("failed to list systemd unit by name %v: %v", unitName, err)
		return false, err
	}
	if len(units) == 0 || units[0].ActiveState == "inactive" || units[0].ActiveState == "failed" {
		return true, nil
	}

	resCh := make(chan string)
	defer close(resCh)

	_, err = conn.StopUnitContext(ctx, unitName, "replace", resCh)
	if err != nil {
		glog.Errorf("failed to stop systemd unit (%s): %v", unitName, err)
		return false, err
	}

	res := <-resCh // wait until is stopped
	glog.Infof("systemd unit is stopped with result (%s): %s", unitName, res)

	return true, nil
}

func FuseUnmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// as fuse quits immediately, we will try to wait until the process is done
	process, err := FindFuseMountProcess(path)
	if err != nil {
		glog.Errorf("error getting PID of fuse mount: %s", err)
		return nil
	}
	if process == nil {
		glog.Warningf("unable to find PID of fuse mount %s, it must have finished already", path)
		return nil
	}
	glog.Infof("found fuse pid %v of mount %s, checking if it still runs", process.Pid, path)
	return waitForProcess(process, 20)
}

func waitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		isMount, err := mount.New("").IsMountPoint(path)
		if err != nil {
			return err
		}
		if isMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return errors.New("timeout waiting for mount")
		}
	}
}

func FindFuseMountProcess(path string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		cmdLine, err := getCmdLine(p.Pid())
		if err != nil {
			glog.Errorf("unable to get cmdline of PID %v: %s", p.Pid(), err)
			continue
		}
		if strings.Contains(cmdLine, path) {
			glog.Infof("found matching pid %v on path %s", p.Pid(), path)
			return os.FindProcess(p.Pid())
		}
	}
	return nil, nil
}

func waitForProcess(p *os.Process, limit int) error {
	for backoff := 0; backoff < limit; backoff++ {
		cmdLine, err := getCmdLine(p.Pid)
		if err != nil {
			glog.Warningf("error checking cmdline of PID %v, assuming it is dead: %s", p.Pid, err)
			_, err := p.Wait()
			if err != nil {
				return err
			}
			return nil
		}
		if cmdLine == "" {
			glog.Warning("fuse process seems dead, returning")
			_, err := p.Wait()
			if err != nil {
				return err
			}
			return nil
		}
		if err := p.Signal(syscall.Signal(0)); err != nil {
			glog.Warningf("fuse process does not seem active or we are unprivileged: %s", err)
			_, err := p.Wait()
			if err != nil {
				return err
			}
			return nil
		}
		glog.Infof("fuse process with PID %v still active, waiting...", p.Pid)
		time.Sleep(time.Duration(math.Pow(1.5, float64(backoff))*100) * time.Millisecond)
	}
	err := p.Release()
	if err != nil {
		return err
	}
	return fmt.Errorf("timeout waiting for PID %v to end", p.Pid)
}

func getCmdLine(pid int) (string, error) {
	cmdLineFile := fmt.Sprintf("/proc/%v/cmdline", pid)
	cmdLine, err := os.ReadFile(cmdLineFile)
	if err != nil {
		return "", err
	}
	return string(cmdLine), nil
}
