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

package types

type KeyValue struct {
	Key   string
	Value string
}

type BaseCnsEntityMetadata interface {
	GetCnsEntityMetadata() *CnsEntityMetadata
}

type CnsEntityMetadata struct {
	EntityName string
	Labels     []KeyValue
	Delete     bool
}

func (b *CnsEntityMetadata) GetCnsEntityMetadata() *CnsEntityMetadata { return b }

type CnsKubernetesEntityMetadata struct {
	CnsEntityMetadata

	EntityType string
	Namespace  string
}

type CnsContainerCluster struct {
	ClusterType string
	ClusterId   string
	UserName    string
}

type CnsVolumeMetadata struct {
	ContainerCluster CnsContainerCluster
	EntityMetadata   []BaseCnsEntityMetadata
}

type BaseCnsBackingObjectDetails interface {
	GetCnsBackingObjectDetails() *CnsBackingObjectDetails
}

type CnsBackingObjectDetails struct {
	CapacityInMb int64
}

func (b *CnsBackingObjectDetails) GetCnsBackingObjectDetails() *CnsBackingObjectDetails { return b }

type CnsBlockBackingDetails struct {
	CnsBackingObjectDetails
	BackingDiskId string
}

type CnsVolumeCreateSpec struct {
	Name                 string
	VolumeType           string
	DatastoreID          string
	Metadata             CnsVolumeMetadata
	BackingObjectDetails BaseCnsBackingObjectDetails
}

type CnsVolumeMetadataUpdateSpec struct {
	VolumeId string
	Metadata CnsVolumeMetadata
}

type CnsVolume struct {
	VolumeId             string
	Name                 string
	VolumeType           string
	DatastoreUrl         string
	Metadata             CnsVolumeMetadata
	BackingObjectDetails CnsBackingObjectDetails
}
