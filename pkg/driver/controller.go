package driver

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	fmt.Println("CreateVolume of the controller service was called")

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume must be called with a req name")
	}

	sizeBytes := req.CapacityRange.GetRequiredBytes()
	const mb = 1024 * 1024

	// Support create block volume only, by default in filesystem mode
	volumeId, err := d.vsphereClient.CreateVolume(ctx, req.Name, sizeBytes/mb)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to provsion the volume error %s\n", err.Error()))
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeId,
			CapacityBytes: sizeBytes,
		},
	}, nil
}

func (d *Driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	fmt.Println("DeleteVolume of the controller service was called")

	err := d.vsphereClient.DeleteVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to delete the volume error %s\n", err.Error()))
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	fmt.Println("ControllerPublishVolume of the controller service was called")

	if req.VolumeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "VolumeID is required in ControllerPublishVolume request")
	}

	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodeId is required in ControllerPublishVolume request")
	}

	// TODO: check volume is existing

	// attach volume to the node
	diskUUID, err := d.vsphereClient.AttachVolume(ctx, req.VolumeId, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to attach volume %s to node %s: %s", req.VolumeId, req.NodeId, err.Error())
	}

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			"diskUUID": FormatDiskUUID(diskUUID),
			"message":  "hello pv is attached!",
		},
	}, nil
}

func (d *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	fmt.Println("ControllerUnpublishVolume of the controller service was called")

	if req.VolumeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "VolumeID is required in ControllerUnpublishVolume request")
	}

	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodeId is required in ControllerUnpublishVolume request")
	}

	// attach volume to the node
	err := d.vsphereClient.DetachVolume(ctx, req.VolumeId, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fail to detach volume %s from node %s: %s", req.VolumeId, req.NodeId, err.Error())
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (d *Driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateVolumeCapabilities not implemented")
}
func (d *Driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListVolumes not implemented")
}
func (d *Driver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCapacity not implemented")
}

func (d *Driver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	caps := []*csi.ControllerServiceCapability{}
	for _, c := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	} {
		caps = append(caps, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}
func (d *Driver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSnapshot not implemented")
}
func (d *Driver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSnapshot not implemented")
}
func (d *Driver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSnapshots not implemented")
}
func (d *Driver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerExpandVolume not implemented")
}
func (d *Driver) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerGetVolume not implemented")
}
func (d *Driver) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerModifyVolume not implemented")
}
