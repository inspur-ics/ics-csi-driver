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
)

func CreateVolume(restReq CreateVolumeReq) (string, RestError) {
	rp, err := NewRestProxy()
	taskId := ""
	if err != nil {
		msg := fmt.Sprintf("cannot create REST client: %s", err.Error())
		return taskId, GetError(RestRequestMalfunction, msg)
	}

	stat, body, err := rp.Send("POST", "volumes", restReq)
	fmt.Printf("create volume rsp:%v\n", stat)
	if err != nil {
		msg := fmt.Sprintf("create volume failed: %s", err.Error())
		return taskId, GetError(RestRequestMalfunction, msg)
	}

	var createVolumeRsp CreateVolumeRsp
	if err := json.Unmarshal(body, &createVolumeRsp); err != nil {
		msg := fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return taskId, GetError(RestRequestMalfunction, msg)
	}

	taskId = createVolumeRsp.TaskId
	klog.V(4).Infof("create volume taskId: %s\n", taskId)
	return taskId, nil
}
