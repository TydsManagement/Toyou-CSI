package rpcserver

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"toyou_csi/pkg/common"
	"toyou_csi/pkg/driver"
	"toyou_csi/pkg/service"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/util/resizefs"
	"k8s.io/kubernetes/pkg/volume"
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
func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	funcName := "NodeStageVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)
	// 0. Check node capability
	if flag := ns.Driver.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); !flag {
		return nil, status.Error(codes.Unimplemented, "Node has not stage capability")
	}

	// 1. Validate the request
	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	volCap := req.GetVolumeCapability()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if stagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging target path missing in request")
	}
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	// 2. skip staging if volume is in block mode
	if req.GetVolumeCapability().GetBlock() != nil {
		klog.Infof("Skipping staging of volume %s on path %s since it's in block mode", volumeID, stagingTargetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	// 3. ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeID)
	if acquired := ns.Locks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeID)
	}
	defer ns.Locks.Release(volumeID)

	// 4. set fsType
	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	// 5. Fetch the volume from the storage backend
	volInfo, err := ns.TydsManager.FindVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find volume %s: %v", volumeID, err)
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s not found", volumeID)
	}

	// 4. Prepare for staging the volume
	// if volume already mounted
	notMnt, err := mount.New("").IsLikelyNotMountPoint(stagingTargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(stagingTargetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	// already mount
	if !notMnt {
		return &csi.NodeStageVolumeResponse{}, nil
	}
	// 5. get device path
	devicePath := ""
	devicePrefix := "/dev/disk/by-id/virtio-"
	devicePath = devicePrefix + volumeID
	klog.Infof("Find volume %s's device path is %s", volumeID, devicePath)

	// 6. Mount the volume to the staging path
	// You may need to handle different VolumeAccessModes and filesystem types.
	// The following is a generic mount operation:
	// do mount
	klog.Infof("Mounting %s to %s format %s...", volumeID, stagingTargetPath, fsType)
	if err := ns.Mounter.FormatAndMount(devicePath, stagingTargetPath, fsType, []string{}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Mount %s to %s succeed", volumeID, stagingTargetPath)
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unstages a volume from a staging path.
func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	funcName := "NodeUnstageVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)
	if flag := ns.Driver.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); !flag {
		return nil, status.Error(codes.Unimplemented, "Node has not unstage capability")
	}
	// 1. Validate the request
	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if stagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging target path missing in request")
	}
	// 2. ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeID)
	if acquired := ns.Locks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeID)
	}
	defer ns.Locks.Release(volumeID)
	// 3. Check volume exist
	volInfo, err := ns.TydsManager.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}
	// 4. Unmount
	// check targetPath is mounted
	// For idempotent:
	// If the volume corresponding to the volume id is not staged to the staging target path,
	// the plugin MUST reply 0 OK.
	notMnt, err := ns.Mounter.IsLikelyNotMountPoint(stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if notMnt {
		return &csi.NodeUnstageVolumeResponse{}, nil
	}
	// count mount point
	_, cnt, err := mount.GetDeviceNameFromMount(ns.Mounter, stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// do unmount
	err = ns.Mounter.Unmount(stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Disk volume %s has been unmounted.", volumeID)
	cnt--
	klog.Infof("Disk volume mount count: %d", cnt)
	if cnt > 0 {
		klog.Errorf("Volume %s still mounted in instance %s", volumeID, ns.Driver.GetInstanceId())
		return nil, status.Error(codes.Internal, "unmount failed")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server.
func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	funcName := "NodeGetCapabilities"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.GetNodeCapability(),
	}, nil
}

// NodePublishVolume mounts the volume to the target path.
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	funcName := "NodePublishVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	// 1. Validate the request
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	volCap := req.GetVolumeCapability()
	stagePath := req.GetStagingTargetPath()

	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	} else if !ns.Driver.ValidateVolumeCapability(req.GetVolumeCapability()) {
		return nil, status.Error(codes.FailedPrecondition, "Exceed capabilities")
	}

	// 2. Fetch the volume from the storage backend
	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeID)
	if acquired := ns.Locks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, "Operation pending for volume %s", volumeID)
	}
	defer ns.Locks.Release(volumeID)
	klog.Infof("Successfully acquired lock for volume %s", volumeID)
	// determine if volume is in block mode
	isBlockMode := req.GetVolumeCapability().GetBlock() != nil
	// according to the CSI spec, CO is only responsible for ensuring the existence of the parent dir of the target path
	// in block mode, this is a nil, which will also work as expected in following mounting
	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	exists, err := ns.Mounter.ExistsPath(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !exists && isBlockMode {
		pathFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0750)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err = pathFile.Close(); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		// Create a directory
		if err := os.MkdirAll(targetPath, 0750); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// check wether targetPath is mounted
	notMnt, err := ns.Mounter.IsNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.Infof("Successfully mounted volume %s at path %s", volumeID, targetPath)
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}
	// set bind mount options
	options := []string{"bind"}
	if req.GetReadonly() {
		options = append(options, "ro")
	}
	klog.Infof("Bind mount %s at %s, isBlockMode: %t, fsType %s, options %v ...", stagePath, targetPath, isBlockMode, fsType, options)
	if err := ns.Mounter.Mount(stagePath, targetPath, fsType, options); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Mount bind %s at %s succeed", stagePath, targetPath)
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

	volInfo, err := ns.TydsManager.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}

	// 1. Unmount
	err = mount.CleanupMountPoint(targetPath, ns.Mounter.Interface, true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unmount target path %s error: %v", targetPath, err)
	}
	klog.Infof("Unbound mount volume succeed")

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetVolumeStats returns statistics about the volume.
func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	funcName := "NodeGetVolumeStats"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	// 1. 验证请求
	klog.Infof("%s: Check input arguments", hash)
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetVolumePath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume path missing in request")
	}

	volumePath := req.GetVolumePath()
	// block mode volume's stats can't be retrieved like those filesystem volumes
	pathType, err := ns.Mounter.GetFileType(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to determine volume %s 's mode: %v", volumePath, err)
	}
	isBlockMode := pathType == mount.FileTypeBlockDev

	// Get metrics
	if isBlockMode {
		blockSize, err := ns.getBlockSizeBytes(volumePath)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get block capacity on path %s: %v", volumePath, err)
		}
		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{
				{
					Unit:  csi.VolumeUsage_BYTES,
					Total: blockSize,
				},
			},
		}, nil
	}

	metricsStatFs := volume.NewMetricsStatFS(volumePath)
	metrics, err := metricsStatFs.GetMetrics()
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	klog.Infof("%s: Succeed to get metrics", hash)
	// 4. 创建并返回响应
	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: metrics.Available.Value(),
				Total:     metrics.Capacity.Value(),
				Used:      metrics.Used.Value(),
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: metrics.InodesFree.Value(),
				Total:     metrics.Inodes.Value(),
				Used:      metrics.InodesUsed.Value(),
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

// NodeExpandVolume expands the volume on the node.
func (ns *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	funcName := "NodeExpandVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)

	// 1. 验证请求
	klog.Info("Check input arguments")
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetVolumePath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume path missing in request")
	}
	requestSizeBytes, err := common.GetRequestSizeBytes(req.GetCapacityRange())
	if err != nil {
		return nil, status.Error(codes.OutOfRange, err.Error())
	}
	// Set parameter
	volumeID := req.GetVolumeId()
	volumePath := req.GetVolumePath()
	// 2. ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeID)
	if acquired := ns.Locks.TryAcquire(volumeID); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeID)
	}
	defer ns.Locks.Release(volumeID)
	// 2. Check volume exist
	klog.Infof("Get volume %s info", volumeID)
	volInfo, err := ns.TydsManager.FindVolume(volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeID)
	}

	// get device path
	devicePath := ""
	devicePrefix := "/dev/disk/by-id/virtio-"
	devicePath = devicePrefix + volumeID
	resizer := resizefs.NewResizeFs(ns.Mounter)
	klog.Infof("Resize file system device %s, mount path %s ...", devicePath, volumePath)
	ok, err := resizer.Resize(devicePath, volumePath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !ok {
		return nil, status.Error(codes.Internal, "failed to expand volume filesystem")
	}
	klog.Info("Succeed to resize file system")

	//  Check the block size
	blkSizeBytes, err := ns.getBlockSizeBytes(devicePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"expand volume error when getting size of block volume at path %s: %v", devicePath, err)
	}
	klog.Infof("Block size %d Byte, request size %d Byte", blkSizeBytes, requestSizeBytes)

	if blkSizeBytes < requestSizeBytes {
		// It's possible that somewhere the volume size was rounded up, getting more size than requested is a success
		return nil, status.Errorf(codes.Internal, "resize requested for %v but after resize volume was size %v",
			requestSizeBytes, blkSizeBytes)
	}
	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: blkSizeBytes,
	}, nil
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
		MaxVolumesPerNode: ns.Driver.GetMaxVolumePerNode(),
		AccessibleTopology: &csi.Topology{
			Segments: topology,
		},
	}, nil
}

// Helper functions

func (ns *NodeServer) getBlockSizeBytes(devicePath string) (int64, error) {
	output, err := ns.Mounter.Exec.Run("blockdev", "--getsize64", devicePath)
	if err != nil {
		return -1, fmt.Errorf("error when getting size of block volume at path %s: output: %s, err: %v", devicePath, string(output), err)
	}
	strOut := strings.TrimSpace(string(output))
	gotSizeBytes, err := strconv.ParseInt(strOut, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse size %s into int a size", strOut)
	}
	return gotSizeBytes, nil
}

func getTopologyInfo() map[string]string {
	// 实现逻辑以获取拓扑信息
	// 这可能包括区域、可用区等信息
	return map[string]string{
		"topology.kubernetes.io/region": "us-east-1",
		"topology.kubernetes.io/zone":   "us-east-1a",
	}
}

func getNodeID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		klog.Errorf("Failed to get hostname: %v", err)
		return "", err
	}
	return hostname, nil
}
