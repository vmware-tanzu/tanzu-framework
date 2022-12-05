// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// DefaultFeatureFlagsForClusterPlugin is used to populate default feature-flags for the cluster plugin
var (
	DefaultFeatureFlagsForClusterPlugin = map[string]bool{
		constants.FeatureFlagClusterDualStackIPv4Primary: false,
		constants.FeatureFlagClusterDualStackIPv6Primary: false,
		constants.FeatureFlagClusterCustomNameservers:    false,
		constants.FeatureFlagAllowLegacyCluster:          false,
		constants.FeatureFlagPackageBasedCC:              true,
	}
)
