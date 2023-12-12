package rpcserver

import (
	"context"
	"os"

	"toyou_csi/pkg/common"
	"toyou_csi/pkg/driver"
	"toyou_csi/pkg/service"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// NodeServer is the server API for Node service.
type NodeServer struct {
	Driver      *driver.ToyouDriver
	TydsManager service.TydsManager
	Mounter     *mount.SafeFormatAndMount
	Locks       *common.ResourceLocks
}

// NewNodeServer creates a new NodeServer.
func NewNodeServer(d *driver.ToyouDriver, tm service.TydsManager, mnt *mount.SafeFormatAndMount) *NodeServer {
	return &NodeServer{
		Driver:      d,
		TydsManager: tm,
		Mounter:     mnt,
	}
}

// NodeStageVolume stages a volume to a staging path.
func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	klog.Info("NodeStageVolume called")
	// 1. Validate the request
	if req.GetVolumeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if req.GetStagingTargetPath() == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging target path missing in request")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	volCap := req.GetVolumeCapability()

	// 2. Verify if the volume is already staged
	isStaged, err := isVolumeStaged(volumeID, stagingTargetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to verify if volume is staged: %v", err)
	}
	if isStaged {
		klog.Infof("Volume %s is already staged at path %s", volumeID, stagingTargetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	// 3. Fetch the volume from the storage backend
	volumeInterface, err := ns.TydsManager.FindVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find volume %s: %v", volumeID, err)
	}
	if volumeInterface == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s not found", volumeID)
	}

	// Dereference the pointer and then use type assertion
	volume, ok := volumeInterface.(*service.Volume)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Type assertion for volume %s failed", volumeID)
	}

	// Now you can use `volume` as *service.Volume

	// 4. Prepare for staging the volume
	// This step might include fetching necessary volume information like device path, etc.
	// For example, you might need to find the device path of the volume to mount it.

	devicePath := getDevicePath(volume) // Replace with actual logic to get the device path

	// 5. Mount the volume to the staging path
	// You may need to handle different VolumeAccessModes and filesystem types.
	// The following is a generic mount operation:
	if err := mountVolume(stagingTargetPath, devicePath, volCap); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to mount volume %s at path %s: %v", volumeID, stagingTargetPath, err)
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unstages a volume from a staging path.
func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	klog.Info("NodeUnstageVolume called")
	// Implement your logic here
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server.
func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.Info("NodeGetCapabilities called")
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			// Add other capabilities as needed
		},
	}, nil
}

// NodePublishVolume mounts the volume to the target path.
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	funcName := "NodePublishVolume"
	klog.Infof("Entering function %s", funcName)
	defer klog.Infof("Exiting function %s", funcName)

	// 1. Validate the request
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	volCap := req.GetVolumeCapability()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	// 2. Check if the volume is already published at the target path
	if isPublished, err := ns.isVolumePublished(volumeID, targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Error checking if volume is published: %v", err)
	} else if isPublished {
		klog.Infof("Volume %s is already published at path %s", volumeID, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}
	// 3. Fetch the volume from the storage backend
	volume, err := ns.fetchVolume(volumeID)
	if err != nil {
		return nil, err // fetchVolume will format the error appropriately
	}
	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeID)
	if acquired := ns.Locks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, "Operation pending for volume %s", volumeID)
	}
	defer ns.Locks.Release(volumeID)
	klog.Infof("Successfully acquired lock for volume %s", volumeID)

	// 4. Mount the volume to the target path
	// Implement the mounting logic here. The details depend on the volume type and access mode.
	if err := mountVolumeToTarget(targetPath, volume, volCap); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to mount volume %s at path %s: %v", volumeID, targetPath, err)
	}
	klog.Infof("Successfully mounted volume %s at path %s", volumeID, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path.
func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	funcName := "NodeUnpublishVolume"
	klog.Infof("Entering function %s", funcName)
	defer klog.Infof("Exiting function %s", funcName)

	// 1. Validate the request
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	// 2. Acquire lock
	klog.Infof("Attempting to acquire lock for volume %s", volumeID)
	if acquired := ns.Locks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, "Operation pending for volume %s", volumeID)
	}
	defer ns.Locks.Release(volumeID)
	klog.Infof("Successfully acquired lock for volume %s", volumeID)

	// 3.Check if the volume exists
	if volume, err := ns.TydsManager.FindVolume(volumeID); err != nil {
		return nil, err // findVolume will format the error appropriately
	} else if volume == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}

	// 4. Check if the volume is actually published at the target path
	isPublished, err := ns.isVolumePublished(volumeID, targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if volume is published: %v", err)
	}
	if !isPublished {
		klog.Infof("Volume %s is not published at path %s", volumeID, targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// 5. Unmount the volume from the target path
	if err := unmountVolume(targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to unmount volume %s from path %s: %v", volumeID, targetPath, err)
	}

	klog.Infof("Successfully unmounted volume %s from path %s", volumeID, targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetVolumeStats returns statistics about the volume.
func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	klog.Info("NodeGetVolumeStats called")

	// 1. 验证请求
	if req.GetVolumeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if req.GetVolumePath() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume path missing in request")
	}

	volumeID := req.GetVolumeId()
	volumePath := req.GetVolumePath()

	// 2. 确保卷在指定路径上已经挂载
	isMounted, err := isVolumeMounted(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if volume is mounted: %v", err)
	}
	if !isMounted {
		return nil, status.Errorf(codes.NotFound, "Volume %s not mounted at path %s", volumeID, volumePath)
	}

	// 3. 获取卷的统计信息
	stats, err := getVolumeStats(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get stats for volume %s: %v", volumeID, err)
	}

	// 4. 创建并返回响应
	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: stats.Available,
				Total:     stats.Total,
				Used:      stats.Used,
				Unit:      csi.VolumeUsage_BYTES,
			},
			// 可以添加其他统计信息，如 IOPS 或吞吐量
		},
	}, nil
}

// NodeExpandVolume expands the volume on the node.
func (ns *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	klog.Info("NodeExpandVolume called")

	// 1. 验证请求
	if req.GetVolumeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if req.GetVolumePath() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume path missing in request")
	}

	volumeID := req.GetVolumeId()
	volumePath := req.GetVolumePath()
	requiredBytes := req.GetCapacityRange().GetRequiredBytes()

	// 2. 确保卷在指定路径上已经挂载
	isMounted, err := isVolumeMounted(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if volume is mounted: %v", err)
	}
	if !isMounted {
		return nil, status.Errorf(codes.NotFound, "Volume %s not mounted at path %s", volumeID, volumePath)
	}

	// 3. 执行卷扩展
	if err := expandVolume(volumePath, requiredBytes); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to expand volume %s: %v", volumeID, err)
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}

// NodeGetInfo provides information about the node.
func (ns *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.Info("NodeGetInfo called")

	// 获取节点的相关信息。这可能包括节点ID、拓扑信息等。
	// 这里的具体实现依赖于你的环境和需求。
	nodeID, err := getNodeID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get node ID: %v", err)
	}

	topology := getTopologyInfo()

	return &csi.NodeGetInfoResponse{
		NodeId:            nodeID,
		MaxVolumesPerNode: getMaxVolumesPerNode(),
		AccessibleTopology: &csi.Topology{
			Segments: topology,
		},
	}, nil
}

// Helper functions

func getNodeID() (string, error) {
	// 实现逻辑以获取节点ID
	// 例如，可以从 Kubernetes 节点信息或特定于环境的配置中获取
	return "node-123", nil
}

func getTopologyInfo() map[string]string {
	// 实现逻辑以获取拓扑信息
	// 这可能包括区域、可用区等信息
	return map[string]string{
		"topology.kubernetes.io/region": "us-east-1",
		"topology.kubernetes.io/zone":   "us-east-1a",
	}
}

func getMaxVolumesPerNode() int64 {
	// 实现逻辑以返回节点上允许的最大卷数
	// 这个值可能取决于特定的环境或硬件限制
	return 255
}

// Helper functions to check if volume is already staged, to get device path, and to mount the volume
// You will need to implement these functions based on your storage backend.
// Helper functions
func expandVolume(volumePath string, requiredBytes int64) error {
	// 实现扩展卷的逻辑
	// 这可能包括文件系统的调整，或者其他与存储相关的操作
	return nil
}

func isVolumeMounted(volumePath string) (bool, error) {
	// 实现逻辑以检查卷是否已挂载
	return false, nil
}

type VolumeStats struct {
	Available, Total, Used int64
}

func getVolumeStats(volumePath string) (*VolumeStats, error) {
	// 实现逻辑以获取卷的统计信息
	return &VolumeStats{}, nil
}

func unmountVolume(targetPath string) error {
	// Implement the logic to unmount the volume
	return nil
}

func mountVolumeToTarget(targetPath string, volume *service.Volume, volCap *csi.VolumeCapability) error {
	// Implement the logic to mount the volume to the target path
	return nil
}

func isVolumeStaged(volumeID, targetPath string) (bool, error) {
	// Implement logic to check if the volume is already staged
	return false, nil
}

func getDevicePath(volume *service.Volume) string {
	// Implement logic to fetch the device path of the volume
	return ""
}

// fetchVolume fetches the volume from the storage backend
func (ns *NodeServer) fetchVolume(volumeID string) (*service.Volume, error) {
	volumeInterface, err := ns.TydsManager.FindVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find volume %s: %v", volumeID, err)
	}
	if volumeInterface == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s not found", volumeID)
	}

	volume, ok := volumeInterface.(*service.Volume)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Type assertion for volume %s failed", volumeID)
	}

	return volume, nil
}

// mountVolume mounts the volume to the target path
func (ns *NodeServer) mountVolume(targetPath string, volume *service.Volume, volCap *csi.VolumeCapability) error {
	// Implement the logic to mount the volume
	// Return any error encountered during the mount
}

// isVolumePublished checks if the volume is already published at the target path
func (ns *NodeServer) isVolumePublished(volumeID, targetPath string) (bool, error) {
	// Implement the logic to check if the volume is already published
	// Return true if published, false otherwise
	// Return any error encountered during the check
}

// createTargetMountPathIfNotExists creates the mountPath if it doesn't exist
// if in block volume mode, a file will be created
// otherwise a directory will be created
func (ns *NodeServer) createTargetMountPathIfNotExists(mountPath string, isBlockMode bool) error {
	exists, err := ns.Mounter.ExistsPath(mountPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if isBlockMode {
		pathFile, err := os.OpenFile(mountPath, os.O_CREATE|os.O_RDWR, 0750)
		if err != nil {
			return err
		}
		if err = pathFile.Close(); err != nil {
			return err
		}
	} else {
		// Create a directory
		if err := os.MkdirAll(mountPath, 0750); err != nil {
			return err
		}
	}

	return nil
}
