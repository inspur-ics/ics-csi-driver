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
	//"context"
	"fmt"
	"sync"
	//csictx "github.com/rexray/gocsi/context"
	//"k8s.io/klog"
	//cnsconfig "ics-csi-driver/pkg/common/config"
)

const (
	// DefaultScheme is the default connection scheme.
	DefaultScheme = "https"
	// DefaultRoundTripperCount is the default SOAP round tripper count.
	DefaultRoundTripperCount = 3
)

// VirtualCenter holds details of a virtual center instance.
type VirtualCenter struct {
	// Config represents the virtual center configuration.
	Config *VirtualCenterConfig
	// Client represents the govmomi client instance for the connection.
	//Client *govmomi.Client
	// PbmClient represents the govmomi PBM Client instance.
	//PbmClient *pbm.Client
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
