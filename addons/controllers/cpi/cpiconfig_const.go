// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

// supported modes for CPIConfig
const (
	CPINonParavirtualMode = "vsphereCPI"
	CPIParavirtualMode    = "vsphereParavirtualCPI"
)

// cluster variable names required by CPIConfig to derive values
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
