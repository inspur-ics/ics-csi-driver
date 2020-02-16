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
	icsvol "github.com/inspur-ics/ics-go-sdk/volume"
	"sync"
	//"github.com/davecgh/go-spew/spew"
	"k8s.io/klog"
)

// VolumeManager provides functionality to manage volumes.
type VolumeManager interface {
	// CreateVolume creates a new volume given its spec.
	CreateVolume(req types.VolumeReq) (string, error)
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
