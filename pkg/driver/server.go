/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func NewNonBlockingGRPCServer() *NonBlockingGRPCServer {
	return &NonBlockingGRPCServer{}
}

// NonBlockingGRPCServer server
type NonBlockingGRPCServer struct {
	wg      sync.WaitGroup
	server  *grpc.Server
	cleanup func()
}

func (s *NonBlockingGRPCServer) Start(
	endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer,
	gcs csi.GroupControllerServer, sms csi.SnapshotMetadataServer,
) {

	s.wg.Add(1)

	go s.serve(endpoint, ids, cs, ns, gcs, sms)
}

func (s *NonBlockingGRPCServer) Wait() {
	s.wg.Wait()
}

func (s *NonBlockingGRPCServer) Stop() {
	s.server.GracefulStop()
	s.cleanup()
}

func (s *NonBlockingGRPCServer) ForceStop() {
	s.server.Stop()
	s.cleanup()
}

func (s *NonBlockingGRPCServer) serve(
	ep string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer,
	gcs csi.GroupControllerServer, sms csi.SnapshotMetadataServer,
) {
	proto, addr, err := ParseEndpoint(ep)
	if err != nil {
		fmt.Printf("unable to parse addr %s: %s\n", addr, err)
		return
	}

	cleanup := func() {}
	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return
		}
		cleanup = func() {
			_ = os.Remove(addr)
		}
	}

	l, err := net.Listen(proto, addr)
	if err != nil {
		fmt.Printf("unable to bind to addr %s: %s", addr, err)
		return
	}

	var opts []grpc.ServerOption
	server := grpc.NewServer(opts...)
	s.server = server
	s.cleanup = cleanup

	if ids != nil {
		csi.RegisterIdentityServer(server, ids)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}
	if gcs != nil {
		csi.RegisterGroupControllerServer(server, gcs)
	}
	if sms != nil {
		csi.RegisterSnapshotMetadataServer(server, sms)
	}

	err = server.Serve(l)
	if err != nil {
		fmt.Printf("unable to serve: %s\n", err)
	}

}
