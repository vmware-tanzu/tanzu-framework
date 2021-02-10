// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

const (
	// TKGAddonsAppNamespace is the TKG add on app Namespace.
	TKGAddonsAppNamespace = "tkg-system"

	// TKGAddonsAppServiceAccount is the TKG add on ServiceAccount.
	TKGAddonsAppServiceAccount = "tkg-addons-app-sa"

	// TKGAddonsAppClusterRole is the TKG add on ClusterRole.
	TKGAddonsAppClusterRole = "tkg-addons-app-cluster-role"

	// TKGAddonsAppClusterRoleBinding is the TKG add on app ClusterRoleBinding.
	TKGAddonsAppClusterRoleBinding = "tkg-addons-app-cluster-role-binding"

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
)
