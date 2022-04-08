// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

// supported modes for VSphereCPIConfig
const (
	VsphereCPINonParavirtualMode = "vsphereCPI"
	VSphereCPIParavirtualMode    = "vsphereParavirtualCPI"
)

// cluster variable names required by VSphereCPIConfig to derive values
const (
	NsxtPodRoutingEnabledVarName  = "NSXT_POD_ROUTING_ENABLED"
	NsxtRouterPathVarName         = "NSXT_ROUTER_PATH"
	ClusterCIDRVarName            = "CLUSTER_CIDR"
	NsxtUsernameVarName           = "NSXT_USERNAME"
	NsxtPasswordVarName           = "NSXT_PASSWORD"
	NsxtManagerHostVarName        = "NSXT_MANAGER_HOST"
	NsxtAllowUnverifiedSSLVarName = "NSXT_ALLOW_UNVERIFIED_SSL"
	NsxtRemoteAuthVarName         = "NSXT_REMOTE_AUTH"
	NsxtVmcAccessTokenVarName     = "NSXT_VMC_ACCESS_TOKEN" // nolint:gosec
	NsxtVmcAuthHostVarName        = "NSXT_VMC_AUTH_HOST"
	NsxtClientCertKeyDataVarName  = "NSXT_CLIENT_CERT_KEY_DATA"
	NsxtClientCertDataVarName     = "NSXT_CLIENT_CERT_DATA"
	NsxtRootCADataB64VarName      = "NSXT_ROOT_CA_DATA_B64"
	NsxtSecretNameVarName         = "NSXT_SECRET_NAME"      // nolint:gosec
	NsxtSecretNamespaceVarName    = "NSXT_SECRET_NAMESPACE" // nolint:gosec
)

// secret in the target cluster that contains the generated service account
// token.
const (
	ProviderServiceAccountSecretName      = "cloud-provider-creds"
	ProviderServiceAccountSecretNamespace = "vmware-system-cloud-provider" // nolint:gosec
)

// TODO: make these constants accessible to other controllers (for example csi) https://github.com/vmware-tanzu/tanzu-framework/issues/2086
// constants used for deriving supervisor API server endpoint
const (
	SupervisorEndpointHostname = "supervisor.default.svc"
	SupervisorEndpointPort     = 6443

	SupervisorLoadBalancerSvcNamespace = "kube-system"
	SupervisorLoadBalancerSvcName      = "kube-apiserver-lb-svc"

	// ConfigMapClusterInfo defines the name for the ConfigMap where the information how to connect and trust the cluster exist
	ConfigMapClusterInfo = "cluster-info"

	// KubeConfigKey defines at which key in the Data object of the ConfigMap the KubeConfig object is stored
	KubeConfigKey = "kubeconfig"
)
