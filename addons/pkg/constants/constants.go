// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants defines various constants used in the code.
package constants

import "time"

const (
	/* Addon constants section */

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

	/* log key section */

	// NameLogKey is the log key for "name".
	NameLogKey = "name"

	// NamespaceLogKey is the log key for "namespace".
	NamespaceLogKey = "namespace"

	// AddonSecretNameLogKey is the log key for "addon-secret-name".
	AddonSecretNameLogKey = "addon-secret-name"

	// AddonSecretNamespaceLogKey is the log key for "addon-secret-ns"
	AddonSecretNamespaceLogKey = "addon-secret-ns" // nolint:gosec

	// AddonNameLogKey is the log key for "addon-name"
	AddonNameLogKey = "addon-name"

	// ImageNameLogKey is the log key for "image-name".
	ImageNameLogKey = "image-name"

	// ImageURLLogKey is the log key for "image-url".
	ImageURLLogKey = "image-url"

	// ComponentNameLogKey is the log key for "component-name".
	ComponentNameLogKey = "component-name"

	// KCPNameLogKey is the log key for "kcp-name"
	KCPNameLogKey = "kcp-name"

	// KCPNamespaceLogKey is the log key for "kcp-ns"
	KCPNamespaceLogKey = "kcp-ns"

	// TKRNameLogKey is the log key for "tkr-name"
	TKRNameLogKey = "tkr-name"

	// ClusterNameLogKey is the log key for "cluster-name"
	ClusterNameLogKey = "cluster-name"

	// ClusterNamespaceLogKey is the log key for "cluster-ns"
	ClusterNamespaceLogKey = "cluster-ns"

	// BOMNameLogKey is the log key for "bom-name"
	BOMNameLogKey = "bom-name"

	// BOMNamespaceLogKey is the log key for "bom-ns"
	BOMNamespaceLogKey = "bom-ns"

	// PackageRepositoryLogKey is the log key for "core-package-repository"
	PackageRepositoryLogKey = "core-package-repository"

	// AddonControllerName is name of addon-controller
	AddonControllerName = "addon-controller"

	// CRDWaitPollInterval is poll interval for checking server resources
	CRDWaitPollInterval = time.Second * 5

	// CRDWaitPollTimeout is poll timeout for checking server resources
	CRDWaitPollTimeout = time.Minute * 10
)
