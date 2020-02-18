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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	ics "ics-csi-driver/pkg/common/icsphere"
	"k8s.io/klog"
	"strconv"
	"strings"
)

// CreateVolumeUtil is the helper function to create CNS volume
func CreateVolumeUtil(ctx context.Context, manager *Manager, spec *CreateVolumeSpec) (string, error) {
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
		klog.V(4).Infof("Failed to create volume %s with err: %v", createVolumeReq.Name, err)
		return volumeId, err
	} else {
		klog.V(4).Infof("Successfully created volume %s. volumeId: %s", spec.Name, volumeId)
		return volumeId, nil
	}
}

// AttachVolumeUtil is the helper function to attach CNS volume to specified vm
func AttachVolumeUtil(ctx context.Context, manager *Manager, vm *ics.VirtualMachine, volumeId string) (string, error) {
	diskUUID, err := manager.VolumeManager.AttachVolume(vm, volumeId)
	if err != nil {
		klog.Errorf("Failed to attach disk %s to VM %v with err %+v", volumeId, vm, err)
		return "", err
	}
	klog.V(4).Infof("Successfully attached disk %s to vm %s. Disk UUID:%s", volumeId, vm.VirtualMachine.Name, diskUUID)
	return diskUUID, nil
}

// DetachVolumeUtil is the helper function to detach CNS volume from specified vm
func DetachVolumeUtil(ctx context.Context, manager *Manager, vm *ics.VirtualMachine, volumeId string) error {
	err := manager.VolumeManager.DetachVolume(vm, volumeId)
	if err != nil {
		return err
	}
	klog.V(4).Infof("Successfully detached disk %s from vm %s", volumeId, vm.VirtualMachine.Name)
	return nil
}

// DeleteVolumeUtil is the helper function to delete CNS volume for given volumeId
func DeleteVolumeUtil(ctx context.Context, manager *Manager, volumeId string, deleteVolume bool) error {
	err := manager.VolumeManager.DeleteVolume(volumeId, deleteVolume)
	if err != nil {
		return err
	}

	klog.V(4).Infof("Successfully deleted volume %s", volumeId)
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
