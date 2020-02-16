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

package common

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ics "ics-csi-driver/pkg/common/icsphere"
	"ics-csi-driver/pkg/common/rest"
	"k8s.io/klog"
	"strconv"
	"strings"
)

// CreateVolumeUtil is the helper function to create CNS volume
func CreateVolumeUtil(ctx context.Context, manager *Manager, spec *CreateVolumeSpec) (string, error) {
	klog.V(4).Infof("creating volume %s with create spec %+v", spec.Name, *spec)

	createVolumeReq := types.VolumeReq{
		Name:          spec.Name,
		Size:          strconv.FormatInt(spec.CapacityGB, 10),
		DataStoreId:   spec.DatastoreID,
		DataStoreType: "LOCAL",
		VolumePolicy:  "THIN",
		Description:   "CSI Persistent Volume",
		Bootable:      false,
		Shared:        false,
	}

	volumeId, err := manager.VolumeManager.CreateVolume(createVolumeReq)
	if err != nil {
		klog.V(4).Infof("create volume failed  with args %+v", createVolumeReq)
		return volumeId, err
	} else {
		klog.V(4).Infof("successfully created volume %s. volumeId: %s", spec.Name, volumeId)
		return volumeId, nil
	}
}

func AttachVolumeUtil(ctx context.Context, vmUuid string, vmName string, volumeId string) (string, error) {
	scsiId, vmId := "", ""
	rp, err := rest.NewRestProxy()
	if err != nil {
		klog.Error("create restProxy failed.")
		return scsiId, err
	}

	vmList, err := rest.GetVmList(rp)
	if err != nil {
		klog.Error("get vm list failed.")
		return scsiId, err
	}

	if len(vmList) > 0 {
		for _, vmInfo := range vmList {
			if vmInfo.Uuid == vmUuid {
				vmId = vmInfo.Id
				klog.V(5).Infof("find vm %s %s with id %s", vmUuid, vmName, vmId)
				break
			}
		}
	}

	if vmId == "" {
		klog.Errorf("vm %s %s not found", vmUuid, vmName)
		return scsiId, status.Errorf(codes.Internal,
			"vm %s %s not found", vmUuid, vmName)
	}

	vmInfo, err := rest.GetVmInfo(rp, vmId)
	if err != nil {
		klog.Error("get vm info failed.")
		return scsiId, err
	}

	volumeInfo, err := rest.GetVolumeInfo(rp, volumeId)
	if err != nil {
		klog.Error("get volume info failed.")
		return scsiId, err
	}

	volId := volumeInfo.Id
	diskInfo := rest.VmDiskInfo{
		Id:             volumeInfo.Id,
		Label:          "",
		ScsiId:         fmt.Sprintf("6%.15s", volId[len(volId)-15:]),
		Enabled:        false,
		Volume:         volumeInfo,
		BusModel:       "SCSI",
		ReadWriteModel: "NONE",
		EnableNativeIO: false,
	}
	vmInfo.Disks = append(vmInfo.Disks, diskInfo)
	for _, disk := range vmInfo.Disks {
		if disk.BusModel == "SCSI" {
			disk.Volume.DiskType = "SAS"
		}
	}
	vmInfo.VncPasswd = "00000000"
	klog.V(4).Infof("attaching volume %s to vm %s %s", volumeId, vmUuid, vmName)
	taskId, err := rest.SetVmInfo(rp, vmInfo)
	if err != nil {
		klog.Errorf("attach volume %s to vm %s %s failed.", volumeId, vmUuid, vmName)
		return scsiId, err
	}

	taskStat, err := rest.GetTaskState(rp, taskId)
	if err != nil {
		return scsiId, err
	} else if taskStat != "FINISHED" {
		return scsiId, status.Errorf(codes.Internal,
			"set vm info failed: taskId %s stat %s", taskId, taskStat)
	}

	klog.V(4).Infof("successfully attached disk %s to vm %s %s. scsi wwn: %s",
		volumeId, vmUuid, vmName, diskInfo.ScsiId)
	return diskInfo.ScsiId, nil
}

func DeleteVolumeUtil(ctx context.Context, volumeId string, deleteVolume bool) error {
	rp, err := rest.NewRestProxy()
	if err != nil {
		klog.Error("create restProxy failed.")
		return err
	}

	klog.V(4).Infof("deleting volume: %s", volumeId)
	taskId, err := rest.DeleteVolume(rp, volumeId, deleteVolume)
	if err != nil {
		klog.Errorf("delete volume %s failed.", volumeId)
		return err
	}

	taskStat, err := rest.GetTaskState(rp, taskId)
	if err != nil {
		klog.Errorf("get task state failed.")
		return err
	} else if taskStat != "FINISHED" {
		return status.Errorf(codes.Internal,
			"delete volume %s failed: taskId %s stat %s", volumeId, taskId, taskStat)
	}

	klog.V(4).Infof("successfully deleted volume: %s", volumeId)
	return nil
}

// GetVCenter returns VirtualCenter object from specified Manager object.
// Before returning VirtualCenter object, vcenter connection is established if session doesn't exist.

func GetVCenter(ctx context.Context, manager *Manager) (*ics.VirtualCenter, error) {
	var err error
	vcenter, err := manager.VcenterManager.GetVirtualCenter(manager.VcenterConfig.Host)
	if err != nil {
		klog.Errorf("Failed to get VirtualCenter instance for host: %q. err=%v", manager.VcenterConfig.Host, err)
		return nil, err
	}

	err = vcenter.Connect(ctx)
	if err != nil {
		klog.Errorf("Failed to connect to VirtualCenter host: %q. err=%v", manager.VcenterConfig.Host, err)
		return nil, err
	}

	return vcenter, nil
}

// GetUUIDFromProviderID Returns VM UUID from Node's providerID
func GetUUIDFromProviderID(providerID string) string {
	return strings.TrimPrefix(providerID, ProviderPrefix)
}

// FormatDiskUUID removes any spaces and hyphens in UUID
// Example UUID input is 42375390-71f9-43a3-a770-56803bcd7baa and output after format is 4237539071f943a3a77056803bcd7baa
func FormatDiskUUID(uuid string) string {
	uuidwithNoSpace := strings.Replace(uuid, " ", "", -1)
	uuidWithNoHypens := strings.Replace(uuidwithNoSpace, "-", "", -1)
	return strings.ToLower(uuidWithNoHypens)
}

// RoundUpSize calculates how many allocation units are needed to accommodate
// a volume of given size.
func RoundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	roundedUp := volumeSizeBytes / allocationUnitBytes
	if volumeSizeBytes%allocationUnitBytes > 0 {
		roundedUp++
	}
	return roundedUp
}

// GetLabelsMapFromKeyValue creates a  map object from given parameter
/*
func GetLabelsMapFromKeyValue(labels []types.KeyValue) map[string]string {
	labelsMap := make(map[string]string)
	for _, label := range labels {
		labelsMap[label.Key] = label.Value
	}
	return labelsMap
}
*/
// IsValidVolumeCapabilities is the helper function to validate capabilities of volume.
func IsValidVolumeCapabilities(volCaps []*csi.VolumeCapability) bool {
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, c := range VolumeCaps {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}
	foundAll := true
	for _, c := range volCaps {
		if !hasSupport(c) {
			foundAll = false
		}
	}
	return foundAll
}
