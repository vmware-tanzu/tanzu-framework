// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo ...
// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo
package buildinfo

import "github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"

var (
	// Version is the current git version of tkg, set with the go linker's -X flag.
	// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Version
	Version = buildinfo.Version

	// Commit is the actual commit that is being built, set with the go linker's -X flag.
	// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.SHA
	Commit = buildinfo.SHA
)
