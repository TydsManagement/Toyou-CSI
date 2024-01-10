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

package driver

import (
	"toyou-csi/pkg/service"

	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/klog"
	"k8s.io/utils/mount"
)

// Run initializes and starts the CSI driver's gRPC server.
func Run(tydsdriver *ToyouDriver, tydsManager service.TydsManager, mounter *mount.SafeFormatAndMount, endpoint string) {
	// Register the CSI services
	tydsdriver.IdS = NewIdentityServer(tydsdriver)
	tydsdriver.CS = NewControllerServer(tydsdriver, tydsManager)
	tydsdriver.NS = NewNodeServer(tydsdriver, tydsManager, mounter)

	s := csicommon.NewNonBlockingGRPCServer()
	klog.Infof("Starting gRPC server on endpoint %s", endpoint)
	s.Start(endpoint, tydsdriver.IdS, tydsdriver.CS, tydsdriver.NS)
	s.Wait()
}
