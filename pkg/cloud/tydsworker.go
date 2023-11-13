/*
Copyright (C) 2018 Yunify, Inc.

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

import (
	"toyou_csi/pkg/driver"

	qcservice "github.com/yunify/qingcloud-sdk-go/service"
	"k8s.io/klog"
)

var _ TydsManager = &TydsManager{}

// Example implementation of VolumeManager interface
type TydsManager struct {
	// Implement necessary fields and dependencies
}

func (m *VolumeManager) FindVolume(volId string) (*TydsClient.Volume, error) {
	// Implement finding a volume by ID
}

func (m *VolumeManager) FindVolumeByName(volName string) (*TydsClient.Volume, error) {
	// Implement finding a volume by name
}

func (m *VolumeManager) CreateVolume(volName string, requestSize int, replicas int, volType int, zone string, containerConfID string) (string, error) {
	// 0. Set CreateVolume args
	// create volume count
	count := 1
	// volume replicas
	replStr := DiskReplicaTypeName[replicas]
	// set input value
	input := &qcservice.CreateVolumesInput{
		Count:        &count,
		Repl:         &replStr,
		ReplicaCount: &replicas,
		Size:         &requestSize,
		VolumeName:   &volName,
		VolumeType:   &volType,
		Zone:         &zone,
	}
	if volType == int(driver.ThirdPartyStorageType) {
		input.ContainerConfID = &containerConfID
		klog.Infof("Call IaaS CreateVolume request name: %s, size: %d GB, type: %d, zone: %s, count: %d, replica: %s, replica_count: %d, container_conf_id: %s",
			*input.VolumeName, *input.Size, *input.VolumeType, *input.Zone, *input.Count, *input.Repl, *input.ReplicaCount, *input.ContainerConfID)
	} else {
		klog.Infof("Call IaaS CreateVolume request name: %s, size: %d GB, type: %d, zone: %s, count: %d, replica: %s, replica_count: %d",
			*input.VolumeName, *input.Size, *input.VolumeType, *input.Zone, *input.Count, *input.Repl, *input.ReplicaCount)
	}

	// 1. Create volume
	output, err := qm.volumeService.CreateVolumes(input)
	if err != nil {
		return "", err
	}
	// wait job
	klog.Infof("Call IaaS WaitJob %s", *output.JobID)
	if err := qm.waitJob(*output.JobID); err != nil {
		return "", err
	}
	// check output
	if *output.RetCode != 0 {
		klog.Errorf("Ret code: %d, message: %s", *output.RetCode, *output.Message)
		return "", errors.New(*output.Message)
	}
	newVolId = *output.Volumes[0]
	klog.Infof("Call IaaS CreateVolume name %s id %s succeed", volName, newVolId)
	return newVolId, nil
}
}

func (m *VolumeManager) DeleteVolume(volId string) error {
	// Implement deleting a volume
}

func (m *VolumeManager) AttachVolume(volId string, instanceId string) error {
	// Implement attaching a volume to an instance
}

func (m *VolumeManager) DetachVolume(volId string, instanceId string) error {
	// Implement detaching a volume from an instance
}

func (m *VolumeManager) ResizeVolume(volId string, requestSize int) error {
	// Implement resizing a volume
}

func (m *VolumeManager) CloneVolume(volName string, volType int, srcVolId string, zone string) (string, error) {
	// Implement cloning a volume
}

// Example implementation of SnapshotManager interface
type SnapshotManager struct {
	// Implement necessary fields and dependencies
}

func (m *SnapshotManager) FindSnapshot(snapId string) (*TydsClient.Snapshot, error) {
	// Implement finding a snapshot by ID
}

func (m *SnapshotManager) FindSnapshotByName(snapName string) (*TydsClient.Snapshot, error) {
	// Implement finding a snapshot by name
}

func (m *SnapshotManager) CreateSnapshot(snapName string, volId string) (string, error) {
	// Implement creating a snapshot
}

func (m *SnapshotManager) DeleteSnapshot(snapId string) error {
	// Implement deleting a snapshot
}

func (m *SnapshotManager) CreateVolumeFromSnapshot(volName string, snapId string, zone string) (string, error) {
	// Implement creating a volume from a snapshot
}

// Example implementation of TydsManager interface
type TydsManager struct {
	VolumeManager
	SnapshotManager
}
