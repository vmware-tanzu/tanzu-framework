// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo ...
package buildinfo

var (
	// Date is the date the binary was built.
	// Set by go build -ldflags "-X" flag
	Date string

	// SHA is the git commit SHA the binary was built with.
	// Set by go build -ldflags "-X" flag
	SHA string

	// Version is the version the binary was built with.
	// Set by go build -ldflags "-X" flag
	Version string

	// Commit is the actual commit that is being built, set with the go linker's -X flag.
	// Deprecated: use github.com/vmware-tanzu/tanzu-framework/tkg/buildinfo.SHA
	Commit = SHA
)
