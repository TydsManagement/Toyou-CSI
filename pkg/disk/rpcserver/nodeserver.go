package rpcserver

import (
	"os"

	"toyou_csi/pkg/common"
	"toyou_csi/pkg/driver"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
)

type NodeServer struct {
	driver  *driver.DiskDriver
	mounter *mount.SafeFormatAndMount
	locks   *common.ResourceLocks
}

// NewNodeServer
// Create node server
func NewNodeServer(d *driver.DiskDriver) *NodeServer {
	return &NodeServer{
		driver: d,
		locks:  common.NewResourceLocks(),
	}
}

// csi.NodeStageVolumeRequest: 	volume id			+ Required
//
//	stage target path	+ Required
//	volume capability	+ Required
func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse,
	error) {
	funcName := "NodeStageVolume"
	info, hash := common.EntryFunction(funcName)
	defer klog.Info(common.ExitFunction(funcName, hash))
	klog.Info(info)
	if flag := ns.driver.ValidateNodeServiceRequest(csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME); !flag {
		return nil, status.Error(codes.Unimplemented, "Node has not stage capability")
	}
	// 0. Preflight
	// check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	// set parameter
	volumeId := req.GetVolumeId()
	targetPath := req.GetStagingTargetPath()

	// skip staging if volume is in block mode
	if req.GetVolumeCapability().GetBlock() != nil {
		klog.Infof("Skipping staging of volume %s on path %s since it's in block mode", volumeId, targetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	// ensure one call in-flight
	klog.Infof("Try to lock resource %s", volumeId)
	if acquired := ns.locks.TryAcquire(volumeId); !acquired {
		return nil, status.Errorf(codes.Aborted, common.OperationPendingFmt, volumeId)
	}
	defer ns.locks.Release(volumeId)
	// set fsType
	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	// Check volume exist
	volInfo, err := ns.cloud.FindVolume(volumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if volInfo == nil {
		return nil, status.Errorf(codes.NotFound, "Volume %s does not exist", volumeId)
	}
	// 1. Mount
	// if volume already mounted
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
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

	// get device path
	devicePath := ""
	devicePrefix := "/dev/disk/by-id/virtio-"
	if volInfo.Instance != nil && volInfo.Instance.Device != nil && *volInfo.Instance.Device != "" {
		// devicePath = *volInfo.Instance.Device
		devicePath = devicePrefix + volumeId
		klog.Infof("Find volume %s's device path is %s", volumeId, devicePath)
	} else {
		return nil, status.Errorf(codes.Internal, "Cannot find device path of volume %s", volumeId)
	}

	// do mount
	klog.Infof("Mounting %s to %s format %s...", volumeId, targetPath, fsType)
	if err := ns.mounter.FormatAndMount(devicePath, targetPath, fsType, []string{}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.Infof("Mount %s to %s succeed", volumeId, targetPath)
	return &csi.NodeStageVolumeResponse{}, nil
}
