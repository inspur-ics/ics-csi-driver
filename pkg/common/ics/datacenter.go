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
	"fmt"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	icsdc "github.com/inspur-ics/ics-go-sdk/datacenter"
	"k8s.io/klog"
)

// DatastoreInfoProperty refers to the property name info for the Datastore
const DatastoreInfoProperty = "info"

// Datacenter holds virtual center information along with the Datacenter.
type Datacenter struct {
	ID string
	*types.Datacenter
	// VirtualCenterHost represents the virtual center host ip address.
	VirtualCenterHost string
}

func (dc *Datacenter) String() string {
	return fmt.Sprintf("[ID: %s Name: %s VCenter: %s]",
		dc.ID, dc.Datacenter.Name, dc.VirtualCenterHost)
}

// Renew renews the datacenter information. If reconnect is
// set to true, the virtual center connection is also renewed.
func (dc *Datacenter) Renew(reconnect bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vc, err := GetVirtualCenterManager().GetVirtualCenter(dc.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get VC while renewing datacenter %v with err: %v", dc, err)
		return err
	}

	if reconnect {
		if err := vc.Connect(ctx); err != nil {
			klog.Errorf("Failed reconnecting to VC %q while renewing datacenter %v with err: %v", vc.Config.Host, dc, err)
			return err
		}
	}

	dcService := icsdc.NewDatacenterService(vc.Client)
	dcinfo, err := dcService.GetDatacenter(ctx, dc.ID)
	if err != nil {
		klog.Errorf("Failed to renew datacenter %s info with err: %v", dc.Datacenter.Name, err)
		return err
	}
	dc.Datacenter = dcinfo
	return nil
}

func (dc *Datacenter) GetVirtualMachineByUUID(ctx context.Context, hostname string, uuid string, instanceUUID bool) (*VirtualMachine, error) {
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
		klog.Errorf("Get vm list of datacenter %s failed.", dc.Datacenter.Name)
		return nil, err
	}
	for _, vmItem := range vmList {
		if vmItem.UUID == uuid {
			vm := VirtualMachine{
				VirtualCenterHost: dc.VirtualCenterHost,
				UUID:              vmItem.UUID,
				VirtualMachine:    vmItem,
				Datacenter:        dc,
			}
			klog.V(4).Infof("Find vm %s %s in datacenter %s successfully.", uuid, hostname, dc.Datacenter.Name)
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
