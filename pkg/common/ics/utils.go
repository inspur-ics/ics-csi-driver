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

package ics

import (
	"context"
	"errors"
	"fmt"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	icsgo "github.com/inspur-ics/ics-go-sdk/common"
	icstag "github.com/inspur-ics/ics-go-sdk/tag"
	"ics-csi-driver/pkg/common/config"
	"k8s.io/klog"
	"strconv"
	"strings"
)

// GetVirtualCenterConfig returns VirtualCenterConfig Object created using vSphere Configuration
// specified in the argurment.
func GetVirtualCenterConfig(cfg *config.Config) (*VirtualCenterConfig, error) {
	var err error
	vCenterIPs, err := GetVcenterIPs(cfg) //  make([]string, 0)
	if err != nil {
		return nil, err
	}
	host := vCenterIPs[0]
	port, err := strconv.Atoi(cfg.VirtualCenter[host].VCenterPort)
	if err != nil {
		return nil, err
	}
	vcConfig := &VirtualCenterConfig{
		Host:            host,
		Port:            port,
		Username:        cfg.VirtualCenter[host].User,
		Password:        cfg.VirtualCenter[host].Password,
		Insecure:        cfg.VirtualCenter[host].InsecureFlag,
		DatacenterPaths: strings.Split(cfg.VirtualCenter[host].Datacenters, ","),
	}
	for idx := 0; idx < len(vcConfig.DatacenterPaths); {
		vcConfig.DatacenterPaths[idx] = strings.TrimSpace(vcConfig.DatacenterPaths[idx])
		if vcConfig.DatacenterPaths[idx] == "" {
			//delete the empty datacenter elem
			vcConfig.DatacenterPaths = append(vcConfig.DatacenterPaths[:idx], vcConfig.DatacenterPaths[idx+1:]...)
		} else {
			idx++
		}
	}
	return vcConfig, nil
}

// GetVcenterIPs returns list of vCenter IPs from VSphereConfig
func GetVcenterIPs(cfg *config.Config) ([]string, error) {
	var err error
	vCenterIPs := make([]string, 0)
	for key := range cfg.VirtualCenter {
		vCenterIPs = append(vCenterIPs, key)
	}
	if len(vCenterIPs) == 0 {
		err = errors.New("Unable get vCenter Hosts from Config")
	}
	return vCenterIPs, err
}

func GetAttachedTags(ctx context.Context, vc *VirtualCenter, targetType string, targetId string) ([]types.Tag, error) {
	tagService := icstag.NewTagsService(vc.Client)
	tagList, err := tagService.ListAttachedTags(ctx, targetType, targetId)
	if err != nil {
		klog.Errorf("Get attached tag failed for %s  %s with err: %v", targetType, targetId, err)
		return nil, err
	}

	var tags []types.Tag
	for _, tagId := range tagList {
		tag, err := tagService.GetTag(ctx, tagId)
		if err != nil {
			klog.Errorf("Get tag %s info failed with err: %v", tagId, err)
			return tags, err
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}

func GetTaskState(ctx context.Context, vc *VirtualCenter, task *types.Task) (string, error) {
	state := "Unknown"
	if task == nil {
		return state, errors.New("Task value is nil")
	} else if task.TaskId == "" {
		return state, errors.New("TaskId is empty")
	}

	restapi := &icsgo.RestAPI{
		RestAPITripper: vc.Client,
	}
	taskInfo, err := restapi.TraceTaskProcess(task)
	if err != nil || taskInfo == nil {
		errMsg := fmt.Sprintf("Failed to get  task %s state with err: %+v", task.TaskId, err)
		return state, errors.New(errMsg)
	}

	klog.V(5).Infof("Task %s state: %+v", task.TaskId, taskInfo)
	return taskInfo.State, nil
}
