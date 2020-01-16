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

package icsphere

import (
	"context"
	"errors"
	"fmt"
	"github.com/inspur-ics/ics-go-sdk/client/types"
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

func (vm *VirtualMachine) String() string {
	return fmt.Sprintf("VM [Name: %v, Desc: %v, UUID: %v, Datacenter: %v]",
		vm.VirtualMachine.Name, vm.VirtualMachine.Description, vm.UUID, vm.Datacenter)
}

// GetAllAccessibleDatastores gets the list of accessible Datastores for the given Virtual Machine
func (vm *VirtualMachine) GetAllAccessibleDatastores(ctx context.Context) ([]*DatastoreInfo, error) {
	var dsObjList []*DatastoreInfo
	return dsObjList, nil
}

// renew renews the virtual machine and datacenter objects given its virtual center.
func (vm *VirtualMachine) renew(vc *VirtualCenter) {
	//vm.VirtualMachine = object.NewVirtualMachine(vc.Client.Client, vm.VirtualMachine.Reference())
	//vm.Datacenter.Datacenter = object.NewDatacenter(vc.Client.Client, vm.Datacenter.Reference())
}

// Renew renews the virtual machine and datacenter information. If reconnect is
// set to true, the virtual center connection is also renewed.
func (vm *VirtualMachine) Renew(reconnect bool) error {
	vc, err := GetVirtualCenterManager().GetVirtualCenter(vm.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get VC while renewing VM %v with err: %v", vm, err)
		return err
	}

	/*
		if reconnect {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := vc.Connect(ctx); err != nil {
				klog.Errorf("Failed reconnecting to VC %q while renewing VM %v with err: %v", vc.Config.Host, vm, err)
				return err
			}
		}
	*/
	vm.renew(vc)

	return nil
}

const (
	// poolSize is the number of goroutines to run while trying to find a
	// virtual machine.
	poolSize = 4
	// dcBufferSize is the buffer size for the channel that is used to
	// asynchronously receive *Datacenter instances.
	dcBufferSize = poolSize * 10
)

// GetVirtualMachineByNameOrUUID returns virtual machine given its UUID in entire VC.
// If instanceUuid is set to true, then UUID is an instance UUID.
// In this case, this function searches for virtual machines whose instance UUID matches the given uuid.
// If instanceUuid is set to false, then UUID is BIOS UUID.
// In this case, this function searches for virtual machines whose BIOS UUID matches the given uuid.
func GetVirtualMachineByNameOrUUID(name string, uuid string, instanceUUID bool) (*VirtualMachine, error) {
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
					if uuid != "" {
						vm, err = dc.GetVirtualMachineByUUID(context.Background(), uuid, instanceUUID)
					} else {
						vm, err = dc.GetVirtualMachineByName(context.Background(), name)
					}

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

// IsInZoneRegion checks if virtual machine belongs to specified zone and region
// This function returns true if virtual machine belongs to specified zone/region, else returns false.
func (vm *VirtualMachine) IsInZoneRegion(ctx context.Context, zoneCategoryName string, regionCategoryName string, zoneValue string, regionValue string) (bool, error) {
	return false, nil
}
