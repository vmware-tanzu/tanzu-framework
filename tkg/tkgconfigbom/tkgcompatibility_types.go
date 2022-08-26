// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigbom

// TKGSupportedBOMVersions contains ImagePath and Tag name of supported BOM.
type TKGSupportedBOMVersions struct {
	ImagePath string `yaml:"imagePath"`
	ImageTag  string `yaml:"tag"`
}

// ManagementClusterPluginVersion contains management cluster plugin version and the supported TKG BOM versions.
type ManagementClusterPluginVersion struct {
	Version                 string                    `yaml:"version"`
	SupportedTKGBOMVersions []TKGSupportedBOMVersions `yaml:"supportedTKGBomVersions"`
}

// TKGCompatibilityMetadata contains Tanzu CLI supported TKG BOM version matrix
type TKGCompatibilityMetadata struct {
	Version                         string                           `yaml:"version"`
	ManagementClusterPluginVersions []ManagementClusterPluginVersion `yaml:"managementClusterPluginVersions"`
}
