// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

import "time"

// PackageOptions includes fields for package operations
type PackageOptions struct {
	Available              string
	ClusterRoleName        string
	ClusterRoleBindingName string
	Namespace              string
	PackageName            string
	PkgInstallName         string
	SecretName             string
	ServiceAccountName     string
	ValuesFile             string
	Version                string
	PollInterval           time.Duration
	PollTimeout            time.Duration
	AllNamespaces          bool
	CreateNamespace        bool
	Install                bool
	Wait                   bool
	SkipPrompt             bool
	Labels                 map[string]string
}

// NewPackageOptions instantiates PackageOptions
func NewPackageOptions() *PackageOptions {
	return &PackageOptions{}
}
