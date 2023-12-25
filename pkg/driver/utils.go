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
	"toyou-csi/pkg/common"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// Check replica
// Support: 2 MultiReplicas, 1 SingleReplica
func IsValidReplica(replica int) bool {
	switch replica {
	case DiskThreeReplicaType:
		return true
	case DiskMultiReplicaType:
		return true
	case DiskSingleReplicaType:
		return true
	default:
		return false
	}
}

// Check file system type
// Support: ext3, ext4 and xfs
func IsValidFileSystemType(fs string) bool {
	switch fs {
	case common.FileSystemExt3:
		return true
	case common.FileSystemExt4:
		return true
	case common.FileSystemXfs:
		return true
	default:
		return false
	}
}

// NewVolumeCapabilityAccessMode creates CSI volume access mode object.
func NewVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

// NewControllerServiceCapability creates CSI controller capability object.
func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

// NewNodeServiceCapability creates CSI node capability object.
func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}
