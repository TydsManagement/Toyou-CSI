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
	"strings"
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
	poolName, _ := extractPoolName(zone)

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

// CreateVolume creates a volume with the specified parameters
func (m *Manager) CreateVolume(volName string, requestSize int, poolName string, stripSize int) (volId string, err error) {

	volId, err = m.tydsClient.CreateVolume(volName, requestSize, poolName, stripSize)
	if err != nil {
		return "", err
	}
	return volId, nil
}

// DeleteVolume deletes a volume with the given ID
func (m *Manager) DeleteVolume(volId string) (err error) {
	err = m.tydsClient.DeleteVolume(volId)
	if err != nil {
		return err
	}
	return nil
}

// AttachVolume attaches a volume to a specified instance
func (m *Manager) AttachVolume(volId string, instanceId string) error {
	// 获取卷信息
	volume, err := m.tydsClient.GetVolume(volId)
	if err != nil {
		return fmt.Errorf("failed to get volume: %v", err)
	}

	volumeName, _ := volume.(map[string]interface{})["name"].(string)
	poolName, _ := volume.(map[string]interface{})["poolName"].(string)

	// 模拟Cinder中的初始化连接逻辑
	groupName := "initiator-group-" + instanceId

	volInfo := map[string]interface{}{
		"name": volumeName,
		"size": volume.(map[string]interface{})["sizeMB"],
		"pool": poolName,
	}

	// 检查启动器组是否存在，并创建（如果需要）
	initiatorList, err := m.tydsClient.GetInitiatorList()
	if err != nil {
		return fmt.Errorf("failed to get initiator list: %v", err)
	}

	initiatorExistence := false
	for _, initiator := range initiatorList.([]interface{}) {
		if initiator.(map[string]interface{})["group_name"].(string) == groupName {
			initiatorExistence = true
			break
		}
	}

	if !initiatorExistence {
		client := []map[string]string{
			{
				"ip":  "IP_ADDRESS", // 替换为实际的 IP 地址
				"iqn": instanceId,
			},
		}
		err = m.tydsClient.CreateInitiatorGroup(groupName, client)
		if err != nil {
			return fmt.Errorf("failed to create initiator group: %v", err)
		}
	}

	// 创建或更新目标和启动器组之间的连接
	targetNodeList, err := m.tydsClient.GetTarget()
	if err != nil {
		return fmt.Errorf("failed to get target nodes: %v", err)
	}

	targetNameList := make([]string, len(targetNodeList.([]interface{})))
	for i, target := range targetNodeList.([]interface{}) {
		targetNameList[i] = target.(map[string]interface{})["name"].(string)
	}

	_, err = m.tydsClient.CreateTarget(groupName, targetNameList, []interface{}{volInfo})
	if err != nil {
		return fmt.Errorf("failed to create or update target: %v", err)
	}

	// 生成配置
	err = m.tydsClient.GenerateConfig("TARGET_IQN") // 替换为实际的 Target IQN
	if err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	return nil
}

// DetachVolume detaches a volume from a specified instance
func (m *Manager) DetachVolume(volId string, instanceId string) error {
	// 获取卷信息
	volume, err := m.tydsClient.GetVolume(volId)
	if err != nil {
		return fmt.Errorf("failed to get volume: %v", err)
	}

	volumeName, _ := volume.(map[string]interface{})["name"].(string)

	// 生成 initiator 组名
	groupName := "initiator-group-" + instanceId

	// 获取启动器-目标连接信息
	itList, err := m.tydsClient.GetInitiatorTargetConnections()
	if err != nil {
		return fmt.Errorf("failed to get initiator-target connections: %v", err)
	}

	var itInfo map[string]interface{}
	for _, it := range itList {
		itMap := it.(map[string]interface{})
		if itMap["target_name"].(string) == groupName {
			itInfo = itMap
			break
		}
	}

	if itInfo != nil {
		targetIqn := itInfo["target_iqn"].(string)
		targetNameList := itInfo["hostName"].([]string)
		volsInfo := itInfo["block"].([]interface{})

		// 移除卷信息
		newVolsInfo := make([]interface{}, 0)
		for _, vol := range volsInfo {
			if vol.(map[string]interface{})["name"].(string) != volumeName {
				newVolsInfo = append(newVolsInfo, vol)
			}
		}

		// 更新或删除目标
		if len(newVolsInfo) == 0 {
			// 如果没有卷信息，删除目标
			_, err := m.tydsClient.DeleteTarget(targetIqn)
			if err != nil {
				return fmt.Errorf("failed to delete target: %v", err)
			}
		} else {
			// 否则更新目标
			_, err := m.tydsClient.ModifyTarget(targetIqn, targetNameList, newVolsInfo)
			if err != nil {
				return fmt.Errorf("failed to modify target: %v", err)
			}
		}
	}

	return nil
}

// ResizeVolume resizes a volume to the specified size.
func (m *Manager) ResizeVolume(volId string, newSize int) error {
	// 获取卷信息
	volume, err := m.tydsClient.GetVolume(volId)
	if err != nil {
		return fmt.Errorf("failed to get volume: %v", err)
	}

	volumeName, ok := volume.(map[string]interface{})["name"].(string)
	if !ok {
		return fmt.Errorf("failed to get volume name")
	}

	poolName, ok := volume.(map[string]interface{})["poolName"].(string)
	if !ok {
		return fmt.Errorf("failed to get pool name")
	}

	sizeMB := newSize * 1024 // 将 GB 转换为 MB

	// 扩展卷大小
	if err := m.tydsClient.ExtendVolume(volumeName, poolName, sizeMB); err != nil {
		return fmt.Errorf("failed to extend volume: %v", err)
	}

	// 获取启动器-目标连接信息并重启服务
	itList, err := m.tydsClient.GetInitiatorTargetConnections()
	if err != nil {
		return fmt.Errorf("failed to get initiator-target connections: %v", err)
	}

	for _, it := range itList {
		itMap, ok := it.(map[string]interface{})
		if !ok {
			continue // Skip invalid entries
		}

		lunInfo := findLunInfo(itMap["block"].([]interface{}), volumeName)
		if lunInfo != nil {
			hostName, ok := itMap["hostName"].(string)
			if !ok {
				continue // Skip invalid entries
			}

			if err := m.tydsClient.RestartService(hostName); err != nil {
				return fmt.Errorf("failed to restart service on host %s: %v", hostName, err)
			}
		}
	}

	return nil
}

// findLunInfo 查找特定卷的 LUN 信息
func findLunInfo(luns []interface{}, volumeName string) map[string]interface{} {
	for _, lun := range luns {
		lunMap, ok := lun.(map[string]interface{})
		if ok && lunMap["name"].(string) == volumeName {
			return lunMap
		}
	}
	return nil
}

// extractPoolName 从 host 字符串中提取存储池名称。
func extractPoolName(host string) (string, error) {
	defaultPoolName := "_pool0"

	if host == "" {
		return "", fmt.Errorf("volume is not assigned to a host")
	}

	parts := strings.Split(host, "#")
	if len(parts) == 2 {
		// 如果 host 字符串中有两部分，则第二部分是存储池名称
		return parts[1], nil
	}

	// 如果 host 字符串中没有 '#'，则使用默认的存储池名称
	return defaultPoolName, nil
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

// findPoolID 根据存储池名称查找存储池 ID。
func (m *Manager) findPoolID(poolName string) (string, error) {
	poolList, err := m.tydsClient.GetPools()
	if err != nil {
		return "", fmt.Errorf("failed to get pools: %v", err)
	}
	for _, pool := range poolList {
		if poolName == pool["name"] {
			// 假设存储池对象有一个 "id" 字段
			poolID, ok := pool["id"].(string)
			if !ok {
				return "", fmt.Errorf("pool id is not a string for pool: %s", poolName)
			}
			return poolID, nil
		}
	}

	return "", fmt.Errorf("pool not found: %s", poolName)
}
