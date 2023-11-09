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

type VolumeManager interface {
	FindVolume(volId string) (volInfo *TydsClient.Volume, err error)
	FindVolumeByName(volName string) (volInfo *TydsClient.Volume, err error)
	CreateVolume(volName string, requestSize int, replicas int, volType int, zone string, containerConfID string) (volId string, err error)
	DeleteVolume(volId string) (err error)
	AttachVolume(volId string, instanceId string) (err error)
	DetachVolume(volId string, instanceId string) (err error)
	ResizeVolume(volId string, requestSize int) (err error)
	CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error)
}

type SnapshotManager interface {
	FindSnapshot(snapId string) (snapInfo *TydsClient.Snapshot, err error)
	FindSnapshotByName(snapName string) (snapInfo *TydsClient.Snapshot, err error)
	CreateSnapshot(snapName string, volId string) (snapId string, err error)
	DeleteSnapshot(snapId string) (err error)
	CreateVolumeFromSnapshot(volName string, snapId string, zone string) (volId string, err error)
}

type TydsManager interface {
	SnapshotManager
	VolumeManager
}
