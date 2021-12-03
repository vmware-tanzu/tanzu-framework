// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package common defines generic constants and structs
package common

// Plugin status and scope constants
const (
	PluginStatusInstalled       = "installed"
	PluginStatusNotInstalled    = "not installed"
	PluginStatusUpdateAvailable = "update available"
	PluginScopeStandalone       = "Standalone"
	PluginScopeContext          = "Context"
)

// DiscoveryType constants
const (
	DiscoveryTypeOCI        = "oci"
	DiscoveryTypeLocal      = "local"
	DiscoveryTypeGCP        = "gcp"
	DiscoveryTypeKubernetes = "kubernetes"
	DiscoveryTypeREST       = "rest"
)

// DistributionType constants
const (
	DistributionTypeOCI   = "oci"
	DistributionTypeLocal = "local"
)
