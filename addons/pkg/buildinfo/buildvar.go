// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo holds global vars set at build time to provide information about the build.
// This package SHOULD NOT import other packages -- to avoid dependency cycles.
package buildinfo

// This package provides vars that are set at build time for the addons manager binary. This has been copied from
// tanzu-framework/pkg/v1/buildinfo to avoid dependency on tanzu-framework go module

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
)
