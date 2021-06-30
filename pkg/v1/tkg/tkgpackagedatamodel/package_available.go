// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// PackageAvailableOptions includes fields for package available
type PackageAvailableOptions struct {
	KubeConfig    string
	Namespace     string
	PackageName   string
	AllNamespaces bool
	ValuesSchema  bool
}

// NewPackageAvailableOptions instantiates PackageAvailableOptions
func NewPackageAvailableOptions() *PackageAvailableOptions {
	return &PackageAvailableOptions{}
}
