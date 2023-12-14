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

package service

type VolumeManager interface {
	FindVolume(volId string) (interface{}, error)
	FindVolumeByName(volName string) map[string]interface{}
	CreateVolume(volName string, requestSize int) (volId string, err error)
	DeleteVolume(volId string) (err error)
	ListVolumes() []map[string]interface{}
	AttachVolume(volId string, instanceId string) (err error)
	DetachVolume(volId string, instanceId string) (err error)
	ResizeVolume(volId string, requestSize int) (err error)
}

type SnapshotManager interface {
	FindSnapshot(snapId string) (*interface{}, error)
	FindSnapshotByName(snapName string) (*interface{}, error)
	CreateSnapshot(snapName string, volID string) error
	DeleteSnapshot(snapID string) error
	CreateVolumeFromSnapshot(volName string, snapId string) (volId string, err error)
}

type TydsManager interface {
	SnapshotManager
	VolumeManager
}
