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

package cloud

import qcservice "github.com/yunify/qingcloud-sdk-go/service"

type VolumeManager interface {
	// FindVolume finds and gets volume information by volume ID.
	// Return:
	//   nil,  nil:  volume does not exist
	//   volume, nil: found volume and return volume info
	//   nil,  error: storage system internal error
	FindVolume(volId string) (volInfo *qcservice.Volume, err error)
	// FindVolumeByName finds and gets volume information by its name.
	// It will filter volume in deleted and ceased status and return first discovered item.
	// Return:
	//   nil, nil: volume does not exist
	//   volume, nil: found volume and return first discovered volume info
	//   nil, error: storage system internal error
	FindVolumeByName(volName string) (volInfo *qcservice.Volume, err error)
	// CreateVolume creates volume with specified name, size, replicas, type and zone and returns volume id.
	// Return:
	//   volume id, nil: succeed to create volume and return volume id
	//   nil, error: failed to create volume
	CreateVolume(volName string, requestSize int, replicas int, volType int, zone string, containerConfID string) (volId string, err error)
	// DeleteVolume deletes volume by id.
	// Return:
	//   nil: succeed to delete volume
	//   error: failed to delete volume
	DeleteVolume(volId string) (err error)
	// AttachVolume attaches volume on specified node.
	// Return:
	//   nil: succeed to attach volume
	//   error: failed to attach volume
	AttachVolume(volId string, instanceId string) (err error)
	// DetachVolume detaches volume from node.
	// Return:
	//   nil: succeed to detach volume
	//   error: failed to detach volume
	DetachVolume(volId string, instanceId string) (err error)
	// ResizeVolume expands volume to specified capacity.
	// Return:
	//   nil: succeed to expand volume
	//   error: failed to expand volume
	ResizeVolume(volId string, requestSize int) (err error)
	// CloneVolume clones a volume
	// Return:
	//   volume id, nil: succeed to clone volume and return volume id
	//   nil, error: failed to clone volume
	CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error)
}

type SnapshotManager interface {
	// FindSnapshot gets snapshot information by snapshot ID.
	// Return:
	//   nil, nil: snapshot does not exist
	//   snapshot info, nil: found snapshot and return snapshot info
	//   nil, error: storage system internal error
	FindSnapshot(snapId string) (snapInfo *qcservice.Snapshot, err error)
	// FindSnapshotByName finds and gets snapshot information by its name.
	// It will filter snapshot in deleted and ceased status and return first discovered item.
	// Return:
	//   nil, nil: snapshot does not exist
	//   volume, nil: found snapshot and return first discovered snapshot info
	//   nil, error: storage system internal error
	FindSnapshotByName(snapName string) (snapInfo *qcservice.Snapshot, err error)
	// CreateSnapshot creates a snapshot of specified volume.
	// Return:
	//   snapshot id, nil: succeed to create snapshot.
	//   nil, error: failed to create snapshot.
	CreateSnapshot(snapName string, volId string) (snapId string, err error)
	// DeleteSnapshot deletes a specified volume.
	// Return:
	//   nil: succeed to delete snapshot.
	//   error: failed to delete snapshot.
	DeleteSnapshot(snapId string) (err error)
	// CreateVolumeFromSnapshot creates volume from snapshot.
	// Return:
	//   volume id, nil: succeed to create volume
	//   nil, error: failed to create volume
	CreateVolumeFromSnapshot(volName string, snapId string, zone string) (volId string, err error)
}

type CloudManager interface {
	SnapshotManager
	VolumeManager
}
