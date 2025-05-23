/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
)

type Driver struct {
	endpoint string
	nodeid   string

	cap []*csi.ControllerServiceCapability
	vc  []*csi.VolumeCapability_AccessMode

	csi.UnimplementedControllerServer
	csi.UnimplementedNodeServer
	csi.UnimplementedIdentityServer
	csi.UnimplementedGroupControllerServer
	csi.UnimplementedSnapshotMetadataServer
}

var (
	vendorVersion = "v1.34.7"
	driverName    = "ca.gmem.s3.csi"
)

// New initializes the driver
func New(nodeID string, endpoint string) (*Driver, error) {
	s3Driver := &Driver{
		nodeid:   nodeID,
		endpoint: endpoint,
	}
	return s3Driver, nil
}

func (d *Driver) Run() {
	glog.Infof("driver: %v ", driverName)
	glog.Infof("Version: %v ", vendorVersion)
	// Initialize default library driver

	d.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	d.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	})

	s := NewNonBlockingGRPCServer()
	s.Start(d.endpoint, d, d, d, d, d)
	s.Wait()
}

func (d *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability //nolint:prealloc

	for _, c := range cl {
		glog.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}

	d.cap = csc
}

func (d *Driver) AddVolumeCapabilityAccessModes(
	vc []csi.VolumeCapability_AccessMode_Mode,
) []*csi.VolumeCapability_AccessMode {
	var vca []*csi.VolumeCapability_AccessMode //nolint:prealloc

	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	d.vc = vca
	return vca
}

func (d *Driver) GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	return d.vc
}
