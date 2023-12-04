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
	"context"

	"toyou_csi/pkg/driver"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// IdentityServer implements the IdentityServer CSI gRPC interface.
type IdentityServer struct {
	Driver *driver.ToyouDriver
}

// NewIdentityServer creates a new IdentityServer.
func NewIdentityServer(d *driver.ToyouDriver) *IdentityServer {
	return &IdentityServer{
		Driver: d,
	}
}

// Probe checks the health and readiness of the service.
func (is *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.Info("Probe called")
	// Implement your health check logic here.
	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{Value: true},
	}, nil
}

// GetPluginInfo returns metadata of the plugin.
func (is *IdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.Info("GetPluginInfo called")

	if is.Driver.GetName() == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if is.Driver.GetVersion() == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          is.Driver.GetName(),
		VendorVersion: is.Driver.GetVersion(),
	}, nil
}

// GetPluginCapabilities returns the capabilities of the plugin.
func (is *IdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.Info("GetPluginCapabilities called")

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: is.Driver.GetPluginCapability(),
	}, nil
}
