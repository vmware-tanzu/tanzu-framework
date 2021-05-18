// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants defines various constants used in the code.
package constants

const (
	// TKGBomNamespace is the TKG add on BOM namespace.
	TKGBomNamespace = "tkr-system"

	// TKRLabel is the TKR label.
	TKRLabel = "tanzuKubernetesRelease"

	// TKGBomContent is the TKG BOM content.
	TKGBomContent = "bomContent"

	// TKRConfigmapName is the name of TKR config map
	TKRConfigmapName = "tkr-controller-config"

	// TKRRepoKey is the key for image repository in TKR config map data.
	TKRRepoKey = "imageRepository"

	// TKGPackageReconcilerKey is the log key for "name".
	TKGPackageReconcilerKey = "Package"

	// TKGAppReconcilerKey is the log key for "name".
	TKGAppReconcilerKey = "App"

	// TKGDataValueFormatString is required annotations for YTT data value file
	TKGDataValueFormatString = "#@data/values\n#@overlay/match-child-defaults missing_ok=True\n---\n"

	// TKGCorePackageRepositoryComponentName is the name of component that includes the package and repository images
	TKGCorePackageRepositoryComponentName = "tkg-core-packages"

	// TKGCorePackageRepositoryImageName is the name of core package repository image
	TKGCorePackageRepositoryImageName = "tanzuCorePackageRepositoryImage"
)
