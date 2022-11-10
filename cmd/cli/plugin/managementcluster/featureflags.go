// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// DefaultFeatureFlagsForManagementClusterPlugin is used to populate default feature-flags for the management-cluster plugin
var (
	DefaultFeatureFlagsForManagementClusterPlugin = map[string]bool{
		"features.management-cluster.import":                       false,
		"features.management-cluster.export-from-confirm":          true,
		"features.management-cluster.standalone-cluster-mode":      false,
		constants.FeatureFlagManagementClusterDualStackIPv4Primary: false,
		constants.FeatureFlagManagementClusterDualStackIPv6Primary: false,
		constants.FeatureFlagManagementClusterCustomNameservers:    false,
		constants.FeatureFlagAwsInstanceTypesExcludeArm:            true,
	}
)
