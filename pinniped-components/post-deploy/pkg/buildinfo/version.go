// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo holds build-time information.
// This is a separate package so that other packages can import it without
// worrying about introducing circular dependencies.
package buildinfo

// Version is the current version, set by the go linker's -X flag at build time
var Version = "v0.0.1"

// GitSHA is the actual commit that is being built, set by the go linker's -X flag at build time
var GitSHA string
