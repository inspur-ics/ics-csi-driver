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
	icshost "github.com/inspur-ics/ics-go-sdk/host"
	icsvm "github.com/inspur-ics/ics-go-sdk/vm"
	"k8s.io/klog"
	"sync"
)

// ErrVMNotFound is returned when a virtual machine isn't found.
var ErrVMNotFound = errors.New("virtual machine wasn't found")

// VirtualMachine holds details of a virtual machine instance.
type VirtualMachine struct {
	// VirtualCenterHost represents the virtual machine's vCenter host.
	VirtualCenterHost string
	// UUID represents the virtual machine's UUID.
	UUID string
	// VirtualMachine represents the virtual machine.
	*types.VirtualMachine
	// Datacenter represents the datacenter to which the virtual machine belongs.
	Datacenter *Datacenter
}

type IcsObject struct {
	ID   string
	Name string
	Type string
}

func (vm *VirtualMachine) String() string {
	return fmt.Sprintf("[Name: %v, Desc: %v, UUID: %v, Datacenter: %v]",
		vm.VirtualMachine.Name, vm.VirtualMachine.Description, vm.UUID, vm.Datacenter)
}

// GetHostSystem returns the host which the virtual machine belongs to
func (vm *VirtualMachine) GetHostSystem(ctx context.Context) (*Host, error) {
	vc, err := GetVirtualCenterManager().GetVirtualCenter(vm.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get VC for vm %v with err: %v", vm, err)
		return nil, err
	}
	if err := vc.Connect(ctx); err != nil {
		return nil, err
	}

	if err := vm.renew(ctx, vc); err != nil {
		return nil, err
	}

	hostService := icshost.NewHostService(vc.Client)
	hostInfo, err := hostService.GetHost(ctx, vm.VirtualMachine.HostID)
	if err != nil {
		klog.Errorf("Failed to get host %s info for vm %v with err: %v", vm.VirtualMachine.HostName, vm, err)
		return nil, err
	}

	host := &Host{
		Host:              hostInfo,
		VirtualCenterHost: vm.VirtualCenterHost,
	}
	return host, nil
}

// GetAllAccessibleDatastores gets the list of accessible Datastores for the given Virtual Machine
func (vm *VirtualMachine) GetAllAccessibleDatastores(ctx context.Context) ([]*DatastoreInfo, error) {
	host, err := vm.GetHostSystem(ctx)
	if err != nil {
		klog.Errorf("Failed to get host info for vm %v with err:%v", vm, err)
		return nil, err
	}

	return host.GetAllAccessibleDatastores(ctx)
}

// renew renews the virtual machine and datacenter objects given its virtual center.
func (vm *VirtualMachine) renew(ctx context.Context, vc *VirtualCenter) error {
	vmService := icsvm.NewVirtualMachineService(vc.Client)
	vminfo, err := vmService.GetVM(ctx, vm.VirtualMachine.ID)
	if err != nil {
		klog.Errorf("Failed to renew vm %v info with err: %v", vm, err)
		return err
	}
	vm.VirtualMachine = vminfo
	return vm.Datacenter.Renew(false)
}

// Renew renews the virtual machine and datacenter information. If reconnect is
// set to true, the virtual center connection is also renewed.
func (vm *VirtualMachine) Renew(reconnect bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vc, err := GetVirtualCenterManager().GetVirtualCenter(vm.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get VC while renewing VM %v with err: %v", vm, err)
		return err
	}

	if reconnect {
		if err := vc.Connect(ctx); err != nil {
			klog.Errorf("Failed reconnecting to VC %q while renewing VM %v with err: %v", vc.Config.Host, vm, err)
			return err
		}
	}

	return vm.renew(ctx, vc)
}

const (
	// poolSize is the number of goroutines to run while trying to find a
	// virtual machine.
	poolSize = 4
	// dcBufferSize is the buffer size for the channel that is used to
	// asynchronously receive *Datacenter instances.
	dcBufferSize = poolSize * 10
)

// GetVirtualMachineByUUID returns virtual machine given its UUID in entire VC.
// If instanceUuid is set to true, then UUID is an instance UUID.
// In this case, this function searches for virtual machines whose instance UUID matches the given uuid.
// If instanceUuid is set to false, then UUID is BIOS UUID.
// In this case, this function searches for virtual machines whose BIOS UUID matches the given uuid.
func GetVirtualMachineByUUID(name string, uuid string, instanceUUID bool) (*VirtualMachine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	klog.V(2).Infof("Initiating asynchronous datacenter listing with name:%s uuid %s", name, uuid)
	dcsChan, errChan := AsyncGetAllDatacenters(ctx, dcBufferSize)

	var wg sync.WaitGroup
	var vm, nodeVM *VirtualMachine
	var err, poolErr error
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case err, ok := <-errChan:
					if !ok {
						// Async function finished.
						klog.V(2).Infof("AsyncGetAllDatacenters finished with name %s uuid %s", name, uuid)
						return
					} else if err == context.Canceled {
						// Canceled by another instance of this goroutine.
						klog.V(2).Infof("AsyncGetAllDatacenters ctx was canceled with name %s uuid %s", name, uuid)
						return
					} else {
						// Some error occurred.
						klog.Errorf("AsyncGetAllDatacenters with name %s uuid %s sent an error: %v", name, uuid, err)
						poolErr = err
						return
					}

				case dc, ok := <-dcsChan:
					if !ok {
						// Async function finished.
						klog.V(2).Infof("AsyncGetAllDatacenters finished with name %s uuid %s", name, uuid)
						return
					}

					// Found some Datacenter object.
					klog.V(2).Infof("AsyncGetAllDatacenters with name %s uuid %s sent a dc %v", name, uuid, dc)
					vm, err = dc.GetVirtualMachineByUUID(context.Background(), name, uuid, instanceUUID)

					if err != nil {
						if err == ErrVMNotFound {
							// Didn't find VM on this DC, so, continue searching on other DCs.
							klog.V(2).Infof("Couldn't find VM given name %s uuid %s on DC %v with err: %v, continuing search", name, uuid, dc, err)
							continue
						} else {
							// Some serious error occurred, so stop the async function.
							klog.Errorf("Failed finding VM given name %s uuid %s on DC %v with err: %v, canceling context", name, uuid, dc, err)
							cancel()
							poolErr = err
							return
						}
					} else {
						// Virtual machine was found, so stop the async function.
						klog.V(2).Infof("Found VM %v given name %s uuid %s on DC %v, canceling context", vm, name, uuid, dc)
						nodeVM = vm
						cancel()
						return
					}
				}
			}
		}()
	}
	wg.Wait()

	if nodeVM != nil {
		klog.V(2).Infof("Returning VM %v for Name %s UUID %s", nodeVM, name, uuid)
		return nodeVM, nil
	} else if poolErr != nil {
		klog.Errorf("Returning err: %v for Name %s UUID %s", poolErr, name, uuid)
		return nil, poolErr
	} else {
		klog.Errorf("Returning VM not found err for Name %s UUID %s", name, uuid)
		return nil, ErrVMNotFound
	}
}

func (vm *VirtualMachine) GetAncestors(ctx context.Context) ([]IcsObject, error) {
	host, err := vm.GetHostSystem(ctx)
	if err != nil {
		klog.Errorf("Failed to get host info for vm %s with err: %v", vm, err)
		return nil, err
	}

	var icsObjs []IcsObject
	hostObj := IcsObject{
		ID:   host.Host.ID,
		Name: host.Host.Name,
		Type: "HOST",
	}
	icsObjs = append(icsObjs, hostObj)

	if host.Host.ClusterID != "" {
		cluster := IcsObject{
			ID:   host.Host.ClusterID,
			Name: host.Host.ClusterName,
			Type: "CLUSTER",
		}
		icsObjs = append(icsObjs, cluster)
	}

	datacenter := IcsObject{
		ID:   host.Host.DataCenterID,
		Name: host.Host.DataCenterName,
		Type: "DATACENTER",
	}
	icsObjs = append(icsObjs, datacenter)

	return icsObjs, nil
}

// GetZoneRegion returns zone and region of the node vm
func (vm *VirtualMachine) GetZoneRegion(ctx context.Context, zoneCategoryName string, regionCategoryName string) (zone string, region string, err error) {
	icsObjs, err := vm.GetAncestors(ctx)
	if err != nil {
		klog.Errorf("GetAncestors failed for %v with err %v", vm, err)
		return "", "", err
	}
	klog.V(5).Infof("Vm's ancestors:%+v", icsObjs)

	vc, err := GetVirtualCenterManager().GetVirtualCenter(vm.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get iCenter %s with err: %v", vm.VirtualCenterHost, err)
		return "", "", err
	}

	zone, region = "", ""
	// search the hierarchy, example order: ["HOST", "CLUSTER", "DATACENTER"]
	for _, obj := range icsObjs {
		if obj.Type != "DATACENTER" && obj.ID != "" {
			tags, err := GetAttachedTags(ctx, vc, obj.Type, obj.ID)
			if err != nil {
				klog.Errorf("Get attached tags faild for %s %s with err %v", obj.Type, obj.Name, err)
				return "", "", err
			}
			klog.V(5).Infof("Get %s %s tags:%+v", obj.Type, obj.Name, tags)
			for _, tag := range tags {
				if tag.Description == regionCategoryName && region == "" {
					region = tag.Name
				} else if tag.Description == zoneCategoryName && zone == "" {
					zone = tag.Name
				}
				if zone != "" && region != "" {
					return zone, region, nil
				}
			}
		} else if obj.Type == "DATACENTER" && region == "" {
			region = vm.Datacenter.Datacenter.Description
		}
	}
	return zone, region, nil
}

// IsInZoneRegion checks if virtual machine belongs to specified zone and region
// This function returns true if virtual machine belongs to specified zone/region, else returns false.
func (vm *VirtualMachine) IsInZoneRegion(ctx context.Context, zoneCategoryName string, regionCategoryName string, zoneValue string, regionValue string) (bool, error) {
	klog.V(4).Infof("IsInZoneRegion: called with zoneCategoryName: %s, regionCategoryName: %s, zoneValue: %s, regionValue: %s", zoneCategoryName, regionCategoryName, zoneValue, regionValue)

	vmZone, vmRegion, err := vm.GetZoneRegion(ctx, zoneCategoryName, regionCategoryName)
	if err != nil {
		klog.Errorf("failed to get accessibleTopology for vm: %s, err: %v", vm.VirtualMachine.Name, err)
		return false, err
	}
	klog.V(4).Infof("VM [%s] belongs to zone [%s] and region [%s]", vm.VirtualMachine.Name, vmZone, vmRegion)

	if regionValue == "" && zoneValue != "" && vmZone == zoneValue {
		// region is not specified, if zone matches with look up zone value, return true
		klog.V(4).Infof("VM [%s] belongs to zone [%s]", vm.VirtualMachine.Name, zoneValue)
		return true, nil
	}
	if zoneValue == "" && regionValue != "" && vmRegion == regionValue {
		// zone is not specified, if region matches with look up region value, return true
		klog.V(4).Infof("VM [%s] belongs to region [%s]", vm.VirtualMachine.Name, regionValue)
		return true, nil
	}
	if vmZone != "" && vmRegion != "" && vmRegion == regionValue && vmZone == zoneValue {
		klog.V(4).Infof("VM [%s] belongs to zone [%s] and region [%s]", vm.VirtualMachine.Name, zoneValue, regionValue)
		return true, nil
	}
	return false, nil
}
