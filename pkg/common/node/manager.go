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

package node

import (
	"errors"
	"sync"

	k8s "ics-csi-driver/pkg/common/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"ics-csi-driver/pkg/common/ics"
)

var (
	// ErrNodeNotFound is returned when a node isn't found.
	ErrNodeNotFound = errors.New("node wasn't found")
	// ErrEmptyProviderID is returned when it is observed that provider id is not set on the kubernetes cluster
	ErrEmptyProviderID = errors.New("node with empty providerId present in the cluster")
)

// Manager provides functionality to manage nodes.
type Manager interface {
	// SetKubernetesClient sets kubernetes client for node manager
	SetKubernetesClient(clientset.Interface)
	// RegisterNode registers a node given its UUID, name.
	RegisterNode(nodeUUID string, nodeName string) error
	// DiscoverNode discovers a registered node given its UUID. This method
	// scans all virtual centers registered on the VirtualCenterManager for a
	// virtual machine with the given UUID.
	DiscoverNode(nodeUUID string, nodeName string) error
	//GetNodeUUID return UUID for a node given its nodeName
	GetNodeUUID(nodeName string) (string, error)
	// GetNode refreshes and returns the VirtualMachine for a registered node
	// given its UUID.
	GetNode(nodeUUID string, nodeName string) (*ics.VirtualMachine, error)
	// GetNodeByName refreshes and returns the VirtualMachine for a registered node
	// given its name.
	GetNodeByName(nodeName string) (*ics.VirtualMachine, error)
	// GetAllNodes refreshes and returns VirtualMachine for all registered
	// nodes. If nodes are added or removed concurrently, they may or may not be
	// reflected in the result of a call to this method.
	GetAllNodes() ([]*ics.VirtualMachine, error)
	// UnregisterNode unregisters a registered node given its name.
	UnregisterNode(nodeName string) error
}

// Metadata represents node metadata.
type Metadata interface{}

var (
	// managerInstance is a Manager singleton.
	managerInstance *nodeManager
	// onceForManager is used for initializing the Manager singleton.
	onceForManager sync.Once
)

// GetManager returns the Manager singleton.
func GetManager() Manager {
	onceForManager.Do(func() {
		klog.V(1).Info("Initializing node.nodeManager...")
		managerInstance = &nodeManager{
			nodeVMs: sync.Map{},
		}
		klog.V(1).Info("node.nodeManager initialized")
	})
	return managerInstance
}

// nodeManager holds node information and provides functionality around it.
type nodeManager struct {
	// nodeVMs maps node UUIDs to VirtualMachine objects.
	nodeVMs sync.Map
	// node name to node UUI map.
	nodeNameToUUID sync.Map
	// k8s client
	k8sClient clientset.Interface
}

// SetKubernetesClient sets specified kubernetes client to nodeManager.k8sClient
func (m *nodeManager) SetKubernetesClient(client clientset.Interface) {
	m.k8sClient = client
}

// RegisterNode registers a node with node manager using its UUID, name.
func (m *nodeManager) RegisterNode(nodeUUID string, nodeName string) error {
	m.nodeNameToUUID.Store(nodeName, nodeUUID)
	klog.V(2).Infof("Successfully registered node: %q with nodeUUID %q", nodeName, nodeUUID)
	err := m.DiscoverNode(nodeUUID, nodeName)
	if err != nil {
		klog.Errorf("Failed to discover VM with uuid: %q for node: %q", nodeUUID, nodeName)
		return err
	}
	klog.V(2).Infof("Successfully discovered node: %q with nodeUUID %q", nodeName, nodeUUID)
	return nil
}

// DiscoverNode discovers a registered node given its UUID from vCenter.
// If node is not found in the vCenter for the given UUID, for ErrVMNotFound is returned to the caller
func (m *nodeManager) DiscoverNode(nodeUUID string, nodeName string) error {
	vm, err := ics.GetVirtualMachineByUUID(nodeName, nodeUUID, false)
	if err != nil {
		klog.Errorf("Couldn't find VM instance with nodeUUID %s, failed to discover with err: %v", nodeUUID, err)
		return err
	}
	m.nodeVMs.Store(nodeUUID, vm)
	klog.V(2).Infof("Successfully discovered node with nodeUUID %s in vm %v", nodeUUID, vm)
	return nil
}

func (m *nodeManager) GetNodeUUID(nodeName string) (string, error) {
	nodeUUID, found := m.nodeNameToUUID.Load(nodeName)
	if !found {
		klog.Errorf("Node not found with nodeName %s", nodeName)
		return "", ErrNodeNotFound
	}
	if nodeUUID != nil && nodeUUID.(string) != "" {
		return nodeUUID.(string), nil
	}

	klog.V(2).Infof("Empty nodeUUID observed in cache for the node: %q", nodeName)
	k8snodeUUID, err := k8s.GetNodeVMUUID(m.k8sClient, nodeName)
	if err != nil {
		klog.Errorf("Failed to get providerId from node: %q. Err: %v", nodeName, err)
		return "", err
	}
	m.nodeNameToUUID.Store(nodeName, k8snodeUUID)
	return k8snodeUUID, nil
}

// GetNodeByName refreshes and returns the VirtualMachine for a registered node
// given its name.
func (m *nodeManager) GetNodeByName(nodeName string) (*ics.VirtualMachine, error) {
	nodeUUID, found := m.nodeNameToUUID.Load(nodeName)
	if !found {
		klog.Errorf("Node not found with nodeName %s", nodeName)
		return nil, ErrNodeNotFound
	}
	if nodeUUID != nil && nodeUUID.(string) != "" {
		return m.GetNode(nodeUUID.(string), nodeName)
	}
	klog.V(2).Infof("Empty nodeUUID observed in cache for the node: %q", nodeName)
	k8snodeUUID, err := k8s.GetNodeVMUUID(m.k8sClient, nodeName)
	if err != nil {
		klog.Errorf("Failed to get providerId from node: %q. Err: %v", nodeName, err)
		return nil, err
	}
	m.nodeNameToUUID.Store(nodeName, k8snodeUUID)
	return m.GetNode(k8snodeUUID, nodeName)
}

// GetNode refreshes and returns the VirtualMachine for a registered node
// given its UUID
func (m *nodeManager) GetNode(nodeUUID string, nodeName string) (*ics.VirtualMachine, error) {
	vmInf, discovered := m.nodeVMs.Load(nodeUUID)
	if !discovered {
		klog.V(2).Infof("Node hasn't been discovered yet with nodeUUID %s", nodeUUID)

		if err := m.DiscoverNode(nodeUUID, nodeName); err != nil {
			klog.Errorf("Failed to discover node with nodeUUID %s with err: %v", nodeUUID, err)
			return nil, err
		}

		vmInf, _ = m.nodeVMs.Load(nodeUUID)
		klog.V(2).Infof("Node was successfully discovered with nodeUUID %s in vm %v", nodeUUID, vmInf)

		return vmInf.(*ics.VirtualMachine), nil
	}

	vm := vmInf.(*ics.VirtualMachine)
	klog.V(1).Infof("Renewing virtual machine %v with nodeUUID %s", vm, nodeUUID)

	if err := vm.Renew(true); err != nil {
		klog.Errorf("Failed to renew VM %v with nodeUUID %s with err: %v", vm, nodeUUID, err)
		return nil, err
	}

	klog.V(1).Infof("VM %v was successfully renewed with nodeUUID %s", vm, nodeUUID)
	return vm, nil
}

// GetAllNodes refreshes and returns VirtualMachine for all registered nodes.
func (m *nodeManager) GetAllNodes() ([]*ics.VirtualMachine, error) {
	var vms []*ics.VirtualMachine
	var err error
	reconnectedHosts := make(map[string]bool)

	m.nodeNameToUUID.Range(func(nodeName, nodeUUID interface{}) bool {
		if nodeName != nil && nodeUUID != nil && nodeUUID.(string) == "" {
			klog.V(2).Infof("Empty node UUID observed for the node: %q", nodeName)
			k8snodeUUID, err := k8s.GetNodeVMUUID(m.k8sClient, nodeName.(string))
			if err != nil {
				klog.Errorf("Failed to get providerId from node: %q. Err: %v", nodeName, err)
				return true
			}
			if k8snodeUUID == "" {
				klog.Errorf("Node: %q with empty providerId found in the cluster. aborting get all nodes", nodeName)
				return true
			}
			m.nodeNameToUUID.Store(nodeName, k8snodeUUID)
			return false
		}
		return true
	})

	if err != nil {
		return nil, err
	}
	m.nodeVMs.Range(func(nodeUUIDInf, vmInf interface{}) bool {
		// If an entry was concurrently deleted from vm, Range could
		// possibly return a nil value for that key.
		// See https://golang.org/pkg/sync/#Map.Range for more info.
		if vmInf == nil {
			klog.Warningf("VM instance was nil, ignoring with nodeUUID %v", nodeUUIDInf)
			return true
		}

		nodeUUID := nodeUUIDInf.(string)
		vm := vmInf.(*ics.VirtualMachine)

		if reconnectedHosts[vm.VirtualCenterHost] {
			klog.V(3).Infof("Renewing VM %v, no new connection needed: nodeUUID %s", vm, nodeUUID)
			err = vm.Renew(false)
		} else {
			klog.V(3).Infof("Renewing VM %v with new connection: nodeUUID %s", vm, nodeUUID)
			err = vm.Renew(true)
			reconnectedHosts[vm.VirtualCenterHost] = true
		}

		if err != nil {
			klog.Errorf("Failed to renew VM %v with nodeUUID %s, aborting get all nodes", vm, nodeUUID)
			return false
		}

		klog.V(3).Infof("Updated VM %v for node with nodeUUID %s", vm, nodeUUID)
		vms = append(vms, vm)
		return true
	})

	if err != nil {
		return nil, err
	}
	return vms, nil
}

// UnregisterNode unregisters a registered node given its name.
func (m *nodeManager) UnregisterNode(nodeName string) error {
	nodeUUID, found := m.nodeNameToUUID.Load(nodeName)
	if !found {
		klog.Errorf("Node wasn't found, failed to unregister node: %q", nodeName)
		return ErrNodeNotFound
	}
	m.nodeNameToUUID.Delete(nodeName)
	m.nodeVMs.Delete(nodeUUID)
	klog.V(2).Infof("Successfully unregistered node with nodeName %s", nodeName)
	return nil
}
