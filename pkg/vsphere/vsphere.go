package vsphere

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dougm/pretty"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/cns"
	cnstypes "github.com/vmware/govmomi/cns/types"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	vim25types "github.com/vmware/govmomi/vim25/types"
)

const (
	defaultClusterID           = "my-cluster-id"
	defaultClusterDistribution = "my-distribution"
)

type VsphereClient struct {
	client *govmomi.Client

	vcUser       string
	vcDatacenter string
	vcDatastore  string
}

func NewVsphereClient(vcHost, vcUser, vcPassword, vcDatacenter, vcDatastore string) (*VsphereClient, error) {
	ctx := context.Background()
	u := &url.URL{
		Scheme: "https",
		Host:   vcHost,
		Path:   "/sdk",
	}
	u.User = url.UserPassword(vcUser, vcPassword)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return nil, err
	}

	return &VsphereClient{
		client:       client,
		vcUser:       vcUser,
		vcDatacenter: vcDatacenter,
		vcDatastore:  vcDatastore,
	}, nil
}

func (c *VsphereClient) DetachVolume(ctx context.Context, volumeID, nodeID string) error {
	nodeVM, err := c.GetVMByUUID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to find vm by uuid %s: %s", nodeID, err.Error())
	}

	c.client.UseServiceVersion("vsan")
	cnsClient, err := cns.NewClient(ctx, c.client.Client)
	if err != nil {
		return fmt.Errorf("error to create cns client: %s", err.Error())
	}

	var cnsVolumeDetachSpecList []cnstypes.CnsVolumeAttachDetachSpec
	cnsVolumeDetachSpec := cnstypes.CnsVolumeAttachDetachSpec{
		VolumeId: cnstypes.CnsVolumeId{
			Id: volumeID,
		},
		Vm: nodeVM.Reference(),
	}
	cnsVolumeDetachSpecList = append(cnsVolumeDetachSpecList, cnsVolumeDetachSpec)
	fmt.Printf("Detaching volume using the spec: %+v\n", cnsVolumeDetachSpec)
	detachTask, err := cnsClient.DetachVolume(ctx, cnsVolumeDetachSpecList)
	if err != nil {
		return fmt.Errorf("Failed to detach volume. Error: %+v \n", err)
	}
	detachTaskInfo, err := cns.GetTaskInfo(ctx, detachTask)
	if err != nil {
		return fmt.Errorf("Failed to detach volume. Error: %+v \n", err)
	}
	detachTaskResult, err := cns.GetTaskResult(ctx, detachTaskInfo)
	if err != nil {
		return fmt.Errorf("Failed to detach volume. Error: %+v \n", err)
	}
	if detachTaskResult == nil {
		return fmt.Errorf("Empty detach task results")
	}
	detachVolumeOperationRes := detachTaskResult.GetCnsVolumeOperationResult()
	if detachVolumeOperationRes.Fault != nil {
		return fmt.Errorf("Failed to detach volume: fault=%+v", detachVolumeOperationRes.Fault)
	}

	fmt.Printf("Volume detached sucessfully\n")
	return nil
}

func (c *VsphereClient) AttachVolume(ctx context.Context, volumeID, nodeID string) (string, error) {
	nodeVM, err := c.GetVMByUUID(ctx, nodeID)
	if err != nil {
		return "", fmt.Errorf("failed to find vm by uuid %s: %s", nodeID, err.Error())
	}

	c.client.UseServiceVersion("vsan")
	cnsClient, err := cns.NewClient(ctx, c.client.Client)
	if err != nil {
		return "", fmt.Errorf("error to create cns client: %s", err.Error())
	}

	var cnsVolumeAttachSpecList []cnstypes.CnsVolumeAttachDetachSpec
	cnsVolumeAttachSpec := cnstypes.CnsVolumeAttachDetachSpec{
		VolumeId: cnstypes.CnsVolumeId{
			Id: volumeID,
		},
		Vm: nodeVM.Reference(),
	}
	cnsVolumeAttachSpecList = append(cnsVolumeAttachSpecList, cnsVolumeAttachSpec)
	fmt.Printf("Attaching volume using the spec: %+v\n", cnsVolumeAttachSpec)
	attachTask, err := cnsClient.AttachVolume(ctx, cnsVolumeAttachSpecList)
	if err != nil {
		return "", fmt.Errorf("Failed to attach volume. Error: %+v \n", err)
	}
	attachTaskInfo, err := cns.GetTaskInfo(ctx, attachTask)
	if err != nil {
		return "", fmt.Errorf("Failed to attach volume. Error: %+v \n", err)
	}
	attachTaskResult, err := cns.GetTaskResult(ctx, attachTaskInfo)
	if err != nil {
		return "", fmt.Errorf("Failed to attach volume. Error: %+v \n", err)
	}
	if attachTaskResult == nil {
		return "", fmt.Errorf("Empty attach task results")
	}
	attachVolumeOperationRes := attachTaskResult.GetCnsVolumeOperationResult()
	if attachVolumeOperationRes.Fault != nil {
		return "", fmt.Errorf("Failed to attach volume: fault=%+v", attachVolumeOperationRes.Fault)
	}
	diskUUID := interface{}(attachTaskResult).(*cnstypes.CnsVolumeAttachResult).DiskUUID
	fmt.Printf("Volume attached sucessfully. Disk UUID: %s\n", diskUUID)

	return diskUUID, nil
}

func (c *VsphereClient) DeleteVolume(ctx context.Context, volumeID string) error {
	c.client.UseServiceVersion("vsan")
	cnsClient, err := cns.NewClient(ctx, c.client.Client)
	if err != nil {
		return fmt.Errorf("error to create cns client: %s", err.Error())
	}

	var deleteVolumeIDList []cnstypes.CnsVolumeId
	deleteVolumeIDList = append(deleteVolumeIDList, cnstypes.CnsVolumeId{Id: volumeID})
	deleteTask, err := cnsClient.DeleteVolume(ctx, deleteVolumeIDList, true)
	if err != nil {
		return fmt.Errorf("failed to delete volume. Error: %+v \n", err)
	}
	deleteTaskInfo, err := cns.GetTaskInfo(ctx, deleteTask)
	if err != nil {
		return fmt.Errorf("Failed to delete volume. Error: %+v \n", err)
	}
	deleteTaskResult, err := cns.GetTaskResult(ctx, deleteTaskInfo)
	if err != nil {
		return fmt.Errorf("Failed to delete volume. Error: %+v \n", err)
	}
	if deleteTaskResult == nil {
		return fmt.Errorf("Empty delete task results")
	}
	deleteVolumeFromSnapshotOperationRes := deleteTaskResult.GetCnsVolumeOperationResult()
	if deleteVolumeFromSnapshotOperationRes.Fault != nil {
		return fmt.Errorf("Failed to delete volume: fault=%+v", deleteVolumeFromSnapshotOperationRes.Fault)
	}

	fmt.Printf("Volume: %q deleted sucessfully\n", volumeID)
	return nil
}

func (c *VsphereClient) CreateVolume(ctx context.Context, volumeName string, capacityInMb int64) (string, error) {
	datacenter := c.vcDatacenter
	datastore := c.vcDatastore

	c.client.UseServiceVersion("vsan")
	cnsClient, err := cns.NewClient(ctx, c.client.Client)
	if err != nil {
		return "", fmt.Errorf("error, %s", err.Error())
	}

	finder := find.NewFinder(c.client.Client, false)

	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return "", fmt.Errorf("error, %s", err.Error())
	}

	finder.SetDatacenter(dc)
	ds, err := finder.Datastore(ctx, datastore)
	if err != nil {
		return "", fmt.Errorf("error, %s", err.Error())
	}

	props := []string{"info", "summary"}
	pc := property.DefaultCollector(c.client.Client)
	var dsSummaries []mo.Datastore
	err = pc.Retrieve(ctx, []vim25types.ManagedObjectReference{ds.Reference()}, props, &dsSummaries)
	if err != nil {
		return "", fmt.Errorf("error, %s", err.Error())
	}

	var dsList []vim25types.ManagedObjectReference
	dsList = append(dsList, ds.Reference())

	containerCluster := cnstypes.CnsContainerCluster{
		ClusterType:         string(cnstypes.CnsClusterTypeKubernetes),
		ClusterId:           defaultClusterID,
		VSphereUser:         c.vcUser,
		ClusterFlavor:       string(cnstypes.CnsClusterFlavorVanilla),
		ClusterDistribution: defaultClusterDistribution,
	}

	var cnsVolumeCreateSpecList []cnstypes.CnsVolumeCreateSpec
	cnsVolumeCreateSpec := cnstypes.CnsVolumeCreateSpec{
		Name:       volumeName,
		VolumeType: string(cnstypes.CnsVolumeTypeBlock),
		Datastores: dsList,
		Metadata: cnstypes.CnsVolumeMetadata{
			ContainerCluster: containerCluster,
		},
		BackingObjectDetails: &cnstypes.CnsBlockBackingDetails{
			CnsBackingObjectDetails: cnstypes.CnsBackingObjectDetails{
				CapacityInMb: capacityInMb,
			},
		},
	}
	cnsVolumeCreateSpecList = append(cnsVolumeCreateSpecList, cnsVolumeCreateSpec)

	// TODO: query if already existing before creation, to make it idempotent
	fmt.Printf("Creating volume using the spec: %+v", pretty.Sprint(cnsVolumeCreateSpec))
	createTask, err := cnsClient.CreateVolume(ctx, cnsVolumeCreateSpecList)
	if err != nil {
		return "", fmt.Errorf("Failed to create volume. Error: %+v \n", err)
	}
	createTaskInfo, err := cns.GetTaskInfo(ctx, createTask)
	if err != nil {
		return "", fmt.Errorf("Failed to create volume. Error: %+v \n", err)
	}
	createTaskResult, err := cns.GetTaskResult(ctx, createTaskInfo)
	if err != nil {
		return "", fmt.Errorf("Failed to create volume. Error: %+v \n", err)
	}
	if createTaskResult == nil {
		return "", fmt.Errorf("Empty create task results")
	}
	createVolumeOperationRes := createTaskResult.GetCnsVolumeOperationResult()
	if createVolumeOperationRes.Fault != nil {
		return "", fmt.Errorf("Failed to create volume: fault=%+v", createVolumeOperationRes.Fault)
	}
	volumeId := createVolumeOperationRes.VolumeId.Id
	volumeCreateResult := (createTaskResult).(*cnstypes.CnsVolumeCreateResult)
	fmt.Printf("volumeCreateResult %+v\n", volumeCreateResult)
	fmt.Printf("Volume created sucessfully. volumeId: %s\n", volumeId)

	return volumeId, nil
}

func (c *VsphereClient) GetVMByUUID(ctx context.Context, uuid string) (*object.VirtualMachine, error) {
	searchIndex := object.NewSearchIndex(c.client.Client)
	reference, err := searchIndex.FindByUuid(ctx, nil, uuid, true, nil)
	if reference == nil {
		return nil, fmt.Errorf("failed to find object reference by uuid %s: %s", uuid, err.Error())
	}

	vm := object.NewVirtualMachine(c.client.Client, reference.Reference())
	return vm, nil
}
