// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// ConfigFilePermissions defines the permissions of the config file
const (
	ConfigFilePermissions       = 0o600
	DefaultDirectoryPermissions = 0o700
)

// File name related constants
const (
	TKGRegistryCertFile                    = "registry_certs"
	TKGRegistryTrustedRootCAFileForWindows = ".registry_trusted_root_certs_win"
)
