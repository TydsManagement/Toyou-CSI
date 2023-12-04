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

package rpcserver

import (
	"context"

	"toyou_csi/pkg/driver"
	"toyou_csi/pkg/service"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// ControllerServer implements the CSI controller service.
type ControllerServer struct {
	Driver      *driver.ToyouDriver
	TydsManager service.TydsManager
}

// NewControllerServer creates a new ControllerServer.
func NewControllerServer(d *driver.ToyouDriver, tm service.TydsManager) *ControllerServer {
	return &ControllerServer{
		Driver:      d,
		TydsManager: tm,
	}
}

// CreateVolume creates a new volume.
func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.Info("CreateVolume called")

	volName := req.GetName()
	requestSize := req.GetCapacityRange().GetRequiredBytes()
	replicas := 1         // Assuming a default value, adjust as needed.
	volType := 0          // Assuming a default value, adjust as needed.
	zone := ""            // Extract from parameters or other parts of the request as needed.
	containerConfID := "" // Extract as needed.

	volId, err := cs.TydsManager.CreateVolume(volName, int(requestSize), replicas, volType, zone, containerConfID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create volume: %v", err)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volId,
			CapacityBytes: requestSize,
			VolumeContext: req.GetParameters(),
		},
	}, nil
}

// DeleteVolume deletes a volume.
func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.Info("DeleteVolume called")

	volId := req.GetVolumeId()
	if volId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}

	err := cs.TydsManager.DeleteVolume(volId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete volume: %v", err)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ListVolumes lists all volumes managed by the storage system.
func (cs *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.Info("ListVolumes called")

	// Implement logic to list all volumes. This will depend on how your storage system manages volume information.
	// Here we return an empty list as an example.
	return &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{},
	}, nil
}

// ControllerPublishVolume attaches a volume to a node.
func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.Info("ControllerPublishVolume called")

	volId := req.GetVolumeId()
	nodeId := req.GetNodeId()
	if volId == "" || nodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID and Node ID are required")
	}

	err := cs.TydsManager.AttachVolume(volId, nodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to attach volume: %v", err)
	}

	return &csi.ControllerPublishVolumeResponse{}, nil
}

// ControllerUnpublishVolume detaches a volume from a node.
func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.Info("ControllerUnpublishVolume called")

	volId := req.GetVolumeId()
	nodeId := req.GetNodeId()
	if volId == "" || nodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID and Node ID are required")
	}

	err := cs.TydsManager.DetachVolume(volId, nodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to detach volume: %v", err)
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks if a volume has the given capabilities.
func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	klog.Info("ValidateVolumeCapabilities called")

	// This implementation is a placeholder. Replace it with your actual validation logic.
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.GetVolumeCapabilities(),
		},
	}, nil
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.Info("ControllerGetCapabilities called")
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			// Add other capabilities as needed
		},
	}, nil
}

// CreateSnapshot creates a new snapshot from a volume.
func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	klog.Info("CreateSnapshot called")

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
	klog.Info("DeleteSnapshot called")

	snapID := req.GetSnapshotId()
	if err := cs.TydsManager.DeleteSnapshot(snapID); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete snapshot: %v", err)
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots lists all snapshots, optionally for a specific volume.
func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	klog.Info("ListSnapshots called")

	// Implement the logic to retrieve and list all snapshots.
	// This is a simplified example assuming no pagination and only listing all snapshots.
	return &csi.ListSnapshotsResponse{
		Entries: []*csi.ListSnapshotsResponse_Entry{
			// Populate with actual snapshot entries
		},
	}, nil
}

// GetCapacity returns the capacity of the storage pool.
func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	klog.Info("GetCapacity called")

	// Implement logic to return the total capacity of the storage.
	// This is often a static value for a given storage system or could be calculated based on available resources.
	// Here we return a placeholder value.
	totalCapacity := int64(1000000000000) // 1 TB as an example

	return &csi.GetCapacityResponse{
		AvailableCapacity: totalCapacity,
	}, nil
}

// ControllerExpandVolume expands the volume to a new size.
func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	klog.Info("ControllerExpandVolume called")

	volumeId := req.GetVolumeId()
	requiredSize := req.GetCapacityRange().GetRequiredBytes()

	err := cs.TydsManager.ResizeVolume(volumeId, int(requiredSize))
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
	klog.Info("ControllerGetVolume called")

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
