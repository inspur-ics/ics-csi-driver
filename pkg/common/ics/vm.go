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

type getVmsReq struct {
	PageSize    int    `json:"pageSize"`
	CurrentPage int    `json:"currentPage"`
	SortField   string `json:"sortField"`
	Sort        string `json:"sort"`
}
type getVmsRsp struct {
	TotalPage   int  `json:"totalPage"`
	CurrentPage int  `json:"currentPage"`
	TotalSize   int  `json:"totalSize"`
	Ttems       []Vm `json:"items"`
}
type Vm struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	State          string      `json:"state"`
	Status         string      `json:"status"`
	HostID         string      `json:"hostId"`
	HostName       string      `json:"hostName"`
	HostIP         string      `json:"hostIp"`
	HostStatus     string      `json:"hostStatus"`
	HostMemory     float64     `json:"hostMemory"`
	DataCenterID   string      `json:"dataCenterId"`
	HaEnabled      bool        `json:"haEnabled"`
	RouterFlag     bool        `json:"routerFlag"`
	Migratable     bool        `json:"migratable"`
	ToolsInstalled bool        `json:"toolsInstalled"`
	ToolsVersion   string `json:"toolsVersion"`
	ToolsType      string      `json:"toolsType"`
	Description    string      `json:"description"`
	HaMaxLimit     int         `json:"haMaxLimit"`
	Template       bool        `json:"template"`
	Initialized    bool        `json:"initialized"`
	GuestosLabel   string      `json:"guestosLabel"`
	GuestosType    string      `json:"guestosType"`
	GuestOsInfo    struct {
		Model               string `json:"model"`
		SocketLimit         int    `json:"socketLimit"`
		SupportCPUHotPlug   bool   `json:"supportCpuHotPlug"`
		SupportMemHotPlug   bool   `json:"supportMemHotPlug"`
		SupportDiskHotPlug  bool   `json:"supportDiskHotPlug"`
		SupportUefiBootMode bool   `json:"supportUefiBootMode"`
	} `json:"guestOsInfo"`
	InnerName         string        `json:"innerName"`
	UUID              string        `json:"uuid"`
	MaxMemory         int           `json:"maxMemory"`
	Memory            int           `json:"memory"`
	MemoryUsage       float64       `json:"memoryUsage"`
	MemHotplugEnabled bool          `json:"memHotplugEnabled"`
	CPUNum            int           `json:"cpuNum"`
	CPUSocket         int           `json:"cpuSocket"`
	CPUCore           int           `json:"cpuCore"`
	CPUUsage          float64       `json:"cpuUsage"`
	MaxCPUNum         int           `json:"maxCpuNum"`
	CPUHotplugEnabled bool          `json:"cpuHotplugEnabled"`
	CPUModelType      string        `json:"cpuModelType"`
	CPUModelEnabled   bool          `json:"cpuModelEnabled"`
	RunningTime       float64       `json:"runningTime"`
	Boot              string        `json:"boot"`
	BootMode          string        `json:"bootMode"`
	SplashTime        int           `json:"splashTime"`
	StoragePriority   int           `json:"storagePriority"`
	Usb               string   `json:"usb"`
	Usbs              []string `json:"usbs"`
	Cdrom             struct {
		Path           string      `json:"path"`
		Type           string      `json:"type"`
		Connected      bool        `json:"connected"`
		StartConnected bool        `json:"startConnected"`
		CifsDto        string `json:"cifsDto"`
		DataStore      string `json:"dataStore"`
	} `json:"cdrom"`
	Floppy struct {
		Path string `json:"path"`
	} `json:"floppy"`
	Disks []struct {
		ID        string      `json:"id"`
		Label     string      `json:"label"`
		ScsiID    string `json:"scsiId"`
		Enabled   bool        `json:"enabled"`
		WriteBps  int         `json:"writeBps"`
		ReadBps   int         `json:"readBps"`
		TotalBps  int         `json:"totalBps"`
		TotalIops int         `json:"totalIops"`
		WriteIops int         `json:"writeIops"`
		ReadIops  int         `json:"readIops"`
		Volume    struct {
			ID             string      `json:"id"`
			UUID           string      `json:"uuid"`
			Size           float64     `json:"size"`
			RealSize       float64     `json:"realSize"`
			Name           string      `json:"name"`
			FileName       string      `json:"fileName"`
			Offset         int         `json:"offset"`
			Shared         bool        `json:"shared"`
			DeleteModel    string      `json:"deleteModel"`
			VolumePolicy   string      `json:"volumePolicy"`
			Format         string      `json:"format"`
			BlockDeviceID  string `json:"blockDeviceId"`
			DiskType       string `json:"diskType"`
			DataStoreID    string      `json:"dataStoreId"`
			DataStoreName  string      `json:"dataStoreName"`
			DataStoreSize  float64     `json:"dataStoreSize"`
			FreeStorage    float64     `json:"freeStorage"`
			DataStoreType  string      `json:"dataStoreType"`
			VMName         string `json:"vmName"`
			VMStatus       string `json:"vmStatus"`
			Type           string `json:"type"`
			Description    string `json:"description"`
			Bootable       bool        `json:"bootable"`
			VolumeStatus   string      `json:"volumeStatus"`
			MountedHostIds string `json:"mountedHostIds"`
			Md5            string `json:"md5"`
			DataSize       int         `json:"dataSize"`
			OpenStackID    string `json:"openStackId"`
			VvSourceDto    string `json:"vvSourceDto"`
			FormatDisk     bool        `json:"formatDisk"`
			ToBeConverted  bool        `json:"toBeConverted"`
			RelatedVms     string `json:"relatedVms"`
			ClusterSize    int         `json:"clusterSize"`
		} `json:"volume"`
		BusModel        string  `json:"busModel"`
		Usage           float64 `json:"usage"`
		MonReadIops     float64 `json:"monReadIops"`
		MonWriteIops    float64 `json:"monWriteIops"`
		ReadThroughput  float64 `json:"readThroughput"`
		WriteThroughput float64 `json:"writeThroughput"`
		ReadWriteModel  string  `json:"readWriteModel"`
		EnableNativeIO  bool    `json:"enableNativeIO"`
		L2CacheSize     int     `json:"l2CacheSize"`
	} `json:"disks"`
	Nics []struct {
		ID              string      `json:"id"`
		AutoGenerated   bool        `json:"autoGenerated"`
		Name            string      `json:"name"`
		NolocalName     string      `json:"nolocalName"`
		InnerName       string `json:"innerName"`
		DevName         string      `json:"devName"`
		IP              string `json:"ip"`
		Netmask         string `json:"netmask"`
		Gateway         string `json:"gateway"`
		Mac             string      `json:"mac"`
		Model           string      `json:"model"`
		DeviceID        string      `json:"deviceId"`
		DeviceName      string      `json:"deviceName"`
		DeviceType      string      `json:"deviceType"`
		SwitchType      string      `json:"switchType"`
		VswitchID       string      `json:"vswitchId"`
		UplinkRate      int         `json:"uplinkRate"`
		UplinkBurst     int         `json:"uplinkBurst"`
		DownlinkRate    int         `json:"downlinkRate"`
		DownlinkBurst   int         `json:"downlinkBurst"`
		DownlinkQueue   string `json:"downlinkQueue"`
		Enable          bool        `json:"enable"`
		Status          string      `json:"status"`
		InboundRate     float64     `json:"inboundRate"`
		OutboundRate    float64     `json:"outboundRate"`
		ConnectStatus   bool        `json:"connectStatus"`
		VMName          string `json:"vmName"`
		VMID            string `json:"vmId"`
		VMStatus        string `json:"vmStatus"`
		VMTemplate      bool        `json:"vmTemplate"`
		NetworkName     string `json:"networkName"`
		NetworkVlan     string `json:"networkVlan"`
		VlanRange       string `json:"vlanRange"`
		NetworkID       string      `json:"networkId"`
		NetworkType     string `json:"networkType"`
		HostIP          string `json:"hostIp"`
		HostStatus      string `json:"hostStatus"`
		HostID          string `json:"hostId"`
		DirectObjName   string `json:"directObjName"`
		TotalOctets     float64     `json:"totalOctets"`
		TotalDropped    float64     `json:"totalDropped"`
		TotalPackets    float64     `json:"totalPackets"`
		TotalBytes      float64     `json:"totalBytes"`
		TotalErrors     float64     `json:"totalErrors"`
		WriteOctets     float64     `json:"writeOctets"`
		WriteDropped    float64     `json:"writeDropped"`
		WritePackets    float64     `json:"writePackets"`
		WriteBytes      float64     `json:"writeBytes"`
		WriteErrors     float64     `json:"writeErrors"`
		ReadOctets      float64     `json:"readOctets"`
		ReadDropped     float64     `json:"readDropped"`
		ReadPackets     float64     `json:"readPackets"`
		ReadBytes       float64     `json:"readBytes"`
		ReadErrors      float64     `json:"readErrors"`
		SecurityGroups  string `json:"securityGroups"`
		AdvancedNetIP   string `json:"advancedNetIp"`
		PortID          string `json:"portId"`
		OpenstackID     string `json:"openstackId"`
		BindIPEnable    bool        `json:"bindIpEnable"`
		BindIP          string `json:"bindIp"`
		PriorityEnabled bool        `json:"priorityEnabled"`
		NetPriority     string `json:"netPriority"`
		VMType          string      `json:"vmType"`
		SystemVMType    string `json:"systemVmType"`
		Dhcp            bool        `json:"dhcp"`
		DhcpIP          string `json:"dhcpIp"`
	} `json:"nics"`
	Gpus               []string `json:"gpus"`
	VMPcis             []string `json:"vmPcis"`
	ConfigLocation     string        `json:"configLocation"`
	HotplugEnabled     bool          `json:"hotplugEnabled"`
	VncPort            int           `json:"vncPort"`
	VncPasswd          string        `json:"vncPasswd"`
	VncSharePolicy     string        `json:"vncSharePolicy"`
	VcpuPin            string        `json:"vcpuPin"`
	VcpuPins           []string      `json:"vcpuPins"`
	CPUShares          int           `json:"cpuShares"`
	PanickPolicy       string        `json:"panickPolicy"`
	DataStoreID        string   `json:"dataStoreId"`
	SdsdomainID        string   `json:"sdsdomainId"`
	ClockModel         string        `json:"clockModel"`
	CPULimit           int           `json:"cpuLimit"`
	MemShares          int           `json:"memShares"`
	CPUReservation     int           `json:"cpuReservation"`
	MemReservation     int           `json:"memReservation"`
	LastBackup         string   `json:"lastBackup"`
	VMType             string        `json:"vmType"`
	SystemVMType       string   `json:"systemVmType"`
	MemBalloonEnabled  bool          `json:"memBalloonEnabled"`
	Completed          bool          `json:"completed"`
	GraphicsCardModel  string        `json:"graphicsCardModel"`
	GraphicsCardMemory int           `json:"graphicsCardMemory"`
	VMHostName         string   `json:"vmHostName"`
	DiskTotalSize      float64       `json:"diskTotalSize"`
	DiskUsedSize       float64       `json:"diskUsedSize"`
	DiskUsage          float64       `json:"diskUsage"`
	Tags               string   `json:"tags"`
	StartPriority      string        `json:"startPriority"`
	OwnerName          string        `json:"ownerName"`
}

func getVms(restReq getVmsReq) ([]Vm, rest.RestError) {
	rp, err := rest.NewRestProxy()
	vms := []Vm{}
	if err != nil {
		msg := fmt.Sprintf("cannot create REST client: %s", err.Error())
		return vms, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	stat, body, err := rp.Send("GET", "vms", restReq)
	fmt.Printf("get vms rsp:%v\n", stat)
	if err != nil {
		msg := fmt.Sprintf("get vms  failed: %s", err.Error())
		return vms, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	var getVmsR getVmsRsp
	if err := json.Unmarshal(body, &getVmsR); err != nil {
		msg := fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return vms, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	vms = getVmsR.Ttems
	klog.V(4).Infof("get vms: %s\n", vms)
	return vms, nil
}
func GetVmInfo(vmid string) (Vm, rest.RestError) {
	rp, err := rest.NewRestProxy()
	vminfo := Vm{}
	if err != nil {
		msg := fmt.Sprintf("cannot create REST client: %s", err.Error())
		return vminfo, rest.GetError(rest.RestRequestMalfunction, msg)
	}
	vmpath := "vms/" + vmid
	stat, body, err := rp.Send("GET", vmpath, nil)
	fmt.Printf("get vminfo rsp:%v\n", stat)
	if err != nil {
		msg := fmt.Sprintf("get vminfo  failed: %s", err.Error())
		return vminfo, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	if err := json.Unmarshal(body, &vminfo); err != nil {
		msg := fmt.Sprintf("json unmarshal failed: %s", err.Error())
		return vminfo, rest.GetError(rest.RestRequestMalfunction, msg)
	}

	klog.V(4).Infof("get vminfo: %s\n", vminfo)
	return vminfo, nil
}

func GetVMsList() ([]Vm, rest.RestError) {
	getVmsReq := getVmsReq{
		1000,
		1,
		"",
		"desc",
	}

	vms, err := getVms(getVmsReq)
	if err != nil {
		klog.V(4).Infof("get  vmlist failed  with args %+v", vms)
	}
	return vms, err
}
