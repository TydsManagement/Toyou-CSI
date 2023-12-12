/*
Copyright (C) 2023 Toyou, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rpcserver

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"toyou_csi/pkg/driver"
	"toyou_csi/pkg/service"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// Run initializes and starts the CSI driver's gRPC server.
func Run(driver *driver.ToyouDriver, tydsManager service.TydsManager, mounter *mount.SafeFormatAndMount, endpoint string) {
	// Listen on the specified endpoint
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	server := grpc.NewServer()

	// Register the CSI services
	csi.RegisterIdentityServer(server, NewIdentityServer(driver))
	csi.RegisterControllerServer(server, NewControllerServer(driver, tydsManager))
	csi.RegisterNodeServer(server, NewNodeServer(driver, tydsManager, mounter))

	// Start serving incoming connections
	go func() {
		klog.Infof("Starting gRPC server on endpoint %s", endpoint)
		if err := server.Serve(listener); err != nil {
			klog.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	klog.Info("Shutting down gRPC server")
	server.GracefulStop()
}
