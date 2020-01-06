package ics
import (
	"k8s.io/klog"
)
func TestInfo() {
	klog.V(4).Infof("my test ：***********************************************")
	vmlist,_:=GetVMsList()
	klog.V(4).Infof("get vmlist: %s\n", vmlist)
	vmid:=vmlist[0].ID
	klog.V(4).Infof("get vmid: %s\n", vmid)
	vminfo,_:=GetVmInfo(vmid)
	klog.V(4).Infof("get INFO: %s\n", vminfo)

	storagelist,_:=GetStoragesList()
	klog.V(4).Infof("get storagelist: %s\n", storagelist)
	storageid:=storagelist[0].ID
	klog.V(4).Infof("get storageid: %s\n", storageid)
	storageinfo,_:=GetStorageInfo(storageid)
	klog.V(4).Infof("get storageinfo: %s\n", storageinfo)
}

