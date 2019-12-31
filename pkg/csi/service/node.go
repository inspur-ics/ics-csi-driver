/*
Copyright 2018 The Kubernetes Authors.

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

package service

import (
	"context"
	_ "errors"
	_ "fmt"
	_ "io/ioutil"
	"os"
	_ "path"
	_ "path/filepath"
	_ "strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	//"github.com/akutz/gofsutil"
	//csictx "github.com/rexray/gocsi/context"
	//cnsvsphere "sigs.k8s.io/vsphere-csi-driver/pkg/common/cns-lib/vsphere"
	//cnsconfig "ics-csi-driver/pkg/common/config"
	//"ics-csi-driver/pkg/csi/service/common"
	//csitypes "ics-csi-driver/pkg/csi/types"
)

const (
	devDiskID   = "/dev/disk/by-id"
	blockPrefix = "wwn-0x"
	dmiDir      = "/sys/class/dmi"
)

func (s *service) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {

	klog.V(4).Infof("NodeStageVolume: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {

	klog.V(4).Infof("NodeUnstageVolume: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	klog.V(4).Infof("NodePublishVolume: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	klog.V(4).Infof("NodeUnpublishVolume: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {

	klog.V(4).Infof("NodeGetVolumeStats: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	klog.V(4).Infof("NodeGetCapabilities: called with args %+v", *req)

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

func (s *service) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {

	klog.V(4).Infof("NodeGetInfo: called with args %+v", *req)

	nodeID := os.Getenv("NODE_NAME")
	if nodeID == "" {
		return nil, status.Error(codes.Internal, "ENV NODE_NAME is not set")
	}

	topology := &csi.Topology{}
	return &csi.NodeGetInfoResponse{
		NodeId:             nodeID,
		AccessibleTopology: topology,
	}, nil
}

func (s *service) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {

	klog.V(4).Infof("NodeExpandVolume: called with args %+v", *req)
	return nil, status.Error(codes.Unimplemented, "")
}
