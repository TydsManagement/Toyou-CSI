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

package service

import (
	"errors"
	"fmt"
)

var _ TydsManager = &Manager{}

type Manager struct {
	tydsClient *TydsClient
}

func (m *Manager) FindSnapshot(snapID string) (*interface{}, error) {
	url := fmt.Sprintf("snapshot/%s", snapID)
	res, err := m.tydsClient.SendHTTPAPI(url, nil, "GET")
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (m *Manager) FindSnapshotByName(snapName string) (*interface{}, error) {
	url := "snapshot/"
	params := map[string]interface{}{
		"name": snapName,
	}
	res, err := m.tydsClient.SendHTTPAPI(url, params, "GET")
	if err != nil {
		return nil, err
	}
	return &res, nil
}
func (m *Manager) CreateSnapshot(snapName string, volID string) error {
	comment := fmt.Sprintf("%s/%s", volID, snapName)
	err := m.tydsClient.CreateSnapshot(snapName, volID, comment)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) DeleteSnapshot(snapID string) error {
	err := m.tydsClient.DeleteSnapshot(snapID)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) CreateVolumeFromSnapshot(volName string, snapID string, zone string) (string, error) {
	snapshotName := convertStr(snapID)
	volumeName := convertStr(volName)
	poolName := extractPoolName(zone)

	srcVol, _ := m.FindVolumeByName(volumeName)
	if srcVol == nil {
		msg := fmt.Sprintf("Volume \"%s\" not found in create_volume_from_snapshot.", volumeName)
		return "", errors.New(msg)
	}

	err := m.tydsClient.CreateVolumeFromSnapshot(volumeName, poolName, snapshotName, srcVol.VolumeName, srcVol.PoolName)
	if err != nil {
		return "", err
	}

	return volumeName, nil
}

func (m *Manager) FindVolume(volID string) (*interface{}, error) {
	res, err := m.tydsClient.GetVolume(volID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (m *Manager) FindVolumeByName(volName string) (interface{}, error) {
	volumeList := m.tydsClient.GetVolumes()
	if volumeList == nil {
		return nil, err
	}

	for _, vol := range volumeList {
		if vol.BlockName == volName {
			return vol, nil
		}
	}

	// 返回一个空的interface{}，表示未找到对应名称的卷
	return nil, nil
}
func (m Manager) CreateVolume(volName string, requestSize int, replicas int, volType int, zone string, containerConfID string) (volId string, err error) {
	// TODO implement me
	panic("implement me")
}

func (m Manager) DeleteVolume(volId string) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m Manager) AttachVolume(volId string, instanceId string) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m Manager) DetachVolume(volId string, instanceId string) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m Manager) ResizeVolume(volId string, requestSize int) (err error) {
	// TODO implement me
	panic("implement me")
}

func (m Manager) CloneVolume(volName string, volType int, srcVolId string, zone string) (volId string, err error) {
	// TODO implement me
	panic("implement me")
}

func NewManagerClientFromConfig(configPath string) (*Manager, error) {
	// Read configuration from file
	config, err := ReadconfigFromFile(configPath)
	if err != nil {
		return nil, err
	}

	// Create a new instance of TydsClient
	managerClient := NewTydsClient(config.Hostname, config.Port, config.Username, config.Password)
	m := Manager{
		tydsClient: managerClient,
	}
	return &m, nil
}
