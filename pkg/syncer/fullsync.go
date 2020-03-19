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
	"sync"

	//"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	//"ics-csi-driver/pkg/common/ics"
	"ics-csi-driver/pkg/csi/service"
)

// triggerFullSync triggers full sync
func triggerFullSync(k8sclient clientset.Interface, metadataSyncer *MetadataSyncInformer) {
	klog.V(2).Infof("FullSync: start")

	// Get K8s PVs in State "Bound", "Available" or "Released"
	k8sPVs, err := getPVsInBoundAvailableOrReleased(k8sclient)
	if err != nil {
		klog.Warningf("FullSync: Failed to get PVs from kubernetes. Err: %v", err)
		return
	}

	// pvToPVCMap maps pv name to corresponding PVC
	// pvcToPodMap maps pvc to the mounted Pod
	pvToPVCMap, pvcToPodMap := buildPVCMapPodMap(k8sclient, k8sPVs)
	klog.V(4).Infof("FullSync: pvToPVCMap %v", pvToPVCMap)
	klog.V(4).Infof("FullSync: pvcToPodMap %v", pvcToPodMap)

	var createSpecArray []CnsVolumeCreateSpec
	var volToBeDeleted []CnsVolumeId
	var updateSpecArray []CnsVolumeMetadataUpdateSpec

	wg := sync.WaitGroup{}
	wg.Add(3)
	// Perform operations
	go fullSyncCreateVolumes(createSpecArray, metadataSyncer, k8sclient, &wg)
	go fullSyncDeleteVolumes(volToBeDeleted, metadataSyncer, k8sclient, &wg)
	go fullSyncUpdateVolumes(updateSpecArray, metadataSyncer, &wg)
	wg.Wait()

	//cleanupCnsMaps(k8sPVsMap)
	klog.V(4).Infof("FullSync: cnsDeletionMap at end of cycle: %v", cnsDeletionMap)
	klog.V(4).Infof("FullSync: cnsCreationMap at end of cycle: %v", cnsCreationMap)
}

// getPVsInBoundAvailableOrReleased return PVs in Bound, Available or Released state
func getPVsInBoundAvailableOrReleased(k8sclient clientset.Interface) ([]*v1.PersistentVolume, error) {
	var pvsInDesiredState []*v1.PersistentVolume
	// Get all PVs from kubernetes
	allPVs, err := k8sclient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for index, pv := range allPVs.Items {
		if pv.Spec.CSI != nil && pv.Spec.CSI.Driver == service.Name {
			klog.V(4).Infof("FullSync: pv %v is in state %v", pv.Spec.CSI.VolumeHandle, pv.Status.Phase)
			if pv.Status.Phase == v1.VolumeBound || pv.Status.Phase == v1.VolumeAvailable || pv.Status.Phase == v1.VolumeReleased {
				pvsInDesiredState = append(pvsInDesiredState, &allPVs.Items[index])
			}
		}
	}
	return pvsInDesiredState, nil
}

// buildPVCMapPodMap build two maps to help
//  1. find PVC for given PV
//  2. find POD mounted to given PVC
// pvToPVCMap maps PV name to corresponding PVC, key is pv name
// pvcToPodMap maps PVC to the POD attached to the PVC, key is "pvc.Namespace/pvc.Name"
func buildPVCMapPodMap(k8sclient clientset.Interface, pvList []*v1.PersistentVolume) (pvcMap, podMap) {
	pvToPVCMap := make(pvcMap)
	pvcToPodMap := make(podMap)
	for _, pv := range pvList {
		if pv.Spec.ClaimRef != nil && pv.Status.Phase == v1.VolumeBound {
			pvc, err := k8sclient.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name, metav1.GetOptions{})
			if err != nil {
				klog.Warningf("FullSync: Failed to get pvc for namespace %v and name %v. err=%v", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name, err)
				continue
			}
			pvToPVCMap[pv.Name] = pvc
			klog.V(4).Infof("FullSync: pvc %v is backed by pv %v", pvc.Name, pv.Name)
			pods, err := k8sclient.CoreV1().Pods(pvc.Namespace).List(metav1.ListOptions{
				FieldSelector: fields.AndSelectors(fields.SelectorFromSet(fields.Set{"status.phase": string(v1.PodRunning)})).String(),
			})
			if err != nil {
				klog.Warningf("FullSync: Failed to get pods for namespace %v. err=%v", pvc.Namespace, err)
				continue
			}
			for index, pod := range pods.Items {
				if pod.Spec.Volumes != nil {
					for _, volume := range pod.Spec.Volumes {
						pvClaim := volume.VolumeSource.PersistentVolumeClaim
						if pvClaim != nil && pvClaim.ClaimName == pvc.Name {
							key := pod.Namespace + "/" + pvClaim.ClaimName
							pvcToPodMap[key] = &pods.Items[index]
							klog.V(4).Infof("FullSync: pvc %v is mounted by pod %v", key, pod.Name)
							break
						}
					}
				}
			}

		}
	}
	return pvToPVCMap, pvcToPodMap
}

// fullSyncCreateVolumes create volumes with given array of createSpec
// Before creating a volume, all current K8s volumes are retrieved
// If the volume is successfully created, it is removed from cnsCreationMap
func fullSyncCreateVolumes(createSpecArray []CnsVolumeCreateSpec, metadataSyncer *MetadataSyncInformer, k8sclient clientset.Interface, wg *sync.WaitGroup) {
	defer wg.Done()
	currentK8sPVMap := make(map[string]bool)
	volumeOperationsLock.Lock()
	defer volumeOperationsLock.Unlock()
	// Get all K8s PVs
	currentK8sPV, err := getPVsInBoundAvailableOrReleased(k8sclient)
	if err != nil {
		klog.Errorf("FullSync: fullSyncCreateVolumes failed to get PVs from kubernetes. Err: %v", err)
		return
	}
	// Create map for easy lookup
	for _, pv := range currentK8sPV {
		currentK8sPVMap[pv.Spec.CSI.VolumeHandle] = true
	}

	//for _, createSpec := range createSpecArray {
	// Create volume if present in currentK8sPVMap

	//}
}

// fullSyncDeleteVolumes delete volumes with given array of volumeId
// Before deleting a volume, all current K8s volumes are retrieved
// If the volume is successfully deleted, it is removed from cnsDeletionMap
func fullSyncDeleteVolumes(volumeIDDeleteArray []CnsVolumeId, metadataSyncer *MetadataSyncInformer, k8sclient clientset.Interface, wg *sync.WaitGroup) {
	defer wg.Done()
	//deleteDisk := false
	currentK8sPVMap := make(map[string]bool)
	volumeOperationsLock.Lock()
	defer volumeOperationsLock.Unlock()
	// Get all K8s PVs
	currentK8sPV, err := getPVsInBoundAvailableOrReleased(k8sclient)
	if err != nil {
		klog.Errorf("FullSync: fullSyncDeleteVolumes failed to get PVs from kubernetes. Err: %v", err)
		return
	}
	// Create map for easy lookup
	for _, pv := range currentK8sPV {
		currentK8sPVMap[pv.Spec.CSI.VolumeHandle] = true
	}

	//for _, volID := range volumeIDDeleteArray {
	// Delete volume if not present in currentK8sPVMap

	//}
}

// fullSyncUpdateVolumes update metadata for volumes with given array of createSpec
func fullSyncUpdateVolumes(updateSpecArray []CnsVolumeMetadataUpdateSpec, metadataSyncer *MetadataSyncInformer, wg *sync.WaitGroup) {
	defer wg.Done()
	//for _, updateSpec := range updateSpecArray {
	//
	//}
}

// cleanupCnsMaps performs cleanup on cnsCreationMap and cnsDeletionMap
// Removes volume entries from cnsCreationMap that do not exist in K8s
// and volume entries from cnsDeletionMap that exist in K8s
// An entry could have been added to cnsCreationMap (or cnsDeletionMap)
// because full sync was triggered in between the delete (or create)
// operation of a volume
func cleanupCnsMaps(k8sPVs map[string]string) {
	// Cleanup cnsCreationMap
	for volID := range cnsCreationMap {
		if _, existsInK8s := k8sPVs[volID]; !existsInK8s {
			delete(cnsCreationMap, volID)
		}
	}
	// Cleanup cnsDeletionMap
	for volID := range cnsDeletionMap {
		if _, existsInK8s := k8sPVs[volID]; existsInK8s {
			delete(cnsDeletionMap, volID)
		}
	}
}
