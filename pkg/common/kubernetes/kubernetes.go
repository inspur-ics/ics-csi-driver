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

package kubernetes

import (
	"ics-csi-driver/pkg/csi/service/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"strings"
)

// NewClient creates a newk8s client based on a service account
func NewClient() (clientset.Interface, error) {

	var config *restclient.Config
	var err error
	klog.V(2).Info("k8s client using in-cluster config")
	config, err = restclient.InClusterConfig()
	if err != nil {
		klog.Errorf("InClusterConfig failed %q", err)
		return nil, err
	}

	return clientset.NewForConfig(config)
}

// CreateKubernetesClientFromConfig creaates a newk8s client from given kubeConfig file
func CreateKubernetesClientFromConfig(kubeConfigPath string) (clientset.Interface, error) {

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetNodeVMUUID returns VM UUID set by CCM on the Kubernetes Node
func GetNodeVMUUID(k8sclient clientset.Interface, nodeName string) (string, error) {
	var k8sNodeUUID string
	klog.V(2).Infof("GetNodeVMUUID called for the node: %q", nodeName)
	node, err := k8sclient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get kubernetes node with the name: %q. Err: %v", nodeName, err)
		return "", err
	}

	if node.Spec.ProviderID == "" {
		klog.Warningf("node %v ProviderID is empty", nodeName)
		k8sNodeUUID = strings.ToLower(node.Status.NodeInfo.SystemUUID)
	} else {
		k8sNodeUUID = common.GetUUIDFromProviderID(node.Spec.ProviderID)
	}

	klog.V(2).Infof("Retrieved node UUID: %q for the node: %q", k8sNodeUUID, nodeName)
	return k8sNodeUUID, nil
}
