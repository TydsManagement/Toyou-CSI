package rpcserver

import (
	"toyou_csi/pkg/cloud"
	"toyou_csi/pkg/common"
	"toyou_csi/pkg/disk/driver"

	"k8s.io/kubernetes/pkg/util/mount"
)

type NodeServer struct {
	driver  *driver.DiskDriver
	cloud   cloud.CloudManager
	mounter *mount.SafeFormatAndMount
	locks   *common.ResourceLocks
}

// NewNodeServer
// Create node server
func NewNodeServer(d *driver.DiskDriver, c cloud.CloudManager) *NodeServer {
	return &NodeServer{
		driver: d,
		cloud:  c,
		locks:  common.NewResourceLocks(),
	}
}
