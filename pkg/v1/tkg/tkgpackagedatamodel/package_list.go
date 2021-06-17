// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// PackageListOptions includes fields for package list
type PackageListOptions struct {
	Available     bool
	AllNamespaces bool
	ListInstalled bool
	Namespace     string
	PackageName   string
	KubeConfig    string
}

// NewPackageListOptions instantiates PackageListOptions
func NewPackageListOptions() *PackageListOptions {
	return &PackageListOptions{}
}
