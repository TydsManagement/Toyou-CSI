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
	"context"
	"math"

	"toyou-csi/pkg/common"
	"toyou-csi/pkg/service"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/timestamp"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// ControllerServer implements the CSI controller service.
type ControllerServer struct {
	*csicommon.DefaultControllerServer
	Driver      *ToyouDriver
	TydsManager service.TydsManager
	Locks       *common.ResourceLocks
}

func (cs *ControllerServer) ControllerModifyVolume(ctx context.Context, request *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NewControllerServer creates a new ControllerServer.
func NewControllerServer(d *ToyouDriver, tm service.TydsManager) *ControllerServer {
	return &ControllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.Driver),
		Driver:                  d,
		TydsManager:             tm,
		Locks:                   common.NewResourceLocks(),
	}
}

// CreateVolume creates a new volume.
func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	funcName := "CreateVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volName := req.GetName()
	requestSize := req.GetCapacityRange().GetRequiredBytes()

	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volName)
	if acquired := cs.Locks.TryAcquire(volName); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volName)
	}
	defer cs.Locks.Release(volName)

	// Convert bytes to megabytes
	requestSizeMB := int64(math.Ceil(float64(requestSize) / 1024 / 1024))

	volId, err := cs.TydsManager.CreateVolume(volName, int(requestSizeMB))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create volume: %v", err)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volId,
			CapacityBytes: requestSize,
			VolumeContext: req.GetParameters(),
			ContentSource: req.GetVolumeContentSource(),
		},
	}, nil
}

// DeleteVolume deletes a volume.
func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	funcName := "DeleteVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volName := req.GetVolumeId()
	if volName == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}
	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volName)
	if acquired := cs.Locks.TryAcquire(volName); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volName)
	}
	defer cs.Locks.Release(volName)

	volId, err := cs.GetVolumeIDByName(volName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find volume: %v", nil)
	}
	volIDFloat := volId["id"].(float64)
	// volIDStr := strconv.FormatFloat(volIDFloat, 'f', -1, 64)
	// parts := strings.Split(volIDStr, ".")
	// volIDIntStr := parts[0]
	err = cs.TydsManager.DeleteVolume(int(volIDFloat))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete volume: %v", err)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ListVolumes lists all volumes managed by the storage system.
func (cs *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	funcName := "ListVolumes"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volumeListInterface := cs.TydsManager.ListVolumes()

	entries := make([]*csi.ListVolumesResponse_Entry, 0, len(volumeListInterface))
	for _, volMap := range volumeListInterface {
		volumeId, ok := volMap["name"].(string)
		if !ok {
			klog.Errorf("VolumeId is missing or not a string")
			continue
		}
		capacityBytes, ok := volMap["size"].(int64) // 或者是其他你期望的类型
		if !ok {
			klog.Errorf("CapacityBytes is missing or not the correct type")
			continue
		}

		entry := &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				VolumeId:      volumeId,
				CapacityBytes: capacityBytes,
				// 根据需要填充其他字段
			},
			// 可以添加状态或其他元数据
		}
		entries = append(entries, entry)
	}

	return &csi.ListVolumesResponse{
		Entries: entries,
	}, nil
}

// ControllerPublishVolume attaches a volume to a node.
func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	funcName := "ControllerPublishVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volId := req.GetVolumeId()
	nodeId := req.GetNodeId()
	// 获取节点上 iSCSI 发起端的 IP 地址和 IQN
	iqn := req.GetVolumeContext()["iqn"]

	// Log the extracted values using klog
	klog.Infof("IQN: %s", iqn)

	if volId == "" || nodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID and Node ID are required")
	}

	err := cs.TydsManager.AttachVolume(volId, iqn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to attach volume: %v", err)
	}

	return &csi.ControllerPublishVolumeResponse{}, nil
}

// ControllerUnpublishVolume detaches a volume from a node.
func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	funcName := "ControllerUnpublishVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volId := req.GetVolumeId()
	nodeId := req.GetNodeId()
	// 获取节点上 iSCSI 发起端的 IP 地址和 IQN
	if volId == "" || nodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID and Node ID are required")
	}

	err := cs.TydsManager.DetachVolume(volId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to detach volume: %v", err)
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks if a volume has the given capabilities.
func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	funcName := "ValidateVolumeCapabilities"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	// This implementation is a placeholder. Replace it with your actual validation logic.

	volumeCapabilities := req.GetVolumeCapabilities()
	klog.Infof("Validating volume capabilities: %v", volumeCapabilities)

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: volumeCapabilities,
		},
	}, nil
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	funcName := "ControllerGetCapabilities"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	capabilities := cs.Driver.GetControllerCapability()
	klog.Infof("Controller capabilities: %v", capabilities)

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil
}

// CreateSnapshot creates a new snapshot from a volume.
func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	funcName := "CreateSnapshot"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	snapName := req.GetName()
	volID := req.GetSourceVolumeId()
	if err := cs.TydsManager.CreateSnapshot(snapName, volID); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create snapshot: %v", err)
	}

	// Implement logic to retrieve snapshot details and return them.
	// Assuming the snapshot is created instantly for this example.
	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     "snap-12345", // Replace with actual snapshot ID
			SourceVolumeId: volID,
			CreationTime:   &timestamp.Timestamp{}, // Set the correct creation time
			ReadyToUse:     true,
		},
	}, nil
}

// DeleteSnapshot deletes a specific snapshot.
func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	funcName := "DeleteSnapshot"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	snapID := req.GetSnapshotId()
	if err := cs.TydsManager.DeleteSnapshot(snapID); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete snapshot: %v", err)
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots lists all snapshots, optionally for a specific volume.
func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	funcName := "ListSnapshots"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity returns the capacity of the storage pool.
func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	funcName := "GetCapacity"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerExpandVolume expands the volume to a new size.
func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	funcName := "ControllerExpandVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volumeId := req.GetVolumeId()
	requiredSize := req.GetCapacityRange().GetRequiredBytes()
	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeId)
	if acquired := cs.Locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer cs.Locks.Release(volumeId)

	// Convert bytes to megabytes
	requestSizeMB := int64(math.Ceil(float64(requiredSize) / 1024 / 1024))

	err := cs.TydsManager.ResizeVolume(volumeId, requestSizeMB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to resize volume: %v", err)
	}

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         requiredSize,
		NodeExpansionRequired: true, // Set this to true if node expansion is needed
	}, nil
}

// ControllerGetVolume returns information about a specific volume.
func (cs *ControllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	funcName := "ControllerGetVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	volId := req.GetVolumeId()
	volume, err := cs.TydsManager.FindVolume(volId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find volume: %v", err)
	}
	if volume == nil {
		return nil, status.Errorf(codes.NotFound, "Volume with ID '%s' not found", volId)
	}

	// Convert volume information to CSI volume
	csiVolume := &csi.Volume{
		VolumeId: volId,
		// Populate other necessary fields from your volume object
	}

	return &csi.ControllerGetVolumeResponse{
		Volume: csiVolume,
	}, nil
}

func (cs *ControllerServer) GetVolumeIDByName(volumeName string) (map[string]interface{}, error) {
	volumeList := cs.TydsManager.ListVolumes()
	for _, vol := range volumeList {
		if vol["blockName"].(string) == volumeName {
			return vol, nil
		}
	}
	// Returns an empty dictionary indicating that the volume with the
	// corresponding name was not found
	return make(map[string]interface{}), nil
}
