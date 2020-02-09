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
	"fmt"
	//"strings"
	//icsgo "github.com/inspur-ics/ics-go-sdk"
	"github.com/inspur-ics/ics-go-sdk/client"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	icsdc "github.com/inspur-ics/ics-go-sdk/datacenter"
	"k8s.io/klog"
	//"strconv"
)

// DatastoreInfoProperty refers to the property name info for the Datastore
const DatastoreInfoProperty = "info"

// Datacenter holds virtual center information along with the Datacenter.
type Datacenter struct {
	ID string
	*types.Datacenter
	Client *client.Client
	// VirtualCenterHost represents the virtual center host ip address.
	VirtualCenterHost string
}

func (dc *Datacenter) String() string {
	return fmt.Sprintf("Datacenter [ID: %s Name: %s VCenter: %s]",
		dc.ID, dc.Datacenter.Name, dc.VirtualCenterHost)
}

func (dc *Datacenter) GetVirtualMachineByNameOrUUID(ctx context.Context, name string, uuid string, instanceUUID bool) (*VirtualMachine, error) {
	vc, err := GetVirtualCenterManager().GetVirtualCenter(dc.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get VC for datacenter %v with err: %v", dc, err)
		return nil, err
	}
	if err := vc.Connect(ctx); err != nil {
		return nil, err
	}

	dcService := icsdc.NewDatacenterService(vc.Client)
	vmList, err := dcService.GetDatacenterVMList(ctx, dc.Datacenter.ID)
	if err != nil {
		klog.Errorf("get vm list of datacenter %s failed.", dc.Datacenter.Name)
		return nil, err
	}
	for _, vmItem := range vmList {
		if vmItem.UUID == uuid || vmItem.Name == name || vmItem.Description == name {
			vm := VirtualMachine{
				VirtualCenterHost: dc.VirtualCenterHost,
				UUID:              vmItem.UUID,
				VirtualMachine:    vmItem,
				Datacenter:        dc,
			}
			klog.V(4).Infof("find vm %s in datacenter %s successfully.", name, dc.Datacenter.Name)
			return &vm, nil
		}
	}
	return nil, ErrVMNotFound
}

func asyncGetAllDatacenters(ctx context.Context, dcsChan chan<- *Datacenter, errChan chan<- error) {
	defer close(dcsChan)
	defer close(errChan)

	for _, vc := range GetVirtualCenterManager().GetAllVirtualCenters() {
		klog.V(5).Infof("VirtualCenter:%+v", vc)
		// If the context was canceled, we stop looking for more Datacenters.
		select {
		case <-ctx.Done():
			err := ctx.Err()
			klog.V(2).Infof("Context was done, returning with err: %v", err)
			errChan <- err
			return
		default:
		}

		if err := vc.Connect(ctx); err != nil {
			klog.Errorf("Failed connecting to VC %q with err: %v", vc.Config.Host, err)
			errChan <- err
			return
		}

		dcs, err := vc.GetDatacenters(ctx)
		if err != nil {
			klog.Errorf("Failed to fetch datacenters for vc %v with err: %v", vc.Config.Host, err)
			errChan <- err
			return
		}

		for _, dc := range dcs {
			// If the context was canceled, we don't return more Datacenters.
			select {
			case <-ctx.Done():
				err := ctx.Err()
				klog.V(2).Infof("Context was done, returning with err: %v", err)
				errChan <- err
				return
			default:
				klog.V(2).Infof("Publishing datacenter id: %s name: %s\n", dc.Datacenter.ID, dc.Datacenter.Name)
				dcsChan <- dc
			}
		}
	}
}

func AsyncGetAllDatacenters(ctx context.Context, buffSize int) (<-chan *Datacenter, <-chan error) {
	dcsChan := make(chan *Datacenter, buffSize)
	errChan := make(chan error, 1)
	go asyncGetAllDatacenters(ctx, dcsChan, errChan)
	return dcsChan, errChan
}
