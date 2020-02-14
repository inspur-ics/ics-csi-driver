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

package icsphere

import (
	"context"
	"errors"
	"github.com/inspur-ics/ics-go-sdk/client/types"
	icstag "github.com/inspur-ics/ics-go-sdk/tag"
	"ics-csi-driver/pkg/common/config"
	"ics-csi-driver/pkg/common/rest"
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
	for idx := range vcConfig.DatacenterPaths {
		vcConfig.DatacenterPaths[idx] = strings.TrimSpace(vcConfig.DatacenterPaths[idx])
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

func GetDatacenterTopologys(ctx context.Context) ([]rest.DataCenterTopology, error) {
	rp, err := rest.NewRestProxy()
	if err != nil {
		klog.Error("create restProxy failed.")
		return nil, err
	}

	dcTopologys, err := rest.GetDataCenterTopology(rp)
	if err != nil {
		klog.Error("get datacenter topology failed.")
		return nil, err
	}

	return dcTopologys, nil
}

func GetAttachedTags(ctx context.Context, vchost string, targetType string, targetId string) ([]types.Tag, error) {
	vc, err := GetVirtualCenterManager().GetVirtualCenter(vchost)
	if err != nil {
		klog.Errorf("Failed to get iCenter %s with err: %v", vchost, err)
		return nil, err
	}
	if err := vc.Connect(ctx); err != nil {
		return nil, err
	}

	tagService := icstag.NewTagsService(vc.Client)
	tagList, err := tagService.ListAttachedTags(ctx, targetType, targetId)
	if err != nil {
		klog.Errorf("Get attached tag failed for %s  %s", targetType, targetId)
		return nil, err
	}

	var tags []types.Tag
	for _, tagId := range tagList {
		tag, err := tagService.GetTag(ctx, tagId)
		if err != nil {
			klog.Errorf("Get tag %s info failed", tagId)
			return tags, err
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}
