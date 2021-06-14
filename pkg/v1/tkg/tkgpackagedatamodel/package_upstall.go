// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

import "time"

// PackageOptions includes fields for package install/update
type PackageOptions struct {
	ClusterRoleName      string
	InstalledPkgName     string
	KubeConfig           string
	Namespace            string
	PackageName          string
	ServiceAccountName   string
	ValuesFile           string
	Version              string
	PollInterval         time.Duration
	PollTimeout          time.Duration
	CreateNamespace      bool
	CreateSecret         bool
	CreateServiceAccount bool
	Wait                 bool
}

// NewPackageOptions instantiates PackageOptions
func NewPackageOptions() *PackageOptions {
	return &PackageOptions{}
}
