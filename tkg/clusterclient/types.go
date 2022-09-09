// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"

// constants related to tkg operation info that are used for adding
// annotations to the clusters
const (
	TKGOperationInfoKey                  = "TKGOperationInfo"
	TKGOperationLastObservedTimestampKey = "TKGOperationLastObservedTimestamp"
	TKGVersionKey                        = "TKGVERSION"
	CAPAControllerNamespace              = "capa-system"
	CAPACredentialsSecretName            = "capa-manager-bootstrap-credentials"
	CAPAControllerDeploymentName         = "capa-controller-manager"
)

// Operation type constants
const (
	OperationTypeUpgrade = "Upgrade"
	OperationTypeCreate  = "Create"
)

const (
	kubeProxyKey                     = "kube-proxy"
	calicoNodeKey                    = "calico-node"
	calicoKubeControllerKey          = "calico-kube-controllers"
	kubeadmConfigKey                 = "kubeadm-config"
	clusterConfigurationKey          = "ClusterConfiguration"
	kappControllerKey                = "kapp-controller"
	kappControllServiceAccount       = "kapp-controller-sa"
	kappControllerClusterRole        = "kapp-controller-cluster-role"
	kappControllerClusterRoleBinding = "kapp-controller-cluster-role-binding"
	kappControllerOldNamespace       = "vmware-system-tmc"
	kappControllerNamespace          = "tkg-system"
)

// OperationStatus describes current status of running operation
// this struct is used for patching cluster object with the last
// invoked operation information.
// This information combined with TKGOperationLastObservedTimestamp
// will be used for determining stalled state of a cluster
type OperationStatus struct {
	Operation               string `json:"Operation"`
	OperationStartTimestamp string `json:"OperationStartTimestamp"`
	OperationTimeout        int    `json:"OperationTimeout"`
}

// VerificationClientFactory clusterclient verification factory
// implements functions regarding verification which can be replaced with
// fake implementation for unit testing
type VerificationClientFactory struct {
	VerifyKubernetesUpgradeFunc func(clusterStatusInfo *ClusterStatusInfo, newK8sVersion string) error
	GetVCClientAndDataCenter    func(clusterName, clusterNamespace, vsphereMachineTemplateObjectName string) (vc.Client, string, error)
}
