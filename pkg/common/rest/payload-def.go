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

type TaskRsp struct {
	TaskId string `json:"taskId"`
}

type TaskEvents struct {
	Id          string `json:"id"`
	UserName    string `json:"userName"`
	Description string `json:"description"`
	Level       string `json:"level"`
	TargetId    string `json:"targetId"`
	TargetName  string `json:"targetName"`
	TargetType  string `json:"targetType"`
	CreateTime  string `json:"createTime"`
	Events      string `json:"events"`
}

type TaskInfoRsp struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	Detail     string        `json:"detail"`
	State      string        `json:"state"`
	StartTime  string        `json:"startTime"`
	EndTime    string        `json:"endTime"`
	ActorName  string        `json:"actorName"`
	Error      string        `json:"error"`
	Cancelable bool          `json:"cancelable"`
	TargetName string        `json:"targetName"`
	TargetId   string        `json:"targetId"`
	TargetType string        `json:"targetType"`
	Events     []interface{} `json:"events"`
	ProcessId  string        `json:"processId"`
	Progress   int           `json:"progress"`
}

type VolumeListRsp struct {
    Items       []VolumeInfoRsp `json:"items"`
}

type VolumeInfoRsp struct {
    Id              string      `json:"id"`             //"8a878bda6f6f3ca4016f6f64fea9008f"
    Uuid            string      `json:"uuid"`           //"98aa6aae-4da3-4d95-90d5-0d9b9ef185a0"
    Size            float64     `json:"size"`           //40.0
    RealSize        float64     `json:"realSize"`       //6.64
    Name            string      `json:"name"`           //"11.87_licenseserver_Disk1"
    FileName        string      `json:"fileName"`       //"/datastore/4e76e7ce-14ea-4865-8a99-05dd32505bb4/98aa6aae-4da3-4d95-90d5-0d9b9ef185a0"
    Offset          int         `json:"offset"`         //5242880
    Shared          bool        `json:"shared"`         //false
    DeleteModel     string      `json:"deleteModel"`    //"RESET_ZERO_AFTER_DELETE"
    VolumePolicy    string      `json:"volumePolicy"`   //"THIN"
    Format          string      `json:"format"`         //"QCOW2"
    BlockDeviceId   string      `json:"blockDeviceId"`  //null
    DiskType        interface{} `json:"diskType"`       //null
    DataStoreId     string      `json:"dataStoreId"`    //"8a878bda6f6f3ca4016f6f458b50003c"
    DataStoreName   string      `json:"dataStoreName"`  //"local"
    DataStoreSize   float64     `json:"dataStoreSize"`  //1471.17
    FreeStorage     float64     `json:"freeStorage"`    //0.0
    DataStoreType   interface{} `json:"dataStoreType"`  //"LOCAL"
    VmName          string      `json:"vmName"`         //null
    VmStatus        interface{} `json:"vmStatus"`       //null
    Type            interface{} `json:"type"`           //null
    Description     interface{} `json:"description"`    //null
    Bootable        bool        `json:"bootable"`       //true
    VolumeStatus    string      `json:"volumeStatus"`   //"UNAVAILABLE"
    MountedHostIds  interface{} `json:"mountedHostIds"` //null
    Md5             interface{} `json:"md5"`            //null
    DataSize        int         `json:"dataSize"`       //0
    OpenStackId     interface{} `json:"openStackId"`    //null
    VvSourceDto     interface{} `json:"vvSourceDto"`    //null
    FormatDisk      bool        `json:"formatDisk"`     //false
    ToBeConverted   bool        `json:"toBeConverted"`  //false
    RelatedVms      interface{} `json:"relatedVms"`     //null
    ClusterSize     int         `json:"clusterSize"`    //262144
}

type VmDiskInfo struct {
    Id              string      `json:"id"`             //"8a878bda6f6f3ca4016f6f6504670092"
    Label           string      `json:"label"`          //"/dev/vda"
    ScsiId          string      `json:"scsiId"`         //null
    Enabled         bool        `json:"enabled"`        //false
    WriteBps        int         `json:"writeBps"`       //0
    ReadBps         int         `json:"readBps"`        //0
    TotalBps        int         `json:"totalBps"`       //0
    TotalIops       int         `json:"totalIops"`      //0
    WriteIops       int         `json:"writeIops"`      //0
    ReadIops        int         `json:"readIops"`       //0
    Volume          VolumeInfoRsp`json:"volume"`         //{}
    BusModel        string      `json:"busModel"`       //"VIRTIO"
    Usage           float64     `json:"usage"`          //0.0
    MonReadIops     float64     `json:"monReadIops"`    //0.0
    MonWriteIops    float64     `json:"monWriteIops"`   //0.0
    ReadThroughput  float64     `json:"readThroughput"` //0.0
    WriteThroughput float64     `json:"writeThroughput"` //0.0
    ReadWriteModel  string      `json:"readWriteModel"` //"NONE"
    EnableNativeIO  bool        `json:"enableNativeIO"` //false
    L2CacheSize     int         `json:"l2CacheSize"`    //0
}

type VmNicInfo struct {
    Id              string      `json:"id"`             //"8a878bda6f7012c7016f70b41f4100aa"
    AutoGenerated   bool        `json:"autoGenerated"`  //true
    Name            string      `json:"name"`           //"网卡1"
    NolocalName     string      `json:"nolocalName"`    //"1"
    InnerName       interface{} `json:"innerName"`      //null
    DevName         string      `json:"devName"`        //"vnb1ecb8a60"
    Ip              interface{} `json:"ip"`             //null
    Netmask         interface{} `json:"netmask"`        //null
    Gateway         interface{} `json:"gateway"`        //null
    Mac             string      `json:"mac"`            //"00:16:3e:ed:33:bc"
    Model           string      `json:"model"`          //"VIRTIO"
    DeviceId        interface{} `json:"deviceId"`       //null
    DeviceName      interface{} `json:"deviceName"`     //null
    DeviceType      string      `json:"deviceType"`     //"NETWORK"
    SwitchType      string      `json:"switchType"`     //"NORMALSWITCH"
    VswitchId       interface{} `json:"vswitchId"`      //null
    UplinkRate      int         `json:"uplinkRate"`     //0
    UplinkBurst     int         `json:"uplinkBurst"`    //0
    DownlinkRate    int         `json:"downlinkRate"`   //0
    DownlinkBurst   int         `json:"downlinkBurst"`  //0
    DownlinkQueue   interface{} `json:"downlinkQueue"`  //null
    Enable          bool        `json:"enable"`         //false
    Status          interface{} `json:"status"`         //null
    InboundRate     float64     `json:"inboundRate"`    //0.0
    OutboundRate    float64     `json:"outboundRate"`   //0.0
    ConnectStatus   bool        `json:"connectStatus"`  //false
    VmName          interface{} `json:"vmName"`         //null
    VmId            interface{} `json:"vmId"`           //null
    VmStatus        interface{} `json:"vmStatus"`       //null
    VmTemplate      bool        `json:"vmTemplate"`     //false
    NetworkName     interface{} `json:"networkName"`    //null
    NetworkVlan     interface{} `json:"networkVlan"`    //null
    VlanRange       interface{} `json:"vlanRange"`      //null
    NetworkId       string      `json:"networkId"`      //"8a878bda6f6f3ca4016f6f4409bf002e"
    NetworkType     interface{} `json:"networkType"`    //null
    HostIp          interface{} `json:"hostIp"`         //null
    HostStatus      interface{} `json:"hostStatus"`     //null
    HostId          interface{} `json:"hostId"`         //null
    DirectObjName   interface{} `json:"directObjName"`  //null
    TotalOctets     float64     `json:"totalOctets"`    //0.0
    TotalDropped    float64     `json:"totalDropped"`   //0.0
    TotalPackets    float64     `json:"totalPackets"`   //0.0
    TotalBytes      float64     `json:"totalBytes"`     //0.0
    TotalErrors     float64     `json:"totalErrors"`    //0.0
    WriteOctets     float64     `json:"writeOctets"`    //0.0
    WriteDropped    float64     `json:"writeDropped"`   //0.0
    WritePackets    float64     `json:"writePackets"`   //0.0
    WriteBytes      float64     `json:"writeBytes"`     //0.0
    WriteErrors     float64     `json:"writeErrors"`    //0.0
    ReadOctets      float64     `json:"readOctets"`     //0.0
    ReadDropped     float64     `json:"readDropped"`    //0.0
    ReadPackets     float64     `json:"readPackets"`    //0.0
    ReadBytes       float64     `json:"readBytes"`      //0.0
    ReadErrors      float64     `json:"readErrors"`     //0.0
    SecurityGroups  interface{} `json:"securityGroups"` //null
    AdvancedNetIp   interface{} `json:"advancedNetIp"`  //null
    PortId          interface{} `json:"portId"`         //null
    OpenstackId     interface{} `json:"openstackId"`    //null
    BindIpEnable    bool        `json:"bindIpEnable"`   //false
    BindIp          interface{} `json:"bindIp"`         //null
    PriorityEnabled bool        `json:"priorityEnabled"`//false
    NetPriority     interface{} `json:"netPriority"`    //null
    VmType          string      `json:"vmType"`         //"GENERAL_VM"
    SystemVmType    interface{} `json:"systemVmType"`   //null
    Dhcp            bool        `json:"dhcp"`           //false
    DhcpIp          interface{} `json:"dhcpIp"`         //null
}

type VmInfoRsp struct {
    Id              string      `json:"id"`             //"8a878bda6f6f3ca4016f6f57d27b0088"
    Name            string      `json:"name"`           //"test"
    State           string      `json:"state"`          //"UNKNOWN"
    Status          string      `json:"status"`         //"STARTED"
    HostId          string      `json:"hostId"`         //"792b1e3f-8be6-43de-b8c7-27a0327dcd97"
    HostName        string      `json:"hostName"`       //"allinone-88945-43735"
    HostIp          string      `json:"hostIp"`         //"10.7.11.90"
    HostStatus      string      `json:"hostStatus"`     //"CONNECTED"
    HostMemory      float64     `json:"hostMemory"`     //0.0
    DataCenterId    string      `json:"dataCenterId"`   //"3f0094542ebb11eaa2691a130ac12531"
    HaEnabled       bool        `json:"haEnabled"`      //true
    RouterFlag      bool        `json:"routerFlag"`     //false
    Migratable      bool        `json:"migratable"`     //true
    ToolsInstalled  bool        `json:"toolsInstalled"` //false
    ToolsVersion    string      `json:"toolsVersion"`   //null
    ToolsType       string      `json:"toolsType"`      //"VMTOOLS"
    Description     string      `json:"description"`    //""
    HaMaxLimit      int         `json:"haMaxLimit"`     //0
    Template        bool        `json:"template"`       //false
    Initialized     bool        `json:"initialized"`    //false
    GuestosLabel    string      `json:"guestosLabel"`   //"CentOS 7.6(1810) 64bit"
    GuestosType     string      `json:"guestosType"`    //"CentOS"
    GuestOsInfo     interface{} `json:"guestOsInfo"`    //{}
    InnerName       string      `json:"innerName"`      //"i-000004"
    Uuid            string      `json:"uuid"`           //"70e36ec7-599d-4b03-83c0-6e2212a06da4"
    MaxMemory       int         `json:"maxMemory"`      //1047552
    Memory          int         `json:"memory"`         //8192
    MemoryUsage     float64     `json:"memoryUsage"`    //0.0
    MemHotplugEnabled   bool    `json:"memHotplugEnabled"` //false
    CpuNum          int         `json:"cpuNum"`         //8
    CpuSocket       int         `json:"cpuSocket"`      //4
    CpuCore         int         `json:"cpuCore"`        //2
    CpuUsage        float64     `json:"cpuUsage"`       //0.4
    MaxCpuNum       int         `json:"maxCpuNum"`      //8
    CpuHotplugEnabled   bool    `json:"cpuHotplugEnabled"` //false
    CpuModelType    string      `json:"cpuModelType"`   //"SELF_ADAPTING"
    CpuModelEnabled bool        `json:"cpuModelEnabled"`    //true
    RunningTime     float64     `json:"runningTime"`    //1005.0
    Boot            string      `json:"boot"`           //"HD"
    BootMode        string      `json:"bootMode"`       //"BIOS"
    SplashTime      int         `json:"splashTime"`     //0
    StoragePriority int         `json:"storagePriority"`// 0
    Usb             interface{} `json:"usb"`            //null
    Usbs            interface{} `json:"usbs"`           //[]
    Cdrom           interface{} `json:"cdrom"`          //{}
    Floppy          interface{} `json:"floppy"`         //{}
    Disks           []VmDiskInfo`json:"disks"`          //[]
    Nics            interface{} `json:"nics"`           //[]
    Gpus            interface{} `json:"gpus"`           //[]
    VmPcis          interface{} `json:"vmPcis"`         //[]
    ConfigLocation  string      `json:"configLocation"` //"i-000004.conf"
    HotplugEnabled  bool        `json:"hotplugEnabled"` //false
    VncPort         int         `json:"vncPort"`        //5904
    VncPasswd       string      `json:"vncPasswd"`      //"PE6e6ohBLB6GgwzbJQMkeQ=="
    VncSharePolicy  string      `json:"vncSharePolicy"` //"FORCE_SHARED"
    VcpuPin         string      `json:"vcpuPin"`        //"all"
    VcpuPins        interface{} `json:"vcpuPins"`       //[]
	CpuShares       int         `json:"cpuShares"`      //1024
    PanickPolicy    string      `json:"panickPolicy"`   //"NO_ACTION"
    DataStoreId     string      `json:"dataStoreId"`    //null
    SdsdomainId     string      `json:"sdsdomainId"`    //null
    ClockModel      string      `json:"clockModel"`     //"UTC"
    CpuLimit        int         `json:"cpuLimit"`       //-1
    MemShares       int         `json:"memShares"`      //1024
    CpuReservation  int         `json:"cpuReservation"` //0
    MemReservation  int         `json:"memReservation"` //0
    LastBackup      interface{} `json:"lastBackup"`     //null
    VmType          string      `json:"vmType"`         //"GENERAL_VM"
    SystemVmType    interface{} `json:"systemVmType"`   //null
    MemBalloonEnabled   bool    `json:"memBalloonEnabled"`  //false
    Completed       bool        `json:"completed"`      //false
    GraphicsCardModel   string  `json:"graphicsCardModel"`  //"CIRRUS"
    GraphicsCardMemory  int     `json:"graphicsCardMemory"` //16384
    VmHostName      interface{} `json:"vmHostName"`     //null
    DiskTotalSize   float64     `json:"diskTotalSize"`  //0.0
    DiskUsedSize    float64     `json:"diskUsedSize"`   //0.0
    DiskUsage       float64     `json:"diskUsage"`      //0.0
    Tags            interface{} `json:"tags"`           //null
    StartPriority   string      `json:"startPriority"`  //"DEFAULT"
    OwnerName       string      `json:"ownerName"`      //"admin@internal"
}

type VmListRsp struct {
    Items   []VmInfoRsp     `json:"items"`
}
