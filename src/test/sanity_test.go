package main

import (
	"testing"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

func TestMyDriver(t *testing.T) {
	// 设置驱动的配置
	config := sanity.NewTestConfig()
	config.Address = "unix:///var/lib/csi/your-driver.sock"

	// 运行测试用例
	sanity.Test(t, config)
}
