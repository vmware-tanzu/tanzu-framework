// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants defines constants used throughout the codebase.
package constants

var (
	// TKGMgmtClusterType is the "management" cluster type.
	TKGMgmtClusterType = "management"

	// TKGWorkloadClusterType is the "workload" cluster type.
	TKGWorkloadClusterType = "workload"

	// TKGSystemPublicNamespace is the "tkg-system-public" cluster type.
	TKGSystemPublicNamespace = "tkg-system-public"

	// TKGMetaConfigMapName is the "tkg-metadata" cluster type.
	TKGMetaConfigMapName = "tkg-metadata"

	// PinnipedInfoConfigMapName is the "pinniped-info" cluster type.
	PinnipedInfoConfigMapName = "pinniped-info"

	// KubePublicNamespace is the "kube-public" cluster type.
	KubePublicNamespace = "kube-public"

	// ClusterInfoConfigMapName is the "cluster-info" cluster type.
	ClusterInfoConfigMapName = "cluster-info"

	// DexClientID is the "pinniped-client-id" cluster type.
	DexClientID = "pinniped-client-id"

	// PinnipedDefaultAPIGroupSuffix is the API group suffix that Pinniped ships with by default.
	PinnipedDefaultAPIGroupSuffix = "pinniped.dev"
)
