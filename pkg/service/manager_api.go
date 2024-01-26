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
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"k8s.io/klog"
)

var _ TydsManager = &Manager{}

type Manager struct {
	tydsClient *TydsClient
	poolName   string
	stripSize  string
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

func (m *Manager) CreateVolumeFromSnapshot(volName string, snapID string) (string, error) {
	snapshotName := fmt.Sprintf(snapID)
	volumeName := fmt.Sprintf(volName)
	poolName := m.poolName

	srcVol := m.FindVolumeByName(volumeName)
	if srcVol == nil {
		msg := fmt.Sprintf("Volume \"%s\" not found in create_volume_from_snapshot.", volumeName)
		return "", errors.New(msg)
	}
	srcVolName, ok := srcVol["name"].(string)
	if !ok {
		msg := fmt.Sprintf("Invalid type for srcVolName: %T", srcVol["name"])
		return "", errors.New(msg)
	}
	srcVolPoolName, ok := srcVol["PoolName"].(string)
	if !ok {
		msg := fmt.Sprintf("Invalid type for srcVolPoolName: %T", srcVol["PoolName"])
		return "", errors.New(msg)
	}

	err := m.tydsClient.CreateVolumeFromSnapshot(volumeName, poolName, snapshotName, srcVolName, srcVolPoolName)
	if err != nil {
		return "", err
	}

	return volumeName, nil
}

func (m *Manager) FindVolume(volID string) (interface{}, error) {
	res, err := m.tydsClient.GetVolume(volID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (m *Manager) FindVolumeByName(volName string) map[string]interface{} {
	volumeList := m.tydsClient.GetVolumes()
	if volumeList == nil {
		return nil
	}

	for _, vol := range volumeList {
		name := vol["name"].(string)
		if name == volName {
			return vol
		}
	}

	// 返回一个空的interface{}，表示未找到对应名称的卷
	return nil
}

// CreateVolume creates a volume with the specified parameters
func (m *Manager) CreateVolume(volName string, requestSize int) (volId string, err error) {
	// poolName string, stripSize int
	poolName := m.poolName
	stripSize := m.stripSize
	volName, err = m.tydsClient.CreateVolume(volName, requestSize, poolName, stripSize)
	if err != nil {
		return "", err
	}
	return volName, nil
}

// DeleteVolume deletes a volume with the given ID
func (m *Manager) DeleteVolume(volId int) (err error) {
	err = m.tydsClient.DeleteVolume(volId)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) ListVolumes() []map[string]interface{} {
	volumeList := m.tydsClient.GetVolumes()
	return volumeList
}

// 根据 iqn 生成初始化连接的组名
func generateInitiatorGroupName(iqn string) (string, error) {
	// 计算 MD5 哈希值
	hasher := md5.New()
	hasher.Write([]byte(iqn))
	hashBytes := hasher.Sum(nil)

	// 将哈希值转换为 UUID
	namestring, err := uuid.FromBytes(hashBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate initiator group name: %v", err)
	}

	// 拼接组名
	groupName := "initiator-group-" + namestring.String()
	return groupName, nil
}

// AttachVolume attaches a volume to a specified instance
func (m *Manager) AttachVolume(volID, iqn string) error {
	// Get volume information
	volume, err := m.tydsClient.GetVolume(volID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %v", err)
	}

	volumeData, ok := volume.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected volume data type")
	}

	volumeName, _ := volumeData["blockName"].(string)
	poolName, _ := volumeData["poolName"].(string)

	groupName, err := generateInitiatorGroupName(iqn)
	if err != nil {
		return fmt.Errorf("failed to generate initiator group name: %v", err)
	}

	volInfo := map[string]interface{}{
		"name": volumeName,
		"size": volumeData["sizeMB"],
		"pool": poolName,
	}

	// Check if client group exists and create it if needed
	initiatorList, err := m.tydsClient.GetInitiatorList()
	if err != nil {
		return fmt.Errorf("failed to get initiator list: %v", err)
	}

	initiatorListMap, ok := initiatorList.(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to convert initiatorList to map[string]interface{}")
	}

	clientGroupList, ok := initiatorListMap["client_group_list"].([]interface{})
	if !ok {
		return fmt.Errorf("failed to convert client_group_list to []interface{}")
	}

	initiatorExistence := false
	DidCreateInitiator := false
	for _, group := range clientGroupList {
		groupMap, ok := group.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to convert group to map[string]interface{}")
		}

		if groupMap["group_name"].(string) == groupName {
			initiatorExistence = true
			break
		}
	}

	if !initiatorExistence {
		client := []map[string]string{
			{
				"ip":  "",
				"iqn": iqn,
			},
		}
		err = m.tydsClient.CreateInitiatorGroup(groupName, client)
		DidCreateInitiator = true
		if err != nil {
			return fmt.Errorf("failed to create initiator group: %v", err)
		}
	}

	targetNodeList, err := m.tydsClient.GetHost()
	if err != nil {
		return fmt.Errorf("failed to get target nodes: %v", err)
	}

	targetNameList := make([]string, len(targetNodeList))
	for i, target := range targetNodeList {
		targetData, ok := target.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected target data type")
		}

		targetName, ok := targetData["name"].(string)
		if !ok {
			return fmt.Errorf("unexpected target name type")
		}

		targetNameList[i] = targetName
	}

	itList, err := m.tydsClient.GetInitiatorTargetConnections()
	if err != nil {
		return fmt.Errorf("failed to get initiator-target connections: %v", err)
	}

	var itInfo map[string]interface{}
	for _, it := range itList {
		itData, ok := it.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected it data type")
		}

		targetName, ok := itData["target_name"].(string)
		if !ok {
			return fmt.Errorf("unexpected target_name type")
		}

		if strings.Contains(targetName, groupName) {
			itInfo = itData
			break
		}
	}
	DidCreateTlConnection := false
	if itInfo != nil {
		// Update connection between target and client group
		targetIQN := itInfo["target_iqn"].(string)

		var lunInfo map[string]interface{}
		blockList, ok := itInfo["block"].([]interface{})
		if ok {
			for _, block := range blockList {
				blockData, ok := block.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unexpected block data type")
				}

				if blockData["name"].(string) == volumeName {
					lunInfo = blockData
					break
				}
			}
		}

		if lunInfo == nil {
			// Add new volume to existing Initiator-Target connection
			targetNameList := itInfo["hostName"].([]string)
			volsInfo := itInfo["block"].([]interface{})
			volsInfo = append(volsInfo, volInfo)
			_, err = m.tydsClient.ModifyTarget(targetIQN, targetNameList, volsInfo)
			DidCreateTlConnection = true
			if err != nil {
				return fmt.Errorf("failed to modify target: %v", err)
			}
		}
	} else {
		// Create connection between target and client group
		_, err = m.tydsClient.CreateTarget(groupName, targetNameList, []interface{}{volInfo})
		DidCreateTlConnection = true
		if err != nil {
			return fmt.Errorf("failed to create target connection: %v", err)
		}
	}

	// Assure that all target nodes are active
	itList, err = m.tydsClient.GetInitiatorTargetConnections()
	if err != nil {
		return fmt.Errorf("failed to get initiator-target connections: %v", err)
	}

	for _, it := range itList {
		itData, ok := it.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected it data type")
		}

		targetName, ok := itData["target_name"].(string)
		if !ok {
			return fmt.Errorf("unexpected target_name type")
		}

		if strings.Contains(targetName, groupName) {
			itInfo = itData
			break
		}
	}

	if DidCreateInitiator && itInfo != nil && DidCreateTlConnection {
		// Query service status
		services, err := m.tydsClient.GetService()
		if err != nil {
			return fmt.Errorf("failed to get service status: %v", err)
		}

		serviceStatus := make(map[string]string)
		for _, service := range services {
			hostName, ok := service["hostName"].(string)
			if !ok {
				return fmt.Errorf("unexpected hostName type")
			}

			state, ok := service["state"].(string)
			if !ok {
				return fmt.Errorf("unexpected service state type")
			}

			serviceStatus[hostName] = state
		}

		// Check and start inactive target nodes
		for _, targetName := range targetNameList {
			state, ok := serviceStatus[targetName]
			if !ok || state == "down" {
				err = m.tydsClient.StartService(targetName)
				if err != nil {
					return fmt.Errorf("failed to start service for %s: %v", targetName, err)
				}
			}
		}

		targetIQN := itInfo["target_iqn"].(string)

		// Generate configuration
		err = m.tydsClient.GenerateConfig(targetIQN) // Replace with actual Target IQN
		if err != nil {
			return fmt.Errorf("failed to generate config: %v", err)
		}
	}

	return nil
}

// DetachVolume detaches a volume from a specified instance
func (m *Manager) DetachVolume(volId string, iqn string) error {
	// 获取卷信息
	volume, err := m.tydsClient.GetVolume(volId)
	if err != nil {
		return fmt.Errorf("failed to get volume: %v", err)
	}

	volumeName, _ := volume.(map[string]interface{})["name"].(string)

	// 生成 initiator 组名
	groupName, err := generateInitiatorGroupName(iqn)

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
func (m *Manager) ResizeVolume(volId string, newSize int64) error {
	// 获取卷信息
	volume, err := m.tydsClient.GetVolume(volId)
	if err != nil {
		return fmt.Errorf("failed to get volume: %v", err)
	}

	volumeName, ok := volume.(map[string]interface{})["blockName"].(string)
	if !ok {
		return fmt.Errorf("failed to get volume name")
	}

	poolName, ok := volume.(map[string]interface{})["poolName"].(string)
	if !ok {
		return fmt.Errorf("failed to get pool name")
	}

	// 扩展卷大小
	if err := m.tydsClient.ExtendVolume(volumeName, poolName, newSize); err != nil {
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

func NewManagerClientFromSecret(secretPath string) (*Manager, error) {
	// Read secret files from the specified directory
	f, err := os.Open(secretPath)
	if err != nil {
		log.Fatal(err)
	}
	secretFiles, err := f.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret directory: %v", err)
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}

	klog.Infof("Get Secret Directory Files: %v", secretFiles)

	// Initialize variables to store configuration values
	var username, password, hostIP, port, poolName, stripSize string

	// Loop through the secret files
	for _, file := range secretFiles {
		// klog.Infof("Get Secret Files Fields: %v", file.Name())
		// Skip directories and files that are not needed
		if file.IsDir() || !isRequiredField(file.Name()) {
			continue
		}

		fileData, err := os.ReadFile(filepath.Join(secretPath, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file %s: %v", file.Name(), err)
		}

		// Extract the value from the file data
		value := string(fileData)

		// Set the corresponding variable based on the file name
		switch file.Name() {
		case "username":
			username = value
		case "password":
			password = value
		case "hostIP":
			hostIP = value
		case "port":
			port = value
		case "poolName":
			poolName = value
		case "stripSize":
			stripSize = value
		}
	}

	// Convert port and stripSize to their respective types
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port to integer: %v", err)
	}

	// Create a new instance of TydsClient
	managerClient := NewTydsClient(hostIP, portInt, username, password)
	m := &Manager{
		tydsClient: managerClient,
		poolName:   poolName,
		stripSize:  stripSize,
	}
	return m, nil
}

// Function to check if a field is required or not
func isRequiredField(fieldName string) bool {
	requiredFields := map[string]bool{
		"username":  true,
		"password":  true,
		"hostIP":    true,
		"port":      true,
		"poolName":  true,
		"stripSize": true,
	}
	return requiredFields[fieldName]
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
