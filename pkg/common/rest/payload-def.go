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

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Locale   string `json:"locale"`
	Domain   string `json:"domain"`
	Captcha  string `json:"captcha"`
}

type LoginRsp struct {
	UserId     string `json:"userId"`
	SessionId  string `json:"sessonId"`
	Validated  bool   `json:"validated"`
	Message    string `json:"message"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Captcha    string `json:"captcha"`
	Locale     string `json:"locale"`
	Domain     string `json:"domain"`
	Remains    int    `json:"remains"`
	Ip         string `json:"ip"`
	Operator   string `json:"operator"`
	LoginTime  string `json:"loginTime"`
	CreateData string `json:"createDate"`
}

type CreateVolumeReq struct {
	Name          string `json:"name"`
	Size          string `json:"size"`
	DataStoreType string `json:"dataStoreType"`
	DataStoreId   string `json:"dataStoreId"`
	VolumePolicy  string `json:"volumePolicy"`
	Description   string `json:"description"`
	Bootable      bool   `json:"bootable"`
	Shared        bool   `json:"shared"`
}

type CreateVolumeRsp struct {
	TaskId string `json:"taskId"`
}
