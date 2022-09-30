// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

func postInstallHook() error {
	// Configure default feature-flags for management-cluster plugin
	err1 := config.ConfigureDefaultFeatureFlagsIfMissing(DefaultFeatureFlagsForManagementClusterPlugin)

	// Creating a tkgctl client in turn initializes the TKG configuration in the tanzu config directory.
	forceUpdateTKGCompatibilityImage := true
	_, err2 := newTKGCtlClient(forceUpdateTKGCompatibilityImage)

	return kerrors.NewAggregate([]error{err1, err2})
}
