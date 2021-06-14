// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

import "time"

// PackageUninstallOptions includes fields for package uninstall
type PackageUninstallOptions struct {
	InstalledPkgName string
	KubeConfig       string
	Namespace        string
	PollInterval     time.Duration
	PollTimeout      time.Duration
}

// NewPackageUninstallOptions instantiates PackageUninstallOptions
func NewPackageUninstallOptions() *PackageUninstallOptions {
	return &PackageUninstallOptions{}
}
