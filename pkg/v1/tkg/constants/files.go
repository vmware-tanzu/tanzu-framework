/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
