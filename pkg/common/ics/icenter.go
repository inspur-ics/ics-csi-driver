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
	"fmt"
	icsgo "github.com/inspur-ics/ics-go-sdk"
	"github.com/inspur-ics/ics-go-sdk/client"
	icsdc "github.com/inspur-ics/ics-go-sdk/datacenter"
	"k8s.io/klog"
	"strconv"
	"sync"
)

// VirtualCenter holds details of a virtual center instance.
type VirtualCenter struct {
	// Config represents the virtual center configuration.
	Config *VirtualCenterConfig
	// Client represents the govmomi client instance for the connection.
	Client *client.Client
	// CnsClient represents the CNS client instance.
	//CnsClient       *cns.Client
	credentialsLock sync.Mutex
}

// VirtualCenterConfig represents virtual center configuration.
type VirtualCenterConfig struct {
	// Scheme represents the connection scheme. (Ex: https)
	Scheme string
	// Host represents the virtual center host address.
	Host string
	// Port represents the virtual center host port.
	Port int
	// Username represents the virtual center username.
	Username string
	// Password represents the virtual center password in clear text.
	Password string
	// Insecure tells if an insecure connection is allowed.
	Insecure bool
	// RoundTripperCount is the SOAP round tripper count. (retries = RoundTripperCount - 1)
	RoundTripperCount int
	// DatacenterPaths represents paths of datacenters on the virtual center.
	DatacenterPaths []string
}

func (vcc *VirtualCenterConfig) String() string {
	return fmt.Sprintf("VirtualCenterConfig [Scheme: %v, Host: %v, Port: %v, "+
		"Username: %v, Password: %v, Insecure: %v, RoundTripperCount: %v, "+
		"DatacenterPaths: %v]", vcc.Scheme, vcc.Host, vcc.Port, vcc.Username,
		vcc.Password, vcc.Insecure, vcc.RoundTripperCount, vcc.DatacenterPaths)
}

// connect creates a connection to the virtual center host.
func (vc *VirtualCenter) Connect(ctx context.Context) error {
	conn := &icsgo.ICSConnection{
		Username: vc.Config.Username,
		Password: vc.Config.Password,
		Hostname: vc.Config.Host,
		Port:     strconv.Itoa(vc.Config.Port),
		Insecure: vc.Config.Insecure,
	}

	client, err := conn.GetClient()
	if err != nil {
		klog.Errorf("virtual center connect failed: vc %s\n", vc.Config.Host)
		return err
	}
	vc.Client = client
	klog.V(4).Infof("virtual center connect successfully: vc %s\n", vc.Config.Host)
	return nil
}

// GetDatacenters returns Datacenters found on the VirtualCenter. If no
// datacenters are mentioned in the VirtualCenterConfig during registration, all
// Datacenters for the given VirtualCenter will be returned. If DatacenterPaths
// is configured in VirtualCenterConfig during registration, only the listed
// Datacenters are returned.
func (vc *VirtualCenter) GetDatacenters(ctx context.Context) ([]*Datacenter, error) {
	var dcs []*Datacenter
	dcService := icsdc.NewDatacenterService(vc.Client)
	dcList, err := dcService.GetAllDatacenters(ctx)
	if err != nil {
		klog.Errorf("get datacenter list faild for vc: %s\n", vc.Config.Host)
	} else {
		klog.V(5).Infof("successfully get datacenter list for vc: %s\n", vc.Config.Host)
		for _, dcItem := range dcList {
			dcExist := false
			if len(vc.Config.DatacenterPaths) == 0 {
				dcExist = true
			} else {
				for _, dcPath := range vc.Config.DatacenterPaths {
					if dcItem.ID == dcPath || dcItem.Name == dcPath {
						dcExist = true
						break
					}
				}
			}
			if dcExist {
				dc := &Datacenter{ID: dcItem.ID, Datacenter: dcItem, VirtualCenterHost: vc.Config.Host}
				dcs = append(dcs, dc)
			}
		}
	}
	return dcs, err
}
