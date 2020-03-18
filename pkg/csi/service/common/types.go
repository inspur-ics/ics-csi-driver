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

package common

import (
	"github.com/container-storage-interface/spec/lib/go/csi"

	"ics-csi-driver/pkg/common/config"
	"ics-csi-driver/pkg/common/ics"
)

var (
	// VolumeCaps represents how the volume could be accessed.
	// It is SINGLE_NODE_WRITER since ICS CNS Block volume could only be
	// attached to a single node at any given time.
	VolumeCaps = []csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
	}
)

// Manager type comprises VirtualCenterConfig, CnsConfig, VolumeManager and VirtualCenterManager
type Manager struct {
	VcenterConfig  *ics.VirtualCenterConfig
	CnsConfig      *config.Config
	VolumeManager  ics.VolumeManager
	VcenterManager ics.VirtualCenterManager
}

// CreateVolumeSpec is the Volume Spec used by CSI driver
type CreateVolumeSpec struct {
	Name        string
	DatastoreID string
	CapacityGB  int64
}
