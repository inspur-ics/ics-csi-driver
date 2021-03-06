/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cns

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"sort"
	"strings"

	"ics-csi-driver/pkg/common/config"
	"ics-csi-driver/pkg/common/ics"
	"ics-csi-driver/pkg/csi/service/common"
)

var (
	// controllerCaps represents the capability of controller service
	controllerCaps = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	}
)

type nodeManager interface {
	Initialize() error
	GetSharedDatastoresInK8SCluster(ctx context.Context) ([]*ics.DatastoreInfo, error)
	GetSharedDatastoresInTopology(ctx context.Context, topologyRequirement *csi.TopologyRequirement, zoneKey string, regionKey string) ([]*ics.DatastoreInfo, map[string][]map[string]string, error)
	GetNodeUUID(nodeName string) (string, error)
	GetNodeByName(nodeName string) (*ics.VirtualMachine, error)
}

type controller struct {
	manager *common.Manager
	nodeMgr nodeManager
}

// New creates a CNS controller
func New() common.Controller {
	return &controller{}
}

// Init is initializing controller struct
func (c *controller) Init(config *config.Config) error {
	klog.Infof("Initializing ics csi-controller")

	var err error
	vcenterconfig, err := ics.GetVirtualCenterConfig(config)
	if err != nil {
		klog.Errorf("Failed to get VirtualCenterConfig. err=%v", err)
		return err
	}
	vcManager := ics.GetVirtualCenterManager()
	vc, err := vcManager.RegisterVirtualCenter(vcenterconfig)
	if err != nil {
		klog.Errorf("Failed to register VC with virtualCenterManager. err=%v", err)
		return err
	}

	c.manager = &common.Manager{
		VcenterConfig:  vcenterconfig,
		CnsConfig:      config,
		VolumeManager:  ics.GetVolumeManager(vc),
		VcenterManager: ics.GetVirtualCenterManager(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vc, err = common.GetVCenter(ctx, c.manager)
	if err != nil {
		klog.Errorf("Failed to get vcenter. err=%v", err)
		return err
	}
	klog.Infof("Successfully get vcenter %+v", vc)

	c.nodeMgr = &Nodes{}
	err = c.nodeMgr.Initialize()
	if err != nil {
		klog.Errorf("Failed to initialize nodeMgr. err=%v", err)
		return err
	}
	return nil
}

// CreateVolume is creating CNS Volume using volume request specified
// in CreateVolumeRequest
func (c *controller) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {

	klog.V(4).Infof("CreateVolume: called with args %+v", *req)

	// Volume Size - Default is 10 GiB
	volSizeBytes := int64(common.DefaultGbDiskSize * common.GbInBytes)
	if req.GetCapacityRange() != nil && req.GetCapacityRange().RequiredBytes != 0 {
		volSizeBytes = int64(req.GetCapacityRange().GetRequiredBytes())
	}
	volSizeGB := int64(common.RoundUpSize(volSizeBytes, common.GbInBytes))

	var datastoreReq string
	var fsType string

	// Support case insensitive parameters
	for paramName := range req.Parameters {
		param := strings.ToLower(paramName)
		if param == common.AttributeDatastoreURL {
			datastoreReq = req.Parameters[paramName]
		} else if param == common.AttributeFsType {
			fsType = req.Parameters[common.AttributeFsType]
		}
	}

	var createVolumeSpec = common.CreateVolumeSpec{
		CapacityGB: volSizeGB,
		Name:       req.Name,
	}

	var err error
	var sharedDatastores []*ics.DatastoreInfo
	var datastoreTopologyMap = make(map[string][]map[string]string)

	// Get accessibility
	topologyRequirement := req.GetAccessibilityRequirements()
	if topologyRequirement != nil {
		// Get shared accessible datastores for matching topology requirement
		if c.manager.CnsConfig.Labels.Zone == "" || c.manager.CnsConfig.Labels.Region == "" {
			// if zone and region label not specified in the config secret, then return NotFound error.
			errMsg := fmt.Sprintf("Zone/Region category names not specified in the csi config secret")
			klog.Errorf(errMsg)
			return nil, status.Error(codes.NotFound, errMsg)
		}
		sharedDatastores, datastoreTopologyMap, err = c.nodeMgr.GetSharedDatastoresInTopology(ctx, topologyRequirement, c.manager.CnsConfig.Labels.Zone, c.manager.CnsConfig.Labels.Region)
		if err != nil || len(sharedDatastores) == 0 {
			msg := fmt.Sprintf("Failed to get shared datastores in topology: %+v. Error: %+v", topologyRequirement, err)
			klog.Errorf(msg)
			return nil, status.Error(codes.NotFound, msg)
		}
		klog.V(4).Infof("Shared datastores [%+v] retrieved for topologyRequirement [%+v] with datastoreTopologyMap [+%v]", sharedDatastores, topologyRequirement, datastoreTopologyMap)
	} else {
		// Get shared datastores for the Kubernetes cluster
		sharedDatastores, err = c.nodeMgr.GetSharedDatastoresInK8SCluster(ctx)
		if err != nil || len(sharedDatastores) == 0 {
			msg := fmt.Sprintf("Failed to get shared datastores in kubernetes cluster. Error: %+v", err)
			klog.Error(msg)
			return nil, status.Errorf(codes.Internal, msg)
		}
	}

	if datastoreReq != "" {
		// Check datastore ID specified in the storageclass is accessible
		isDataStoreAccessible := false
		for _, sharedDatastore := range sharedDatastores {
			if sharedDatastore.ID == datastoreReq || sharedDatastore.Name == datastoreReq {
				createVolumeSpec.DatastoreID = sharedDatastore.ID
				isDataStoreAccessible = true
				break
			}
		}
		if !isDataStoreAccessible {
			var errMsg string
			if topologyRequirement != nil {
				errMsg = fmt.Sprintf("Datastore: %s specified in the storage class is not accessible in the topology:[+%v]",
					datastoreReq, topologyRequirement)
			} else {
				errMsg = fmt.Sprintf("Datastore: %s specified in the storage class is not accessible", datastoreReq)
			}
			klog.Errorf(errMsg)
			return nil, status.Error(codes.InvalidArgument, errMsg)
		}
	} else if topologyRequirement != nil {
		sort.Slice(sharedDatastores, func(i, j int) bool { return sharedDatastores[i].AvailCapacity > sharedDatastores[j].AvailCapacity })
		createVolumeSpec.DatastoreID = sharedDatastores[0].ID
	} else {
		errMsg := fmt.Sprintf("Datastore not specified in the storage class. CreateVolumeRequest: %+v", *req)
		klog.Errorf(errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	volumeID, err := common.CreateVolumeUtil(ctx, c.manager, &createVolumeSpec)
	if err != nil {
		msg := fmt.Sprintf("Failed to create volume. Error: %+v", err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	attributes := make(map[string]string)
	attributes[common.AttributeDiskType] = common.DiskTypeString
	attributes[common.AttributeFsType] = fsType
	resp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: int64(volSizeGB * common.GbInBytes),
			VolumeContext: attributes,
		},
	}

	if topologyRequirement != nil {
		volumeTopology := &csi.Topology{
			Segments: datastoreTopologyMap[createVolumeSpec.DatastoreID][0],
		}
		resp.Volume.AccessibleTopology = append(resp.Volume.AccessibleTopology, volumeTopology)
	}
	return resp, nil
}

// CreateVolume is deleting CNS Volume specified in DeleteVolumeRequest
func (c *controller) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {

	klog.V(4).Infof("DeleteVolume: called with args %+v", *req)

	var err error
	err = common.ValidateDeleteVolumeRequest(req)
	if err != nil {
		return nil, err
	}
	err = common.DeleteVolumeUtil(ctx, c.manager, req.VolumeId, true)
	if err != nil {
		msg := fmt.Sprintf("Failed to delete volume: %q. Error: %+v", req.VolumeId, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches a volume to the Node VM.
// volume id and node name is retrieved from ControllerPublishVolumeRequest
func (c *controller) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {
	klog.V(4).Infof("ControllerPublishVolume: called with args %+v", *req)

	err := common.ValidateControllerPublishVolumeRequest(req)
	if err != nil {
		msg := fmt.Sprintf("Validation for PublishVolume Request: %+v has failed. Error: %v", *req, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	node, err := c.nodeMgr.GetNodeByName(req.NodeId)
	if err != nil {
		msg := fmt.Sprintf("Failed to find VirtualMachine for node:%q. Error: %v", req.NodeId, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	klog.V(4).Infof("Found VirtualMachine for node:%q.", req.NodeId)

	diskUUID, err := common.AttachVolumeUtil(ctx, c.manager, node, req.VolumeId)
	if err != nil {
		klog.Errorf("ControllerPublishVolume: failed with err: %v", err)
	}

	publishInfo := make(map[string]string)
	publishInfo[common.AttributeDiskType] = common.DiskTypeString
	publishInfo[common.AttributeFirstClassDiskUUID] = diskUUID
	resp := &csi.ControllerPublishVolumeResponse{
		PublishContext: publishInfo,
	}
	return resp, nil
}

// ControllerUnpublishVolume detaches a volume from the Node VM.
// volume id and node name is retrieved from ControllerUnpublishVolumeRequest
func (c *controller) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	klog.V(4).Infof("ControllerUnpublishVolume: called with args %+v", *req)
	err := common.ValidateControllerUnpublishVolumeRequest(req)
	if err != nil {
		msg := fmt.Sprintf("Validation for UnpublishVolume Request: %+v has failed. Error: %v", *req, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	node, err := c.nodeMgr.GetNodeByName(req.NodeId)
	if err != nil {
		msg := fmt.Sprintf("Failed to find VirtualMachine for node:%q. Error: %v", req.NodeId, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	err = common.DetachVolumeUtil(ctx, c.manager, node, req.VolumeId)
	if err != nil {
		msg := fmt.Sprintf("Failed to detach disk: %+q from node: %q err: %+v", req.VolumeId, req.NodeId, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	resp := &csi.ControllerUnpublishVolumeResponse{}
	return resp, nil
}

// ValidateVolumeCapabilities returns the capabilities of the volume.
func (c *controller) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	klog.V(4).Infof("ValidateVolumeCapabilities: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controller) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {

	klog.V(4).Infof("ListVolumes: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controller) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {

	klog.V(4).Infof("GetCapacity: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controller) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	klog.V(4).Infof("ControllerGetCapabilities: called with args %+v", *req)
	var caps []*csi.ControllerServiceCapability
	for _, cap := range controllerCaps {
		c := &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
		caps = append(caps, c)
	}
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (c *controller) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {

	klog.V(4).Infof("CreateSnapshot: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controller) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {

	klog.V(4).Infof("DeleteSnapshot: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controller) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {

	klog.V(4).Infof("ListSnapshots: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controller) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error) {

	klog.V(5).Infof("ControllerExpandVolume: called with args %+v", *req)

	volumeID := req.GetVolumeId()
	volSizeBytes := int64(req.GetCapacityRange().GetRequiredBytes())
	volSizeGB := float64(common.RoundUpSize(volSizeBytes, common.GbInBytes))

	err := common.ExpandVolumeUtil(ctx, c.manager, volumeID, volSizeGB)
	if err != nil {
		msg := fmt.Sprintf("failed to expand volume: %q to size: %d with error: %+v", volumeID, volSizeGB, err)
		klog.Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	nodeExpansionRequired := true
	// Node expansion is not required for raw block volumes
	if _, ok := req.GetVolumeCapability().GetAccessType().(*csi.VolumeCapability_Block); ok {
		nodeExpansionRequired = false
	}
	resp := &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(volSizeGB * float64(common.GbInBytes)),
		NodeExpansionRequired: nodeExpansionRequired,
	}

	klog.V(5).Infof("ControllerExpandVolume: resp %+v", *resp)
	return resp, nil
}
