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

	"toyou_csi/pkg/common"
	"toyou_csi/pkg/driver"
	"toyou_csi/pkg/rpcserver"
	"toyou_csi/pkg/service"

	"k8s.io/klog"
)

// 定义一个包含所有配置项的结构体
type Config struct {
	Version       string
	ProvisionName string
	ConfigPath    string
	Endpoint      string
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rand.NewSource(time.Now().UTC().UnixNano()) // 生成随机数种子

	config := &Config{
		Version:       "1.0.0",
		ProvisionName: "disk.csi.toyou.com",
		ConfigPath:    getFlagValue("config", "/etc/config/config.yaml"),
		Endpoint:      getFlagValue("endpoint", "unix:///tmp/csi.sock"),
	}

	mainProcess(config)
	os.Exit(0)
}

// 从flag包中获取命令行参数值的辅助函数
func getFlagValue(name string, defaultValue string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
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
