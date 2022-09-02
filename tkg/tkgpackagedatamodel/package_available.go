// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// PackageAvailableOptions includes fields for package available
type PackageAvailableOptions struct {
	Namespace                 string
	PackageName               string
	AllNamespaces             bool
	ValuesSchema              bool
	GenerateDefaultValuesFile bool
}

// NewPackageAvailableOptions instantiates PackageAvailableOptions
func NewPackageAvailableOptions() *PackageAvailableOptions {
	return &PackageAvailableOptions{}
}
