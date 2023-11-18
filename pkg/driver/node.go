package driver

import (
	"context"
	"fmt"
	"os"

	"github.com/akutz/gofsutil"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/wyike/my-csi/pkg/vsphere"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (*Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	fmt.Printf("NodeStageVolume was called\n")

	if req.VolumeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "VolumeID must be present in the NodeStageVolumeRequest")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Errorf(codes.InvalidArgument, "StagingTargetPath must be present in the NodeStageVolumeRequest")
	}

	if req.VolumeCapability == nil {
		return nil, status.Errorf(codes.InvalidArgument, "VolumeCapability must be present in the NodeStageVolumeRequest")
	}

	var diskUUID string
	if id, ok := req.PublishContext["diskUUID"]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "attribute disUUID is required in publish context")
	} else {
		diskUUID = id
	}

	// TODO: skip file share volume

	// skip staging block volume in raw blocking mode
	switch req.VolumeCapability.AccessType.(type) {
	case *csi.VolumeCapability_Block:
		fmt.Printf("node staging block volume: Skipping staging for block volume ID %s", req.VolumeId)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	fmt.Printf("node staging mount volume: format and mount the volume ID %s", req.VolumeId)

	// determine file system type
	fsType := "ext4"
	if req.VolumeCapability.GetMount().FsType != "" {
		fsType = req.VolumeCapability.GetMount().FsType
	}

	// get mount source and target
	source := getDeviceSource(diskUUID)
	target := req.StagingTargetPath

	// get mount flgs
	volCap := req.GetVolumeCapability()
	mountVol := volCap.GetMount()
	if mountVol == nil {
		return nil, status.Errorf(codes.InvalidArgument, "access type missing")
	}
	mntFlags := mountVol.GetMountFlags()

	fmt.Printf("nodeStageBlockVolume: Format and mount the device %s at %s with mount flags %v",
		source, target, mntFlags)
	// by default, format to nfsv4 fs type
	err := gofsutil.FormatAndMount(ctx, source, target, fsType, mntFlags...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in formating and mounting volume. Err: %s", err.Error())
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (*Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	stagingTarget := req.GetStagingTargetPath()
	volID := req.GetVolumeId()
	fmt.Printf("NodeUnstageVolume of node serivce was called to unmount target %s for volume %s\n", stagingTarget, volID)

	// TODO: more checkings on target existing or not
	if err := gofsutil.Unmount(ctx, stagingTarget); err != nil {
		return nil, status.Errorf(codes.Internal, "error unmounting stagingTarget: %s", err.Error())
	}

	fmt.Printf("NodeUnstageVolume successful for target %s for volume %s\n", stagingTarget, volID)

	return &csi.NodeUnstageVolumeResponse{}, nil
}
func (*Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// here we are just handling request for block volume in filesystem mode
	// TODO: get req.VolumeCaps and make sure that you handle request for block mode
	// TODO: block volume in raw block mode, the source is going to be the device dir where volume was attached form ControllerPubVolume RPC: getDeviceSource(diskUUID)
	// TODO: file share volume

	source := req.GetStagingTargetPath()
	target := req.GetTargetPath()
	fmt.Printf("NodePublishVolume was called with source %s and target %s\n", source, target)

	// get mount flgs
	volCap := req.GetVolumeCapability()
	mountVol := volCap.GetMount()
	if mountVol == nil {
		return nil, status.Errorf(codes.InvalidArgument, "access type missing")
	}
	mntFlags := mountVol.GetMountFlags()

	if req.Readonly {
		mntFlags = append(mntFlags, "ro")
	}

	// We are responsible for creating target dir:
	// eg: /var/lib/kubelet/pods/c8c18ba2-6d32-4537-acf2-a057d3a7d24e/volumes/kubernetes.io~csi/pvc-5541326c-abe1-47bc-a94c-de5ac1f37501/mount
	err := os.MkdirAll(target, 0777)
	if err != nil {
		return nil, fmt.Errorf("error: %s, creating the target dir\n", err.Error())
	}

	if err := gofsutil.BindMount(ctx, source, target, mntFlags...); err != nil {
		return nil, status.Errorf(codes.Internal, "error mounting volume. err: %s", err.Error())
	}

	fmt.Printf("NodePublishVolume for %s successful to path %s", req.GetVolumeId(), target)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (*Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	target := req.GetTargetPath()

	fmt.Printf("NodeUnpublishVolume of node serivce was called to unmount target %s for volume %s\n", target, volID)

	if err := gofsutil.Unmount(ctx, target); err != nil {
		return nil, status.Errorf(codes.Internal, "error unmounting target %s for volume %s. %s", target, volID, err.Error())
	}

	if err := os.Remove(target); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to remove target path: %s, err: %s", target, err.Error())
	}

	fmt.Printf("NodeUnpublishVolume successful for volume %s\n", volID)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}
func (*Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NodeGetVolumeStats not implemented")
}
func (*Driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NodeExpandVolume not implemented")
}

func (*Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	fmt.Printf("NodeGetCapabilities of node serivce was called\n")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (*Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	nodeID, err := vsphere.GetSystemUUID()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get system uuid for node VM with error: %v", err))
	}

	return &csi.NodeGetInfoResponse{
		NodeId:            nodeID,
		MaxVolumesPerNode: 10,
	}, nil
}
