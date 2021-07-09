// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

// ManagementClusterVersion contains kubernetes versions that are supported by the management cluster with a certain TKG version.
type ManagementClusterVersion struct {
	TKGVersion                  string   `yaml:"version"`
	SupportedKubernetesVersions []string `yaml:"supportedKubernetesVersions"`
}

// CompatibilityMetadata contains tanzu release support matrix
type CompatibilityMetadata struct {
	ManagementClusterVersions []ManagementClusterVersion `yaml:"managementClusterVersions"`
}
