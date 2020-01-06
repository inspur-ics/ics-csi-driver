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
package ics

import (
	"encoding/json"
	"fmt"
	"ics-csi-driver/pkg/common/rest"
	"k8s.io/klog"
)

type Storage struct {
	DataStoreType       string  `json:"dataStoreType"`
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	MountPath           string  `json:"mountPath"`
	Capacity            float64 `json:"capacity"`
	UsedCapacity        float64 `json:"usedCapacity"`
	AvailCapacity       float64 `json:"availCapacity"`
	DataCenterID        string  `json:"dataCenterId"`
	HostID              string  `json:"hostId"`
	MountStatus         string  `json:"mountStatus"`
	HostIP              string  `json:"hostIp"`
	UUID                string  `json:"uuid"`
	AbsolutePath        string  `json:"absolutePath"`
	DataCenterName      string  `json:"dataCenterName"`
	DataCenterOrHostDto struct {
		DataCenterOrHost string `json:"dataCenterOrHost"`
		DataCenterName   string `json:"dataCenterName"`
		HostName         string `json:"hostName"`
		Status           string `json:"status"`
	} `json:"dataCenterOrHostDto"`
	BlockDeviceDto    string `json:"blockDeviceDto"`
	DataCenterDto     string `json:"dataCenterDto"`
	HostNumbers       int    `json:"hostNumbers"`
	VMNumbers         int    `json:"vmNumbers"`
	VolumesNumbers    int    `json:"volumesNumbers"`
	VMTemplateNumbers int    `json:"vmTemplateNumbers"`
	Tags              string `json:"tags"`
	MaxSlots          int    `json:"maxSlots"`
	Creating          bool   `json:"creating"`
	StorageBackUp     bool   `json:"storageBackUp"`
	ExtensionType     string `json:"extensionType"`
	CanBeImageStorage bool   `json:"canBeImageStorage"`
	BlockDeviceUUID   string `json:"blockDeviceUuid"`
	OpHostIP          string `json:"opHostIp"`
	IsMount           string `json:"isMount"`
	HostDto           string `json:"hostDto"`
	ScvmOn            bool   `json:"scvmOn"`
}
type getStoragesReq struct {
	PageSize    int    `json:"pageSize"`
	CurrentPage int    `json:"currentPage"`
	SortField   string `json:"sortField"`
	Sort        string `json:"sort"`
}
type getStoragesRsp struct {
	TotalPage   int       `json:"totalPage"`
	CurrentPage int       `json:"currentPage"`
	TotalSize   int       `json:"totalSize"`
	Ttems       []Storage `json:"items"`
}

func getStorages(restReq getStoragesReq) ([]Storage, rest.RestError) {
	rp, err := rest.NewRestProxy()
	storages := []Storage{}
	if err != nil {
		msg := fmt.Sprintf("cannot create REST client: %s", err.Error())
		return storages, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	stat, body, err := rp.Send("GET", "storages", restReq)
	fmt.Printf("get Storages rsp:%v\n", stat)
	if err != nil {
		msg := fmt.Sprintf("get Storages  failed: %s", err.Error())
		return storages, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	var getStoragesR getStoragesRsp
	if err := json.Unmarshal(body, &getStoragesR); err != nil {
		msg := fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return storages, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	storages = getStoragesR.Ttems
	klog.V(4).Infof("get vms: %s\n", storages)
	return storages, nil
}
func GetStoragesList() ([]Storage, rest.RestError) {
	getStoragesR := getStoragesReq{
		1000,
		1,
		"",
		"desc",
	}

	storages, err := getStorages(getStoragesR)
	if err != nil {
		klog.V(4).Infof("get  vmlist failed  with args %+v", storages)
	}
	return storages, err
}
func GetStorageInfo(storageid string) (Storage, rest.RestError) {
	rp, err := rest.NewRestProxy()
	storageinfo := Storage{}
	if err != nil {
		msg := fmt.Sprintf("cannot create REST client: %s", err.Error())
		return storageinfo, rest.GetError(rest.RestRequestMalfunction, msg)
	}
	storagepath := "storages/" + storageid
	stat, body, err := rp.Send("GET", storagepath, nil)
	fmt.Printf("get storageinfo rsp:%v\n", stat)
	if err != nil {
		msg := fmt.Sprintf("get storageinfo  failed: %s", err.Error())
		return storageinfo, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &storageinfo); err != nil {
		msg := fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return storageinfo, rest.GetError(rest.RestRequestMalfunction, msg)
	}
	klog.V(4).Infof("get storageinfo: %s\n", storageinfo)
	return storageinfo, nil
}
