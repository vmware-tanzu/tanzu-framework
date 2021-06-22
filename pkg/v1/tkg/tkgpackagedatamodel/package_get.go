// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// PackageGetOptions includes fields for package get
type PackageGetOptions struct {
	Available   string
	Version     string
	ValuesFile  string
	Namespace   string
	PackageName string
	KubeConfig  string
}

// NewPackageGetOptions instantiates PackageGetOptions
func NewPackageGetOptions() *PackageGetOptions {
	return &PackageGetOptions{}
}
