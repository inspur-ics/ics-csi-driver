/*
Copyright (c) 2019 Inspur, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package rest

import (
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"time"
)

func CreateVolume(restReq CreateVolumeReq) (string, RestError) {
	var msg string
	rp, err := NewRestProxy()
	taskId, volumeId := "", ""
	if err != nil {
		msg = fmt.Sprintf("cannot create REST client: %s", err.Error())
		return taskId, GetError(RestRequestMalfunction, msg)
	}

	stat, body, err := rp.Send("POST", "volumes", restReq)
	klog.V(4).Infof("create volume rsp:%v\n", stat)
	if err != nil {
		msg = fmt.Sprintf("create volume failed: %s", err.Error())
		return volumeId, GetError(RestRequestMalfunction, msg)
	}

	var createVolumeRsp TaskRsp
	if err := json.Unmarshal(body, &createVolumeRsp); err != nil {
		msg = fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return volumeId, GetError(RestRequestMalfunction, msg)
	}

	taskId = createVolumeRsp.TaskId
	klog.V(5).Infof("create volume taskId: %s\n", taskId)
	taskState, err := GetTaskState(rp, taskId)
	if err != nil || taskState != "FINISHED" {
		klog.Errorf("taskId %s state %s", taskId, taskState)
		msg = fmt.Sprintf("task %s not finished", taskId)
		return volumeId, GetError(RestRequestMalfunction, msg)
	}
	klog.V(5).Infof("create volume task finished: taskId%s", taskId)

	volumeList, err := GetVolumesInDatastore(rp, restReq.DataStoreId)
	if volumeList != nil && len(volumeList) > 0 {
		for _, volumeInfo := range volumeList {
			if volumeInfo.Name == restReq.Name {
				return volumeInfo.Id, nil
			}
		}
	}
	msg = fmt.Sprintf("volume id not found: volume %s err %s", restReq.Name, err.Error())
	klog.Errorf("%s", msg)
	return volumeId, GetError(RestStorageFailureUnknown, msg)
}

func DeleteVolume(rp RestProxyInterface, volumeId string, deleteVolume bool) (string, RestError) {
	taskId, msg := "", ""
	addr := fmt.Sprintf("volumes/%s/?removeData=%v", volumeId, deleteVolume)
	stat, body, err := rp.Send("DELETE", addr, nil)
	if err != nil {
		msg = fmt.Sprintf("delete volume failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return taskId, GetError(RestRequestMalfunction, msg)
	}

	if stat != 200 {
		var resp interface{}
		if err := json.Unmarshal(body, &resp); err != nil {
			msg = fmt.Sprintf("json unmarshal failed: %s", err.Error())
		}
		msg = fmt.Sprintf("delete volume %s failed: %+v", volumeId, resp)
		klog.Errorf("%s", msg)
		return taskId, GetError(RestRequestMalfunction, msg)
	}

	var taskRsp TaskRsp
	if err := json.Unmarshal(body, &taskRsp); err != nil {
		msg = fmt.Sprintf("task info unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return taskId, GetError(RestRequestMalfunction, msg)
	}
	return taskRsp.TaskId, nil
}

func GetVolumeInfo(rp RestProxyInterface, volumeId string) (VolumeInfoRsp, RestError) {
	var msg string
	var restReq interface{}
	var volInfo VolumeInfoRsp
	addr := fmt.Sprintf("volumes/%s", volumeId)
	stat, body, err := rp.Send("GET", addr, restReq)
	if err != nil {
		msg = fmt.Sprintf("get volume info failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return volInfo, GetError(RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &volInfo); err != nil {
		msg = fmt.Sprintf("volume info unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return volInfo, GetError(RestRequestMalfunction, msg)
	}

	klog.V(5).Infof("volume %s info:%+v", volumeId, volInfo)
	return volInfo, nil
}

func GetVolumesInDatastore(rp RestProxyInterface, datastoreId string) ([]VolumeInfoRsp, RestError) {
	var msg string
	var restReq interface{}
	addr := fmt.Sprintf("storages/%s/volumes", datastoreId)
	stat, body, err := rp.Send("GET", addr, restReq)
	if err != nil {
		msg = fmt.Sprintf("get volume list failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}

	var volumeListRsp VolumeListRsp
	if err := json.Unmarshal(body, &volumeListRsp); err != nil {
		msg = fmt.Sprintf("volume list unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return nil, GetError(RestRequestMalfunction, msg)
	}
	return volumeListRsp.Items, nil
}

func GetVmList(rp RestProxyInterface, vmName string) ([]VmInfoRsp, RestError) {
	var msg string
	var restReq interface{}
	var vmList []VmInfoRsp
	stat, body, err := rp.Send("GET", "vms", restReq)
	if err != nil {
		msg = fmt.Sprintf("get vm list failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return vmList, GetError(RestRequestMalfunction, msg)
	}

	var vmListRsp VmListRsp
	if err := json.Unmarshal(body, &vmListRsp); err != nil {
		msg = fmt.Sprintf("vm list unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return vmList, GetError(RestRequestMalfunction, msg)
	}

	return vmListRsp.Items, nil
}

func GetVmInfo(rp RestProxyInterface, vmId string) (VmInfoRsp, RestError) {
	var msg string
	var restReq interface{}
	var vmInfo VmInfoRsp
	addr := fmt.Sprintf("vms/%s", vmId)
	stat, body, err := rp.Send("GET", addr, restReq)
	if err != nil {
		msg = fmt.Sprintf("get vm info failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return vmInfo, GetError(RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &vmInfo); err != nil {
		msg = fmt.Sprintf("vm info unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return vmInfo, GetError(RestRequestMalfunction, msg)
	}

	return vmInfo, nil
}

func SetVmInfo(rp RestProxyInterface, vmInfoReq VmInfoRsp) (string, RestError) {
	var msg string
	addr := fmt.Sprintf("vms/%s", vmInfoReq.Id)
	stat, body, err := rp.Send("PUT", addr, vmInfoReq)
	if err != nil {
		msg = fmt.Sprintf("set volume info failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return "", GetError(RestRequestMalfunction, msg)
	}
	var taskRsp TaskRsp
	if err := json.Unmarshal(body, &taskRsp); err != nil {
		msg = fmt.Sprintf("task rsp unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
	} else {
		return taskRsp.TaskId, nil
	}

	var rsp interface{}
	if err := json.Unmarshal(body, &rsp); err != nil {
		msg = fmt.Sprintf("rsp unmarshal failed: %s", err.Error())
	} else {
		msg = fmt.Sprintf("set vm info failed:%+v", rsp)
	}
	klog.Errorf("%s", msg)

	return "", GetError(RestRequestMalfunction, msg)
}

func GetTaskState(rp RestProxyInterface, taskId string) (string, RestError) {
	var msg string
	taskState := "unknown"
	maxMillisecond := 60 * time.Second
	sleepMillisecond := 100 * time.Millisecond

	for {
		var restReq interface{}
		time.Sleep(100 * time.Millisecond)
		addr := fmt.Sprintf("tasks/%s", taskId)
		_, body, err := rp.Send("GET", addr, restReq)
		if err != nil {
			msg = fmt.Sprintf("get task stat failed: %s", err.Error())
			klog.Errorf("get task state failed: taskId%s", taskId)
			return taskState, GetError(RestRequestMalfunction, msg)
		}

		var taskInfoRsp TaskInfoRsp
		if err := json.Unmarshal(body, &taskInfoRsp); err != nil {
			msg = fmt.Sprintf("json unmarshal failed: %s", err.Error())
			klog.Errorf("task state unmarshal failed: taskId%s", taskId)
			return taskState, GetError(RestRequestMalfunction, msg)
		}
		taskState = taskInfoRsp.State
		if taskState == "FINISHED" || taskState == "ERROR" {
			klog.V(5).Infof("sleep %d ms to wait task finished", sleepMillisecond/time.Millisecond)
			klog.V(5).Infof("task info:%+v", taskInfoRsp)
			return taskState, nil
		}

		if sleepMillisecond < maxMillisecond {
			time.Sleep(sleepMillisecond)
			sleepMillisecond *= 2
		} else {
			klog.V(4).Infof("get task state timeout: %d ms ", sleepMillisecond/time.Millisecond)
			return taskState, nil
		}
	}
}

func GetDataCenterTopology(rp RestProxyInterface) ([]DataCenterTopology, RestError) {
	var msg string
	var dcTopologys []DataCenterTopology
	stat, body, err := rp.Send("GET", "topologys/dataCenterItemDto", nil)
	if err != nil {
		msg = fmt.Sprintf("get datacenter topology failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return dcTopologys, GetError(RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &dcTopologys); err != nil {
		msg = fmt.Sprintf("datacenter topology unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return dcTopologys, GetError(RestRequestMalfunction, msg)
	}

	return dcTopologys, nil
}

func GetTagList(rp RestProxyInterface) ([]TagInfo, RestError) {
	var msg string
	var tagList []TagInfo
	stat, body, err := rp.Send("GET", "tags", nil)
	if err != nil {
		msg = fmt.Sprintf("get tag list failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return tagList, GetError(RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &tagList); err != nil {
		msg = fmt.Sprintf("tag list unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return tagList, GetError(RestRequestMalfunction, msg)
	}
	return tagList, nil
}

func GetClusterList(rp RestProxyInterface) ([]ClusterInfo, RestError) {
	var msg string
	var clusterList []ClusterInfo
	stat, body, err := rp.Send("GET", "clusters", nil)
	if err != nil {
		msg = fmt.Sprintf("get cluster list failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return clusterList, GetError(RestRequestMalfunction, msg)
	}
	var clusterListRsp ClusterListRsp
	if err := json.Unmarshal(body, &clusterListRsp); err != nil {
		msg = fmt.Sprintf("cluster list unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return clusterList, GetError(RestRequestMalfunction, msg)
	}

	return clusterListRsp.Items, nil
}

func GetHostList(rp RestProxyInterface) ([]HostInfo, RestError) {
	var msg string
	var hostList []HostInfo
	stat, body, err := rp.Send("GET", "hosts", nil)
	if err != nil {
		msg = fmt.Sprintf("get host list failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return hostList, GetError(RestRequestMalfunction, msg)
	}

	var hostListRsp HostListRsp
	if err := json.Unmarshal(body, &hostListRsp); err != nil {
		msg = fmt.Sprintf("host list unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return hostList, GetError(RestRequestMalfunction, msg)
	}

	return hostListRsp.Items, nil
}

func GetHostAccessibleDatastoreList(rp RestProxyInterface, hostId string) ([]DatastoreInfo, RestError) {
	var msg string
	var datastoreList []DatastoreInfo
	addr := fmt.Sprintf("hosts/%s/availstorages", hostId)
	stat, body, err := rp.Send("GET", addr, nil)
	if err != nil {
		msg = fmt.Sprintf("get datastore list failed: stat %s err %s", stat, err.Error())
		klog.Errorf("%s", msg)
		return datastoreList, GetError(RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &datastoreList); err != nil {
		msg = fmt.Sprintf("datastore list unmarshal failed: %s", err.Error())
		klog.Errorf("%s", msg)
		return datastoreList, GetError(RestRequestMalfunction, msg)
	}
	return datastoreList, nil
}
