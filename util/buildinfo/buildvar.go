// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo holds global vars set at build time to provide information about the build.
// This package SHOULD NOT import other packages -- to avoid dependency cycles.
package buildinfo

var (
	// Date is the date the binary was built.
	// Set by go build -ldflags "-X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.Date=...'"
	Date string

	// SHA is the git commit SHA the binary was built with.
	// Set by go build -ldflags "-X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.SHA=...'"
	SHA string

	// Version is the version the binary was built with.
	// Set by go build -ldflags "-X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.Version=...'"
	Version string
)
