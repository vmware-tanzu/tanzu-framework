// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo ...
package buildinfo

var (
	// Version is the current git version version of tkg, set with the go linker's -X flag.
	Version string

	// Commit is the actual commit that is being built, set with the go linker's -X flag.
	Commit string
)
