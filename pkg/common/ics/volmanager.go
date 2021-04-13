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

package ics

import (
	"context"
	"errors"
	"fmt"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	icsvm "github.com/inspur-ics/ics-go-sdk/vm"
	icsvol "github.com/inspur-ics/ics-go-sdk/volume"
	"k8s.io/klog"
	"sync"
)

// VolumeManager provides functionality to manage volumes.
type VolumeManager interface {
	// CreateVolume creates a new volume given its spec.
	CreateVolume(req types.VolumeReq) (string, error)
	// DeleteVolume deletes a volume given its spec.
	DeleteVolume(volumeId string, deleteVolume bool) error
	// ExpandVolume expands a volume given its spec.
	ExpandVolume(volumeId string, capacityInGb float64) error
	// AttachVolume attaches a volume to a virtual machine given the spec.
	AttachVolume(vm *VirtualMachine, volumeId string) (string, error)
	// DetachVolume detaches a volume from the virtual machine given the spec.
	DetachVolume(vm *VirtualMachine, volumeId string) error
}

var (
	// managerInstance is a Manager singleton.
	managerInstance *volumeManager
	// onceForManager is used for initializing the Manager singleton.
	onceForManager sync.Once
)

// GetManager returns the Manager singleton.
func GetVolumeManager(vc *VirtualCenter) VolumeManager {
	onceForManager.Do(func() {
		klog.V(1).Infof("Initializing volumeManager...")
		managerInstance = &volumeManager{
			virtualCenter: vc,
		}
		klog.V(1).Infof("volumeManager initialized")
	})
	return managerInstance
}

// DefaultManager provides functionality to manage volumes.
type volumeManager struct {
	virtualCenter *VirtualCenter
}

func validateManager(m *volumeManager) error {
	if m.virtualCenter == nil {
		klog.Error(
			"Virtual Center connection not established")
		return errors.New("Virtual Center connection not established")
	}
	return nil
}

// CreateVolume creates a new volume given its spec.
func (m *volumeManager) CreateVolume(req types.VolumeReq) (string, error) {
	err := validateManager(m)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = m.virtualCenter.Connect(ctx)
	if err != nil {
		klog.Errorf("Virtual Center Connect failed with err: %+v", err)
		return "", err
	}

	volService := icsvol.NewVolumeService(m.virtualCenter.Client)
	task, err := volService.CreateVolume(ctx, req)
	if err != nil {
		klog.Errorf("Create volume %+v failed with err: %+v", req, err)
		return "", err
	}

	klog.V(5).Infof("Creating volume %+v task info: %+v", req, task)
	taskState, err := GetTaskState(ctx, m.virtualCenter, &task)
	if err != nil {
		klog.Errorf("Create volume %+v task failed with err: %+v", req, err)
		return "", err
	} else if taskState != "FINISHED" {
		errMsg := fmt.Sprintf("Create volume task state %s", taskState)
		klog.Errorf(errMsg)
		return "", errors.New(errMsg)
	}
	klog.V(5).Infof("Create volume %s task finished", req.Name)

	volList, err := volService.GetVolumesInDatastore(ctx, req.DataStoreId)
	if err != nil {
		klog.Errorf("Failed to get volume list in storage %s with err: %+v", req.DataStoreId, err)
		return "", err
	}
	for _, volInfo := range volList {
		if volInfo.Name == req.Name {
			return volInfo.ID, nil
		}
	}

	errMsg := fmt.Sprintf("Volume %s not found in storage %s. Create volume failed.", req.Name, req.DataStoreId)
	klog.Errorf(errMsg)
	return "", errors.New(errMsg)
}

// DeleteVolume deletes a volume given id.
func (m *volumeManager) DeleteVolume(volumeId string, deleteVolume bool) error {
	err := validateManager(m)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = m.virtualCenter.Connect(ctx)
	if err != nil {
		klog.Errorf("Virtual Center Connect failed with err: %+v", err)
		return err
	}

	volService := icsvol.NewVolumeService(m.virtualCenter.Client)
	task, err := volService.DeleteVolume(ctx, volumeId, deleteVolume)
	if err != nil {
		klog.Errorf("Delete volume %s failed with err: %+v", volumeId, err)
		return err
	}

	klog.V(5).Infof("Deleting volume %s task info: %+v", volumeId, task)
	taskState, err := GetTaskState(ctx, m.virtualCenter, &task)
	if err != nil {
		klog.Errorf("Deleting volume %s task failed with err: %+v", volumeId, err)
		return err
	} else if taskState != "FINISHED" {
		errMsg := fmt.Sprintf("Delete volume %s task state %s", volumeId, taskState)
		klog.Errorf(errMsg)
		return errors.New(errMsg)
	}
	klog.V(5).Infof("Delete volume %s task finished", volumeId)
	return nil
}

// ExpandVolume expands a volume given id.
func (m *volumeManager) ExpandVolume(volumeId string, capacityInGb float64) error {
	err := validateManager(m)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = m.virtualCenter.Connect(ctx)
	if err != nil {
		klog.Errorf("iCenter Connect failed with err: %+v", err)
		return err
	}

	volService := icsvol.NewVolumeService(m.virtualCenter.Client)
	volInfo, err := volService.GetVolumeInfoById(ctx, volumeId)
	if err != nil {
		klog.Errorf("Get volume %s info failed with err: %+v", volumeId, err)
		return err
	}

	volInfo.Size = capacityInGb
	task, err := volService.SetVolume(ctx, volumeId, volInfo)
	if err != nil {
		klog.Errorf("Expand volume %s failed with err: %+v", volumeId, err)
		return err
	}

	klog.V(5).Infof("Expanding volume %s task info: %+v", volumeId, task)
	taskState, err := GetTaskState(ctx, m.virtualCenter, &task)
	if err != nil {
		klog.Errorf("Expand volume %s task failed with err: %+v", volumeId, err)
		return err
	} else if taskState != "FINISHED" {
		errMsg := fmt.Sprintf("Expand volume task state %s", taskState)
		klog.Errorf(errMsg)
		return errors.New(errMsg)
	}

	klog.V(5).Infof("Expand volume %s task finished", volumeId)
	return nil
}

// AttachVolume attaches a volume to a virtual machine given the spec.
func (m *volumeManager) AttachVolume(vm *VirtualMachine, volumeId string) (string, error) {
	err := validateManager(m)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = m.virtualCenter.Connect(ctx)
	if err != nil {
		klog.Errorf("Virtual Center Connect failed with err: %+v", err)
		return "", err
	}

	volService := icsvol.NewVolumeService(m.virtualCenter.Client)
	volInfo, err := volService.GetVolumeInfoById(ctx, volumeId)
	if err != nil {
		klog.Errorf("Get volume %s info failed with err: %+v", volumeId, err)
		return "", err
	}

	diskInfo := types.Disk{
		ID:             volumeId,
		Enabled:        false,
		Volume:         volInfo,
		BusModel:       "SCSI",
		ReadWriteModel: "NONE",
		EnableNativeIO: false,
		QueueNum:       1,
	}

	vmInfo := *vm.VirtualMachine
	vmInfo.Disks = append(vmInfo.Disks, diskInfo)
	for _, disk := range vmInfo.Disks {
		if disk.BusModel == "SCSI" {
			disk.Volume.DiskType = "SAS"
		}
	}
	vmInfo.VncPasswd = "00000000"
	klog.V(4).Infof("Attaching volume %s to VM %v", volumeId, vm)

	vmService := icsvm.NewVirtualMachineService(m.virtualCenter.Client)
	task, err := vmService.SetVM(ctx, vmInfo)
	if err != nil {
		klog.Errorf("Failed to attach volume %s to VM %v with err: %+v", volumeId, vm, err)
		return "", err
	}

	klog.V(5).Infof("Attach volume %s task info: %+v", volumeId, *task)
	taskState, err := GetTaskState(ctx, m.virtualCenter, task)
	if err != nil {
		klog.Errorf("Attach volume %s task failed with err: %+v", volumeId, err)
		return "", err
	} else if taskState != "FINISHED" {
		errMsg := fmt.Sprintf("Attach volume %s task state %s", volumeId, taskState)
		klog.Errorf(errMsg)
		return "", errors.New(errMsg)
	}

	err = vm.Renew(false)
	if err != nil {
		klog.Errorf("Get VM %v info failed with err: %+v", vm, err)
		return "", err
	}

	for _, diskInfo := range vm.VirtualMachine.Disks {
		if diskInfo.Volume.ID == volumeId {
			klog.V(5).Infof("Attach volume %s task finished, disk label %s", volumeId, diskInfo.Label)
			return diskInfo.Volume.ScsiID, nil
		}
	}

	errMsg := fmt.Sprintf("Attach volume %s task failed, volume not found.", volumeId)
	klog.Errorf(errMsg)
	return "", errors.New(errMsg)
}

// DetachVolume detaches a volume from the virtual machine given the spec.
func (m *volumeManager) DetachVolume(vm *VirtualMachine, volumeId string) error {
	err := validateManager(m)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = m.virtualCenter.Connect(ctx)
	if err != nil {
		klog.Errorf("Virtual Center Connect failed with err: %+v", err)
		return err
	}

	found := false
	vmInfo := *vm.VirtualMachine
	for i, disk := range vmInfo.Disks {
		if disk.ID == volumeId {
			found = true
			vmInfo.Disks = append(vmInfo.Disks[:i], vmInfo.Disks[i+1:]...)
			break
		}
	}
	if !found {
		errMsg := fmt.Sprintf("Volume %s not found for vm %v", volumeId, vm)
		klog.Errorf(errMsg)
		return errors.New(errMsg)
	}

	vmInfo.VncPasswd = "00000000"
	klog.V(4).Infof("Detaching volume %s from VM %v", volumeId, vm)

	vmService := icsvm.NewVirtualMachineService(m.virtualCenter.Client)
	task, err := vmService.SetVM(ctx, vmInfo)
	if err != nil {
		klog.Errorf("Failed to detach volume %s from VM %v with err: %+v", volumeId, vm, err)
		return err
	}

	klog.V(5).Infof("Detach volume %s task info: %+v", volumeId, *task)
	taskState, err := GetTaskState(ctx, m.virtualCenter, task)
	if err != nil {
		klog.Errorf("Detach volume %s task failed with err: %+v", volumeId, err)
		return err
	} else if taskState != "FINISHED" {
		errMsg := fmt.Sprintf("Detach volume %s task state %s", volumeId, taskState)
		klog.Errorf(errMsg)
		return errors.New(errMsg)
	}

	klog.V(5).Infof("Detach volume %s task finished", volumeId)
	return nil
}
