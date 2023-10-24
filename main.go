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

	"toyou_csi/pkg/cloud"
	"toyou_csi/pkg/common"
	"toyou_csi/pkg/disk/driver"
	"toyou_csi/pkg/disk/rpcserver"

	"k8s.io/klog"
)

const (
	version              = "1.0.0"
	defaultProvisionName = "disk.csi.qingcloud.com"
	defaultConfigPath    = "/etc/config/config.yaml"
)

var (
	configPath          = flag.String("config", defaultConfigPath, "server config file path")
	driverName          = flag.String("drivername", defaultProvisionName, "name of the driver")
	endpoint            = flag.String("endpoint", "unix://csi/csi.sock", "CSI endpoint")
	maxVolume           = flag.Int64("maxvolume", 10, "Maximum number of volumes that controller can publish to the node.")
	nodeId              = flag.String("nodeid", "", "If driver cannot get instance ID from /etc/qingcloud/instance-id, we would use this flag.")
	retryDetachTimesMax = flag.Int("retry-detach-times-max", 100, "Maximum retry times of failed detach volume. Set to 0 to disable the limit.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rand.NewSource(time.Now().UTC().UnixNano())
	mainProcess()
	os.Exit(0)
}

func mainProcess() {
	// Get Instance Id
	instanceId, err := driver.GetInstanceIdFromFile(driver.DefaultInstanceIdFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warningf("Failed to get instance id from file, use --nodeId flag. error: %s", err)
			instanceId = *nodeId
		} else {
			klog.Fatalf("Failed to get instance id from file, error: %s", err)
		}
	}
	// Get qingcloud config object
	cloud, err := cloud.NewQingCloudManagerFromFile(*configPath)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof("Version: %s", version)
	// Set DiskDriverInput
	diskDriverInput := &driver.InitDiskDriverInput{
		Name:          *driverName,
		Version:       version,
		NodeId:        instanceId,
		MaxVolume:     *maxVolume,
		VolumeCap:     driver.DefaultVolumeAccessModeType,
		ControllerCap: driver.DefaultControllerServiceCapability,
		NodeCap:       driver.DefaultNodeServiceCapability,
		PluginCap:     driver.DefaultPluginCapability,
	}
	// For resize
	mounter := common.NewSafeMounter()
	tydsDriver := driver.GetDiskDriver()
	tydsDriver.InitDiskDriver(diskDriverInput)
	rpcserver.Run(tydsDriver, cloud, mounter, *endpoint, *retryDetachTimesMax)
}
