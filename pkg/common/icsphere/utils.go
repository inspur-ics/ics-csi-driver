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
	"errors"
	"ics-csi-driver/pkg/common/config"
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
