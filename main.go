// +-------------------------------------------------------------------------
// | Copyright (C) 2023 Toyou, Inc.
// +-------------------------------------------------------------------------
// | Licensed under the Apache License, Version 2.0 (the "License");
// | you may not use this work except in compliance with the License.
// | You may obtain a copy of the License in the LICENSE file, or at:
// |
// | http://www.apache.org/licenses/LICENSE-2.0
// |
// | Unless required by applicable law or agreed to in writing, software
// | distributed under the License is distributed on an "AS IS" BASIS,
// | WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// | See the License for the specific language governing permissions and
// | limitations under the License.
// +-------------------------------------------------------------------------

package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"toyou_csi/pkg/disk/driver"
	"toyou_csi/pkg/disk/rpcserver"

	"k8s.io/klog"
)

const (
	version              = "1.0.0"
	defaultProvisionName = "disk.csi.toyou.com"
	defaultConfigPath    = "/etc/config/config.yaml"
)

var (
	driverName = flag.String("drivername", defaultProvisionName, "name of the driver")
	endpoint   = flag.String("endpoint", "unix:///tmp/csi.sock", "CSI endpoint")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rand.NewSource(time.Now().UTC().UnixNano())
	mainProcess()
	os.Exit(0)
}

func mainProcess() {

	// Set DiskDriverInput
	diskDriverInput := &driver.InitDiskDriverInput{
		Name:          *driverName,
		Version:       version,
		VolumeCap:     driver.DefaultVolumeAccessModeType,
		ControllerCap: driver.DefaultControllerServiceCapability,
		NodeCap:       driver.DefaultNodeServiceCapability,
		PluginCap:     driver.DefaultPluginCapability,
	}
	tydsDriver := driver.GetDiskDriver()
	tydsDriver.InitDiskDriver(diskDriverInput)
	rpcserver.Run(tydsDriver, *endpoint)
}
