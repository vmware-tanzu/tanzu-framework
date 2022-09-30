// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

func postInstallHook() error {
	// Configure default feature-flags for cluster plugin
	return config.ConfigureDefaultFeatureFlagsIfMissing(DefaultFeatureFlagsForClusterPlugin)
}
