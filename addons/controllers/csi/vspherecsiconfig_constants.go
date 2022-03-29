// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package controllers

// supported modes for VSphereCSIConfig
const (
	VSphereCSINonParavirtualMode = "vsphereCSI"
	VSphereCSIParavirtualMode    = "vsphereParavirtualCSI"
)

// cluster variable names required by VSphereCSIConfig to derive values
const (
	NamespaceVarName                = "NAMESPACE"
	ClusterNameVarName              = "CLUSTER_NAME"
	VSphereServerVarName            = "VSPHERE_SERVER"
	VSphereNetworkVarName           = "VSPHERE_NETWORK"
	VSphereDatacenterVarName        = "VSPHERE_DATACENTER"
	VSphereRegionVarName            = "VSPHERE_REGION"
	VSphereUsernameVarName          = "VSPHERE_USERNAME"
	VSpherePasswordVarName          = "VSPHERE_PASSWORD"
	VSphereZoneVarName              = "VSPHERE_ZONE"
	VSphereVersionVarName           = "VSPHERE_VERSION"
	VSphereTLSThumbprintVarName     = "VSPHERE_TLS_THUMBPRINT"
	IsWindowsWorkloadClusterVarName = "IS_WINDOWS_WORKLOAD_CLUSTER"
)
