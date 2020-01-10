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
	"fmt"

	"k8s.io/klog"
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
	//*object.VirtualMachine
	// Datacenter represents the datacenter to which the virtual machine belongs.
	Datacenter *Datacenter
}

func (vm *VirtualMachine) String() string {
	return fmt.Sprintf("%v [VirtualCenterHost: %v, UUID: %v, Datacenter: %v]",
		vm.VirtualMachine, vm.VirtualCenterHost, vm.UUID, vm.Datacenter)
}
