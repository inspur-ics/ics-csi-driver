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
	//"k8s.io/klog"
)

// DatastoreInfoProperty refers to the property name info for the Datastore
const DatastoreInfoProperty = "info"

// Datacenter holds virtual center information along with the Datacenter.
type Datacenter struct {
	// Datacenter represents the govmomi Datacenter.
	//*object.Datacenter
	// VirtualCenterHost represents the virtual center host address.
	VirtualCenterHost string
}

func (dc *Datacenter) String() string {
	return fmt.Sprintf("Datacenter [VirtualCenterHost: %v]", dc.VirtualCenterHost)
}

func (dc *Datacenter) GetVirtualMachineByUUID(ctx context.Context, uuid string, instanceUUID bool) (*VirtualMachine, error) {
	return nil, nil
}

func AsyncGetAllDatacenters(ctx context.Context, buffSize int) (<-chan *Datacenter, <-chan error) {
	dcsChan := make(chan *Datacenter, buffSize)
	errChan := make(chan error, 1)
	//go asyncGetAllDatacenters(ctx, dcsChan, errChan)
	return dcsChan, errChan
}
