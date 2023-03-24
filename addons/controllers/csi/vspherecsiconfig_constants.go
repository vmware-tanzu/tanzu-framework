// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers ...
package controllers

import (
	"time"
)

// supported modes for VSphereCSIConfig
const (
	VSphereCSINonParavirtualMode = "vsphereCSI"
	VSphereCSIParavirtualMode    = "vsphereParavirtualCSI"
)

// cluster variable names required by VSphereCSIConfig to derive values
const (
	VSphereNetworkVarName           = "VSPHERE_NETWORK"
	VSphereRegionVarName            = "VSPHERE_REGION"
	VSphereZoneVarName              = "VSPHERE_ZONE"
	VSphereVersionVarName           = "VSPHERE_VERSION"
	VSphereTLSThumbprintVarName     = "VSPHERE_TLS_THUMBPRINT"
	IsWindowsWorkloadClusterVarName = "IS_WINDOWS_WORKLOAD_CLUSTER"
	ControllerRequeueDelay          = 20 * time.Second
)

const (
	VSphereCSINamespace                 = "kube-system"
	VSphereCSIProvisionTimeout          = "300s"
	VSphereCSIAttachTimeout             = "300s"
	VSphereCSIResizerTimeout            = "300s"
	VSphereCSIMinDeploymentReplicas     = 1
	VSphereCSIMaxDeploymentReplicas     = 3
	VSphereCSIFeatureStateNamespace     = VSphereSystemCSINamepace
	VSphereCSIFeatureStateConfigMapName = "csi-feature-states"
	LegacyZoneName                      = "vmware-system-legacy"
)

const (
	VSphereSystemCSINamepace                = "vmware-system-csi"
	DefaultSupervisorMasterEndpointHostname = "supervisor.default.svc"
	DefaultSupervisorMasterPort             = 6443
)
