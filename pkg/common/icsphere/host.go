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
	"github.com/inspur-ics/ics-go-sdk/client/types"
	icshost "github.com/inspur-ics/ics-go-sdk/host"
	"k8s.io/klog"
)

type Host struct {
	ID string
	// Host represents the ics host info.
	*types.Host
	// VirtualCenterHost represents the virtual center host address.
	VirtualCenterHost string
}

// Datastore holds Datastore and Datacenter information.
type Datastore struct {
	// Datastore represents the ics Datastore instance.
	*types.Datastore
	// Datacenter represents the datacenter on which the Datastore resides.
	Datacenter *Datacenter
}

// DatastoreInfo is a structure to store the Datastore and it's Info.
type DatastoreInfo struct {
	ID            string
	Type          string
	Name          string
	Capacity      float64
	AvailCapacity float64
	//*Datastore
}

func (di DatastoreInfo) String() string {
	return fmt.Sprintf("Datastore ID: %v, Type: %v, Name: %v, Capacity: %v, AvailCapacity: %v",
		di.ID, di.Type, di.Name, di.Capacity, di.AvailCapacity)
}

// GetAllAccessibleDatastores gets the list of accessible datastores for the given host
func (host *Host) GetAllAccessibleDatastores(ctx context.Context) ([]*DatastoreInfo, error) {
	vc, err := GetVirtualCenterManager().GetVirtualCenter(host.VirtualCenterHost)
	if err != nil {
		klog.Errorf("Failed to get VC for host %v with err: %v", host.ID, err)
		return nil, err
	}
	if err := vc.Connect(ctx); err != nil {
		return nil, err
	}

	hostService := icshost.NewHostService(vc.Client)
	datastoreList, err := hostService.GetHostAvailStorages(ctx, host.ID)
	if err != nil {
		klog.Errorf("get datastore list of host %s failed.", host.ID)
		return nil, err
	}
	var dsList []*DatastoreInfo
	for _, datastore := range datastoreList {
		dsList = append(dsList,
			&DatastoreInfo{
				ID:            datastore.ID,
				Type:          datastore.DataStoreType,
				Name:          datastore.Name,
				Capacity:      datastore.Capacity,
				AvailCapacity: datastore.AvailCapacity,
			})
	}
	return dsList, nil
}
