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
)

var rpc = RestProxyCfg{
	Addr: "10.7.11.90",
	User: "admin",
	Pass: "admin@inspur",
}

type CreateVolumeDesc struct {
	Name          string
	Size          string
	DataStoreType string
	DataStoreId   string
	VolumePolicy  string
	Description   string
	Bootable      bool
	Shared        bool
}

func CreateVolume(req CreateVolumeDesc) RestError {
	rp, err := NewRestProxy(rpc)
	if err != nil {
		msg := fmt.Sprintf("cannot create REST client: %s", err.Error())
		return GetError(RestRequestMalfunction, msg)
	}

	restReq := CreateVolumeReq{
		Name:          "test_qcow_zhanghj_1229_a1",
		Size:          "30",
		DataStoreType: "LOCAL",
		DataStoreId:   "8a878bda6c219bca016c235703da0040",
		VolumePolicy:  "THICK",
		Description:   "k8s",
		Bootable:      false,
		Shared:        false,
	}

	stat, body, err := rp.Send("POST", "volumes", restReq)
	fmt.Printf("create volume rsp:%v\n", stat)
	if err != nil {
		msg := fmt.Sprintf("create volume failed: %s", err.Error())
		return GetError(RestRequestMalfunction, msg)
	}

	var createVolumeRsp CreateVolumeRsp
	if err := json.Unmarshal(body, &createVolumeRsp); err != nil {
		msg := fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return GetError(RestRequestMalfunction, msg)
	}
	fmt.Printf("create volume taskId: %s\n", createVolumeRsp.TaskId)
	return nil
}
