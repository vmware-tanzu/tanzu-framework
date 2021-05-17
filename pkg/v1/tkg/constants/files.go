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
	LocalProvidersFolderName  = "providers"
	LocalProvidersZipFileName = "providers.zip"
	LocalTanzuFileLock        = ".tanzu.lock"

	LocalProvidersConfigFileName = "config.yaml"
	LocalBOMsFolderName          = "bom"

	LocalProvidersChecksumFileName = "providers.sha256sum"
	OverrideFolder                 = "overrides"

	TKGKubeconfigDir    = ".kube-tkg"
	TKGKubeconfigFile   = "config"
	TKGKubeconfigTmpDir = "tmp"

	TKGConfigFileName               = "config.yaml"
	TKGDefaultClusterConfigFileName = "cluster-config.yaml"

	TKGClusterConfigFileDirForUI = "clusterconfigs"
	TKGRegistryCertFile          = "registry_certs"
)
