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

package driver

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestCreateVolume(t *testing.T) {
	// 创建 ControllerServer 实例
	cs := &ControllerServer{
		TydsManager: &FakeTydsManager{}, // 使用虚拟的 TydsManager
	}

	// 创建 CreateVolume 请求
	req := &csi.CreateVolumeRequest{
		Name: "test-volume",
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: 1024 * 1024 * 1024, // 1 GB
		},
		Parameters: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}

	// 调用 CreateVolume 函数
	resp, err := cs.CreateVolume(context.Background(), req)

	// 断言函数进行断言
	assert.NoError(t, err, "CreateVolume should not return an error")
	assert.NotNil(t, resp, "CreateVolume response should not be nil")
	assert.Equal(t, "test-volume", resp.GetVolume().GetVolumeId(), "VolumeId should match the requested volume name")
	assert.Equal(t, int64(1024*1024*1024), resp.GetVolume().GetCapacityBytes(), "CapacityBytes should match the requested size")
	assert.Equal(t, req.GetParameters(), resp.GetVolume().GetVolumeContext(), "VolumeContext should match the requested parameters")
}

func TestDeleteVolume(t *testing.T) {
	// Create a new instance of the ControllerServer
	cs := &ControllerServer{
		TydsManager: &FakeTydsManager{}, // Replace with your mock implementation
	}

	// Create a context for the test
	ctx := context.Background()

	// Create a valid DeleteVolumeRequest with a volume ID
	req := &csi.DeleteVolumeRequest{
		VolumeId: "volume-123",
	}

	// Call the DeleteVolume function
	resp, err := cs.DeleteVolume(ctx, req)

	// Check for errors
	if err != nil {
		t.Errorf("DeleteVolume returned an error: %v", err)
	}
	if resp == nil {
		t.Error("DeleteVolume returned a nil response")
	}

	// Add additional assertions based on your requirements
	// For example, you can check if the DeleteVolume function called the DeleteVolume method of the TydsManager mock

	// Cleanup or verify any other side effects of the DeleteVolume function if needed
}


func TestListVolumes(t *testing.T) {
	// 创建一个 ControllerServer 实例，或使用已有的实例
	cs := &ControllerServer{
		TydsManager: &FakeTydsManager{},
	}

	// 创建一个 ListVolumesRequest 实例，根据需要设置请求参数
	req := &csi.ListVolumesRequest{}

	// 调用 ListVolumes 函数
	resp, err := cs.ListVolumes(context.Background(), req)
	assert.NoError(t, err) // 确保没有返回错误
	volumeListInterface := cs.TydsManager.ListVolumes()
	// 根据你的 TydsManager 实现，验证 ListVolumesResponse 的正确性
	expectedEntries := make([]*csi.ListVolumesResponse_Entry, 0, len(volumeListInterface))
		assert.Equal(t, expectedEntries, resp.Entries) // 确保返回的 Entries 列表与期望的一致
}


func TestControllerPublishVolume(t *testing.T) {
	// 创建一个 ControllerServer 实例，或使用已有的实例
	cs := &ControllerServer{
		TydsManager: &FakeTydsManager{},
	}

	// 创建一个 ControllerPublishVolumeRequest 实例，根据需要设置请求参数
	req := &csi.ControllerPublishVolumeRequest{
		VolumeId: "volume-123",
		NodeId:   "node-456",
	}

	// 调用 ControllerPublishVolume 函数
	resp, err := cs.ControllerPublishVolume(context.Background(), req)
	assert.NoError(t, err) // 确保没有返回错误

	// 添加断言以验证返回结果
	assert.NotNil(t, resp) // 确保返回的响应不为 nil
	// 根据你的测试需求，添加更多的断言来验证 resp 中包含正确的信息

	// 在这里添加额外的断言，根据你的测试需求验证函数的行为和返回结果
}

func TestControllerUnpublishVolume(t *testing.T) {
	// 创建一个 ControllerServer 实例，或使用已有的实例
	cs := &ControllerServer{
		TydsManager: &FakeTydsManager{},
	}

	// 创建一个 ControllerUnpublishVolumeRequest 实例，根据需要设置请求参数
	req := &csi.ControllerUnpublishVolumeRequest{
		VolumeId: "volume-123",
		NodeId:   "node-456",
	}

	// 调用 ControllerUnpublishVolume 函数
	_, err := cs.ControllerUnpublishVolume(context.Background(), req)
	assert.NoError(t, err) // 确保没有返回错误

	// 在这里添加额外的断言，根据你的测试需求验证函数的行为和返回结果
}

func TestValidateVolumeCapabilities(t *testing.T) {
	// 创建一个 ControllerServer 实例，或使用已有的实例
	cs := &ControllerServer{}

	// 创建一个 ValidateVolumeCapabilitiesRequest 实例，根据需要设置请求参数
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: // 设置你要验证的 VolumeCapabilities,
	}

	// 调用 ValidateVolumeCapabilities 函数
	_, err := cs.ValidateVolumeCapabilities(context.Background(), req)
	assert.NoError(t, err) // 确保没有返回错误

	// 在这里添加额外的断言，根据你的测试需求验证函数的行为和返回结果
}

func TestControllerGetCapabilities(t *testing.T) {
	// 创建一个 ControllerServer 实例，或使用已有的实例
	cs := &ControllerServer{
		Driver: // 设置你的 Driver 实例,
	}

	// 创建一个 ControllerGetCapabilitiesRequest 实例，根据需要设置请求参数
	req := &csi.ControllerGetCapabilitiesRequest{}

	// 调用 ControllerGetCapabilities 函数
	_, err := cs.ControllerGetCapabilities(context.Background(), req)
	assert.NoError(t, err) // 确保没有返回错误

	// 在这里添加额外的断言，根据你的测试需求验证函数的行为和返回结果
}