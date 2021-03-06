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

package syncer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	csictx "github.com/rexray/gocsi/context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	cnsconfig "ics-csi-driver/pkg/common/config"
	"ics-csi-driver/pkg/common/ics"
	cnstypes "ics-csi-driver/pkg/common/types"
	"ics-csi-driver/pkg/csi/service"
	//"ics-csi-driver/pkg/csi/service/common"
	k8s "ics-csi-driver/pkg/common/kubernetes"
)

// NewInformer returns uninitialized metadataSyncInformer
func NewInformer() *MetadataSyncInformer {
	return &MetadataSyncInformer{}
}

// getFullSyncIntervalInMin return the FullSyncInterval
// If enviroment variable FULL_SYNC_INTERVAL_MINUTES is set and valid,
// return the interval value read from enviroment variable
// otherwise, use the default value 30 minutes
func getFullSyncIntervalInMin() int {
	fullSyncIntervalInMin := defaultFullSyncIntervalInMin
	if v := os.Getenv(envFullSyncIntervalMinutes); v != "" {
		if value, err := strconv.Atoi(v); err == nil {
			if value <= 0 || value > defaultFullSyncIntervalInMin {
				msg := fmt.Sprintf("FullSync: FULL_SYNC_INTERVAL_MINUTES %s is not in valid range, will use the default interval", v)
				klog.Warningf(msg)
			} else {
				fullSyncIntervalInMin = value
				klog.V(2).Infof("FullSync: fullSync interval is set to %d minutes", fullSyncIntervalInMin)
			}
		} else {
			msg := fmt.Sprintf("FullSync: FULL_SYNC_INTERVAL_MINUTES %s is invalid, will use the default interval", v)
			klog.Warningf(msg)
		}
	}
	return fullSyncIntervalInMin
}

// Init initializes the Metadata Sync Informer
func (metadataSyncer *MetadataSyncInformer) Init() error {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfgPath := csictx.Getenv(ctx, cnsconfig.EnvCloudConfig)
	if cfgPath == "" {
		cfgPath = cnsconfig.DefaultCloudConfigPath
	}
	metadataSyncer.cfg, err = cnsconfig.GetCnsconfig(cfgPath)
	if err != nil {
		klog.Errorf("Failed to parse config. Err: %v", err)
		return err
	}

	metadataSyncer.vcconfig, err = ics.GetVirtualCenterConfig(metadataSyncer.cfg)
	if err != nil {
		klog.Errorf("Failed to get VirtualCenterConfig. err=%v", err)
		return err
	}

	// Initialize the virtual center manager
	metadataSyncer.virtualcentermanager = ics.GetVirtualCenterManager()

	// Register virtual center manager
	metadataSyncer.vcenter, err = metadataSyncer.virtualcentermanager.RegisterVirtualCenter(metadataSyncer.vcconfig)
	if err != nil {
		klog.Errorf("Failed to register VirtualCenter . err=%v", err)
		return err
	}

	// Connect to VC
	err = metadataSyncer.vcenter.Connect(ctx)
	if err != nil {
		klog.Errorf("Failed to connect to VirtualCenter host: %q. err=%v", metadataSyncer.vcconfig.Host, err)
		return err
	}
	// Create the kubernetes client from config
	k8sclient, err := k8s.NewClient()
	if err != nil {
		klog.Errorf("Creating Kubernetes client failed. Err: %v", err)
		return err
	}

	// Initialize cnsDeletionMap used by Full Sync
	cnsDeletionMap = make(map[string]bool)
	// Initialize cnsCreationMap used by Full Sync
	cnsCreationMap = make(map[string]bool)

	ticker := time.NewTicker(time.Duration(getFullSyncIntervalInMin()) * time.Minute)
	// Trigger full sync
	go func() {
		for range ticker.C {
			klog.V(2).Infof("fullSync is triggered")
			triggerFullSync(k8sclient, metadataSyncer)
		}
	}()

	stopFullSync := make(chan bool, 1)

	// Set up kubernetes resource listeners for metadata syncer
	metadataSyncer.k8sInformerManager = k8s.NewInformer(k8sclient)
	metadataSyncer.k8sInformerManager.AddPVCListener(
		nil, // Add
		func(oldObj interface{}, newObj interface{}) { // Update
			pvcUpdated(oldObj, newObj, metadataSyncer)
		},
		func(obj interface{}) { // Delete
			pvcDeleted(obj, metadataSyncer)
		})
	metadataSyncer.k8sInformerManager.AddPVListener(
		nil, // Add
		func(oldObj interface{}, newObj interface{}) { // Update
			pvUpdated(oldObj, newObj, metadataSyncer)
		},
		func(obj interface{}) { // Delete
			pvDeleted(obj, metadataSyncer)
		})
	metadataSyncer.k8sInformerManager.AddPodListener(
		nil, // Add
		func(oldObj interface{}, newObj interface{}) { // Update
			podUpdated(oldObj, newObj, metadataSyncer)
		},
		func(obj interface{}) { // Delete
			podDeleted(obj, metadataSyncer)
		})
	metadataSyncer.pvLister = metadataSyncer.k8sInformerManager.GetPVLister()
	metadataSyncer.pvcLister = metadataSyncer.k8sInformerManager.GetPVCLister()
	klog.V(2).Infof("Initialized metadata syncer")
	stopCh := metadataSyncer.k8sInformerManager.Listen()
	<-(stopCh)
	<-(stopFullSync)
	return nil
}

// getCnsKubernetesEntityMetaData creates a CnsKubernetesEntityMetadata object from given parameters
func getCnsKubernetesEntityMetaData(entityName string, labels map[string]string, deleteFlag bool, entityType string, namespace string) *cnstypes.CnsKubernetesEntityMetadata {
	// Create new metadata spec
	var newLabels []cnstypes.KeyValue
	for labelKey, labelVal := range labels {
		newLabels = append(newLabels, cnstypes.KeyValue{
			Key:   labelKey,
			Value: labelVal,
		})
	}

	entityMetadata := &cnstypes.CnsKubernetesEntityMetadata{}
	entityMetadata.EntityName = entityName
	entityMetadata.Delete = deleteFlag
	if labels != nil {
		entityMetadata.Labels = newLabels
	}
	entityMetadata.EntityType = entityType
	entityMetadata.Namespace = namespace
	return entityMetadata
}

// getContainerCluster creates a ContainerCluster object from given parameters
func getContainerCluster(clusterId string, userName string) cnstypes.CnsContainerCluster {
	return cnstypes.CnsContainerCluster{
		ClusterType: "KUBERNETES",
		ClusterId:   clusterId,
		UserName:    userName,
	}
}

// getCnsVolumeMetadataUpdateSpec creates a CnsVolumeMetadataUpdateSpec object from given parameters
func getCnsVolumeMetadataUpdateSpec(volumeId string, metadataList []cnstypes.BaseCnsEntityMetadata, metadataSyncer *MetadataSyncInformer) *cnstypes.CnsVolumeMetadataUpdateSpec {
	return &cnstypes.CnsVolumeMetadataUpdateSpec{
		VolumeId: volumeId,
		Metadata: cnstypes.CnsVolumeMetadata{
			ContainerCluster: getContainerCluster(metadataSyncer.cfg.Global.ClusterID, metadataSyncer.cfg.VirtualCenter[metadataSyncer.vcenter.Config.Host].User),
			EntityMetadata:   metadataList,
		},
	}
}

// pvcUpdated updates persistent volume claim metadata on VC when pvc labels on K8S cluster have been updated
func pvcUpdated(oldObj, newObj interface{}, metadataSyncer *MetadataSyncInformer) {
	// Get old and new pvc objects
	oldPvc, ok := oldObj.(*v1.PersistentVolumeClaim)
	if oldPvc == nil || !ok {
		return
	}
	newPvc, ok := newObj.(*v1.PersistentVolumeClaim)
	if newPvc == nil || !ok {
		return
	}

	if newPvc.Status.Phase != v1.ClaimBound {
		klog.V(3).Infof("PVCUpdated: New PVC %s in %s phase", newPvc.Name, newPvc.Status.Phase)
		return
	}

	// Get pv object attached to pvc
	pv, err := metadataSyncer.pvLister.Get(newPvc.Spec.VolumeName)
	if pv == nil || err != nil {
		klog.Errorf("PVCUpdated: Error getting Persistent Volume for pvc %s in namespace %s with err: %v", newPvc.Name, newPvc.Namespace, err)
		return
	}

	// Verify if pv is ics csi volume
	if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != service.Name {
		klog.V(3).Infof("PVCUpdated: PV %s Not a IncloudSphere CSI Volume", pv.Name)
		return
	}

	// Verify is old and new labels are not equal
	if oldPvc.Status.Phase == v1.ClaimBound && reflect.DeepEqual(newPvc.Labels, oldPvc.Labels) {
		klog.V(3).Infof("PVCUpdated: PVC %s (phase: %s) Old PVC and New PVC labels equal", oldPvc.Name, oldPvc.Status.Phase)
		return
	}

	// Create updateSpec
	var metadataList []cnstypes.BaseCnsEntityMetadata
	pvcMetadata := getCnsKubernetesEntityMetaData(newPvc.Name, newPvc.Labels, false, string(cnstypes.CnsKubernetesEntityTypePVC), newPvc.Namespace)
	metadataList = append(metadataList, cnstypes.BaseCnsEntityMetadata(pvcMetadata))

	updateSpec := getCnsVolumeMetadataUpdateSpec(pv.Spec.CSI.VolumeHandle, metadataList, metadataSyncer)
	klog.V(4).Infof("PVCUpdated: Calling UpdateVolumeMetadata with updateSpec: %+v", spew.Sdump(updateSpec))
	//if err := ics.GetVolumeManager(metadataSyncer.vcenter).UpdateVolumeMetadata(updateSpec); err != nil {
	//	klog.Errorf("PVCUpdated: UpdateVolumeMetadata failed with err %v", err)
	//}
}

// pvDeleted deletes pvc metadata on VC when pvc has been deleted on K8s cluster
func pvcDeleted(obj interface{}, metadataSyncer *MetadataSyncInformer) {
	pvc, ok := obj.(*v1.PersistentVolumeClaim)
	if pvc == nil || !ok {
		klog.Warningf("PVCDeleted: unrecognized object %+v", obj)
		return
	}
	klog.V(5).Infof("PVCDeleted: PVC %s bond to PV %s", pvc.Name, pvc.Spec.VolumeName)
	if pvc.Status.Phase != v1.ClaimBound {
		return
	}
	// Get pv object attached to pvc
	pv, err := metadataSyncer.pvLister.Get(pvc.Spec.VolumeName)
	if pv == nil || err != nil {
		klog.Errorf("PVCDeleted: Error getting Persistent Volume for pvc %s in namespace %s with err: %v", pvc.Name, pvc.Namespace, err)
		return
	}

	// Verify if pv is a ics csi volume
	if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != service.Name {
		klog.V(3).Infof("PVCDeleted: PV %s is not a IncloudSphere CSI Volume", pv.Name)
		return
	}

	// Volume will be deleted by controller when reclaim policy is delete
	if pv.Spec.PersistentVolumeReclaimPolicy == v1.PersistentVolumeReclaimDelete {
		klog.V(3).Infof("PVCDeleted: PV %s Id %s Reclaim policy is Delete", pv.Name, pv.Spec.CSI.VolumeHandle)
		return
	}

	// If the PV reclaim policy is retain we need to delete PVC labels
	var metadataList []cnstypes.BaseCnsEntityMetadata
	pvcMetadata := getCnsKubernetesEntityMetaData(pvc.Name, nil, true, string(cnstypes.CnsKubernetesEntityTypePVC), pvc.Namespace)
	metadataList = append(metadataList, cnstypes.BaseCnsEntityMetadata(pvcMetadata))

	updateSpec := getCnsVolumeMetadataUpdateSpec(pv.Spec.CSI.VolumeHandle, metadataList, metadataSyncer)
	klog.V(4).Infof("PVCDeleted: Calling UpdateVolumeMetadata for volume %s with updateSpec: %+v", updateSpec.VolumeId, spew.Sdump(updateSpec))
	//if err := ics.GetVolumeManager(metadataSyncer.vcenter).UpdateVolumeMetadata(updateSpec); err != nil {
	//	klog.Errorf("PVCDeleted: UpdateVolumeMetadata failed with err %v", err)
	//}
}

// pvUpdated updates volume metadata on VC when volume labels on K8S cluster have been updated
func pvUpdated(oldObj, newObj interface{}, metadataSyncer *MetadataSyncInformer) {
	// Get old and new PV objects
	oldPv, ok := oldObj.(*v1.PersistentVolume)
	if oldPv == nil || !ok {
		klog.Warningf("PVUpdated: unrecognized old object %+v", oldObj)
		return
	}

	newPv, ok := newObj.(*v1.PersistentVolume)
	if newPv == nil || !ok {
		klog.Warningf("PVUpdated: unrecognized new object %+v", newObj)
		return
	}

	// Verify if pv is a ics csi volume
	if oldPv.Spec.CSI == nil || newPv.Spec.CSI == nil || newPv.Spec.CSI.Driver != service.Name {
		klog.V(3).Infof("PVUpdated: PV %s is not a IncloudSphere CSI Volume", newPv.Name)
		return
	}
	// Return if new PV status is Pending or Failed
	if newPv.Status.Phase == v1.VolumePending || newPv.Status.Phase == v1.VolumeFailed {
		klog.V(3).Infof("PVUpdated: PV %s Id %s (newPhase: %s) metadata is not updated", newPv.Name, newPv.Spec.CSI.VolumeHandle, newPv.Status.Phase)
		return
	}
	// Return if labels are unchanged
	if oldPv.Status.Phase == v1.VolumeAvailable && reflect.DeepEqual(newPv.GetLabels(), oldPv.GetLabels()) {
		klog.V(3).Infof("PVUpdated: PV %s Id %s (oldPhase: %s) labels have not changed", oldPv.Name, oldPv.Spec.CSI.VolumeHandle, oldPv.Status.Phase)
		return
	}
	if oldPv.Status.Phase == v1.VolumeBound && newPv.Status.Phase == v1.VolumeReleased && oldPv.Spec.PersistentVolumeReclaimPolicy == v1.PersistentVolumeReclaimDelete {
		klog.V(3).Infof("PVUpdated: PV %s Id %s (oldPhase: %s newPhase: %s reclaimPolicy: %s) will be deleted by controller", newPv.Name,
			newPv.Spec.CSI.VolumeHandle, oldPv.Status.Phase, newPv.Status.Phase, oldPv.Spec.PersistentVolumeReclaimPolicy)
		return
	}
	if newPv.DeletionTimestamp != nil {
		klog.V(3).Infof("PVUpdated: PV %s Id %s already deleted", newPv.Name, newPv.Spec.CSI.VolumeHandle)
		return
	}

	klog.V(5).Infof("PVUpdated: PV %s Id %s (oldPhase: %s newPhase: %s) labels updated from %+v to  %+v", newPv.Name, newPv.Spec.CSI.VolumeHandle,
		oldPv.Status.Phase, newPv.Status.Phase, oldPv.GetLabels(), newPv.GetLabels())

	var metadataList []cnstypes.BaseCnsEntityMetadata
	pvMetadata := getCnsKubernetesEntityMetaData(newPv.Name, newPv.GetLabels(), false, string(cnstypes.CnsKubernetesEntityTypePV), newPv.Namespace)
	metadataList = append(metadataList, cnstypes.BaseCnsEntityMetadata(pvMetadata))

	if oldPv.Status.Phase == v1.VolumeAvailable || newPv.Spec.StorageClassName != "" {
		updateSpec := getCnsVolumeMetadataUpdateSpec(newPv.Spec.CSI.VolumeHandle, metadataList, metadataSyncer)
		klog.V(4).Infof("PVUpdated: Calling UpdateVolumeMetadata for volume %s with updateSpec: %+v", updateSpec.VolumeId, spew.Sdump(updateSpec))
		//if err := ics.GetVolumeManager(metadataSyncer.vcenter).UpdateVolumeMetadata(updateSpec); err != nil {
		//	klog.Errorf("PVUpdated: UpdateVolumeMetadata failed with err %v", err)
		//}
	} else {
		//CreateVolume
	}
}

// pvDeleted deletes volume metadata on VC when volume has been deleted on K8s cluster
func pvDeleted(obj interface{}, metadataSyncer *MetadataSyncInformer) {
	pv, ok := obj.(*v1.PersistentVolume)
	if pv == nil || !ok {
		klog.Warningf("PVDeleted: unrecognized object %+v", obj)
		return
	}
	//klog.V(7).Infof("PVDeleted: Deleting PV: %+v", pv)

	// Verify if pv is a ics csi volume
	if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != service.Name {
		klog.V(3).Infof("PVDeleted: PV %s is not a IncloudSphere CSI Volume", pv.Name)
		return
	}
	var deleteDisk bool
	if pv.Spec.ClaimRef != nil && (pv.Status.Phase == v1.VolumeAvailable || pv.Status.Phase == v1.VolumeReleased) && pv.Spec.PersistentVolumeReclaimPolicy == v1.PersistentVolumeReclaimDelete {
		klog.V(3).Infof("PVDeleted: PV %s Id %s (phase: %s reclaimPolicy: %s) deletion will be handled by Controller", pv.Name, pv.Spec.CSI.VolumeHandle,
			pv.Status.Phase, pv.Spec.PersistentVolumeReclaimPolicy)
		return
	}

	if pv.Spec.ClaimRef == nil || (pv.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimDelete) {
		deleteDisk = false
	} else {
		// We set delete disk=true for the case where PV status is failed after deletion of pvc
		// In this case, metadatasyncer will remove the volume
		deleteDisk = true
	}

	klog.V(4).Infof("PVDeleted: PV %s Id %s (phase: %s reclaimPolicy: %s) setting DeleteDisk to %v", pv.Name, pv.Spec.CSI.VolumeHandle,
		pv.Status.Phase, pv.Spec.PersistentVolumeReclaimPolicy, deleteDisk)

	volumeOperationsLock.Lock()
	defer volumeOperationsLock.Unlock()
	klog.V(4).Infof("PVDeleted: Deleting PV %s Id %s  with deleteDisk %v", pv.Name, pv.Spec.CSI.VolumeHandle, deleteDisk)
	if err := ics.GetVolumeManager(metadataSyncer.vcenter).DeleteVolume(pv.Spec.CSI.VolumeHandle, deleteDisk); err != nil {
		klog.Errorf("PVDeleted: Failed to delete PV %s Id %s with error %+v", pv.Name, pv.Spec.CSI.VolumeHandle, err)
		return
	}
}

// podUpdated updates pod metadata on VC when pod labels have been updated on K8s cluster
func podUpdated(oldObj, newObj interface{}, metadataSyncer *MetadataSyncInformer) {
	// Get old and new pod objects
	oldPod, ok := oldObj.(*v1.Pod)
	if oldPod == nil || !ok {
		klog.Warningf("PodUpdated: unrecognized old object %+v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if newPod == nil || !ok {
		klog.Warningf("PodUpdated: unrecognized new object %+v", newObj)
		return
	}

	// If old pod is in pending state and new pod is running, update metadata
	if oldPod.Status.Phase == v1.PodPending && newPod.Status.Phase == v1.PodRunning {
		klog.V(3).Infof("PodUpdated: Pod %s (oldPhase: %s newPhase: %s) calling updatePodMetadata", newPod.Name,
			oldPod.Status.Phase, newPod.Status.Phase)
		// Update pod metadata
		if errorList := updatePodMetadata(newPod, metadataSyncer, false); len(errorList) > 0 {
			klog.Errorf("PodUpdated: updatePodMetadata failed for pod %s with errors: ", newPod.Name)
			for _, err := range errorList {
				klog.Errorf("PodUpdated: %v", err)
			}
		}
	}
}

func podDeleted(obj interface{}, metadataSyncer *MetadataSyncInformer) {
	// Get pod object
	pod, ok := obj.(*v1.Pod)
	if pod == nil || !ok {
		klog.Warningf("PodDeleted: unrecognized new object %+v", obj)
		return
	}

	if pod.Status.Phase == v1.PodPending {
		return
	}

	klog.V(3).Infof("PodDeleted: Pod %s calling updatePodMetadata", pod.Name)
	// Update pod metadata
	if errorList := updatePodMetadata(pod, metadataSyncer, true); len(errorList) > 0 {
		klog.Errorf("PodDeleted: updatePodMetadata failed for pod %s with errors: ", pod.Name)
		for _, err := range errorList {
			klog.Errorf("PodDeleted: %v", err)
		}
	}
}

// updatePodMetadata updates metadata for volumes attached to the pod
func updatePodMetadata(pod *v1.Pod, metadataSyncer *MetadataSyncInformer, deleteFlag bool) []error {
	var errorList []error
	// Iterate through volumes attached to pod
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.ClaimName
			// Get pvc attached to pod
			pvc, err := metadataSyncer.pvcLister.PersistentVolumeClaims(pod.Namespace).Get(pvcName)
			if err != nil {
				msg := fmt.Sprintf("Error getting Persistent Volume Claim for volume %s with err: %v", volume.Name, err)
				errorList = append(errorList, errors.New(msg))
				continue
			}

			// Get pv object attached to pvc
			pv, err := metadataSyncer.pvLister.Get(pvc.Spec.VolumeName)
			if err != nil {
				msg := fmt.Sprintf("Error getting Persistent Volume for PVC %s in volume %s with err: %v", pvc.Name, volume.Name, err)
				errorList = append(errorList, errors.New(msg))
				continue
			}

			// Verify if pv is ics csi volume
			if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != service.Name {
				klog.V(3).Infof("PV %s is not a IncloudSphere CSI Volume", pv.Name)
				continue
			}
			var metadataList []cnstypes.BaseCnsEntityMetadata
			podMetadata := getCnsKubernetesEntityMetaData(pod.Name, nil, deleteFlag, string(cnstypes.CnsKubernetesEntityTypePOD), pod.Namespace)
			metadataList = append(metadataList, cnstypes.BaseCnsEntityMetadata(podMetadata))
			updateSpec := getCnsVolumeMetadataUpdateSpec(pv.Spec.CSI.VolumeHandle, metadataList, metadataSyncer)
			klog.V(4).Infof("Calling UpdateVolumeMetadata for volume %s with updateSpec: %+v", updateSpec.VolumeId, spew.Sdump(updateSpec))
			//if err := ics.GetVolumeManager(metadataSyncer.vcenter).UpdateVolumeMetadata(updateSpec); err != nil {
			//	msg := fmt.Sprintf("UpdateVolumeMetadata failed for volume %s with err: %v", volume.Name, err)
			//	errorList = append(errorList, errors.New(msg))
			//}
		}
	}
	return errorList
}
