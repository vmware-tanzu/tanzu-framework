// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// config key name constants
const (
	// OverrideFolderKey key for overrides folder to override the default overrides directory($HOME/.cluster-api/overrides)
	OverrideFolderKey          = "overridesFolder"
	ImagesConfigKey            = "images"
	ReleaseKey                 = "release"
	ProvidersConfigKey         = "providers"
	InfrastructureProviderType = "InfrastructureProvider"

	KeyTkg                  = "tkg"
	KeyRegions              = "regions"
	KeyRegionName           = "name"
	KeyCurrentRegionContext = "current-region-context"
	KeyRegionContext        = "context"

	KeyCertManagerTimeout = "cert-manager-timeout"
)
