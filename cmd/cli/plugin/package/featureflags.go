// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

const (
	// Package Plugin Kctrl Command Tree determines whether to use the command tree from kctrl. Setting feature flag to
	// true will allow to use the package command tree from kctrl for package plugin
	FeatureFlagPackagePluginKctrlCommandTree = "features.package.kctrl-package-command-tree"
)

// DefaultFeatureFlagsForPackagePlugin is used to populate default feature-flags for the package plugin
var (
	DefaultFeatureFlagsForPackagePlugin = map[string]bool{
		FeatureFlagPackagePluginKctrlCommandTree: true,
	}
)
