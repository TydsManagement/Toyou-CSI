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

	"toyou-csi/pkg/common"
	"toyou-csi/pkg/driver"
	"toyou-csi/pkg/rpcserver"
	"toyou-csi/pkg/service"

	"k8s.io/klog"
)

const (
	version              = "1.0.1"
	defaultProvisionName = "csi.toyou.com"
	defaultConfigPath    = "/etc/config/config.yaml"
)

var (
	configPath = flag.String("config", defaultConfigPath, "config file path")
	endpoint   = flag.String("endpoint", "unix:///tcsi/csi.sock", "CSI endpoint")
	nodeID     = flag.String("nodeid", "default-node-id", "CSI node ID")
	maxVolume  = flag.Int64("maxvolume", 255, "Maximum volume value")
)

type Config struct {
	Version       string
	ProvisionName string
	ConfigPath    string
	Endpoint      string
	NodeID        string
	MaxVolume     int64
}

func main() {
	// klog.InitFlags(nil)
	flag.Parse()
	rand.NewSource(time.Now().UTC().UnixNano()) // 生成随机数种子

	// 创建 config 对象的指针
	config := &Config{
		Version:       version,
		ProvisionName: defaultProvisionName,
		ConfigPath:    *configPath,
		Endpoint:      *endpoint,
		NodeID:        *nodeID,
		MaxVolume:     *maxVolume,
	}

	mainProcess(config)
	os.Exit(0)
}

func mainProcess(config *Config) {
	tydsManager, err := service.NewManagerClientFromConfig(config.ConfigPath)
	if err != nil {
		klog.Fatal(err)
	}

	// 设置初始化磁盘驱动输入
	diskDriverInput := &driver.InitDiskDriverInput{
		Name:          config.ProvisionName,
		Version:       config.Version,
		NodeId:        config.NodeID,
		MaxVolume:     config.MaxVolume,
		VolumeCap:     driver.DefaultVolumeAccessModeType,
		ControllerCap: driver.DefaultControllerServiceCapability,
		NodeCap:       driver.DefaultNodeServiceCapability,
		PluginCap:     driver.DefaultPluginCapability,
	}

	mounter := common.NewSafeMounter()
	TydsDriver := driver.NewToyouDriver()
	TydsDriver.InitDiskDriver(diskDriverInput)
	rpcserver.Run(TydsDriver, tydsManager, mounter, config.Endpoint)
}
