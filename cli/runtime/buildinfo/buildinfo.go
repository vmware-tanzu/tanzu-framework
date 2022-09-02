// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo holds global variables set at build time to provide information about the plugin build.
package buildinfo

var (
	// Date is the date the plugin binary was built.
	// It should be set using:
	//  go build --ldflags "-X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo.Date=...'"
	Date string

	// SHA is the git commit SHA the plugin binary was built with.
	// Is should be set using:
	//  go build --ldflags "-X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo.SHA=...'"
	SHA string

	// Version is the version of the plugin built.
	// It should be set using:
	//  go build --ldflags "-X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo.Version=...'"
	Version string
)
