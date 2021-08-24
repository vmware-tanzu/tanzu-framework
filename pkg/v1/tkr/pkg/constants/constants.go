// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

const (
	// BomConfigMapTKRLabel is the BOM ConfigMap TKR label
	BomConfigMapTKRLabel = "tanzuKubernetesRelease"

	// BomConfigMapImageTagAnnotation is the BOM ConfigMap image tag annotation
	BomConfigMapImageTagAnnotation = "bomImageTag"

	// BomConfigMapContentKey is the BOM ConfigMap content key
	BomConfigMapContentKey = "bomContent"

	// BOMKubernetesComponentKey is the BOM k8s component key
	BOMKubernetesComponentKey = "kubernetes"

	// ManagememtClusterRoleLabel is the management cluster role label
	ManagememtClusterRoleLabel = "cluster-role.tkg.tanzu.vmware.com/management"

	// TKGVersionKey is the TKG version key
	TKGVersionKey = "TKGVERSION"

	// TKRNamespace is the TKR namespace
	TKRNamespace = "tkr-system"

	// TKRControllerLeaderElectionCM is the ConfigMap used as the TKR Controller leader election lock
	TKRControllerLeaderElectionCM = "abf9f9ab.tanzu.vmware.com"

	// TKGNamespace is the TKG namespace
	TKGNamespace = "tkg-system"

	// TanzuKubernetesReleaseInactiveLabel is the TKR inactive label
	TanzuKubernetesReleaseInactiveLabel = "inactive"

	// BOMMetadataConfigMapName is the name of the ConfigMap holding BOM metadata
	BOMMetadataConfigMapName = "bom-metadata"

	// BOMMetadataCompatibilityKey in binaryData in bom-metadata ConfigMap holds compatibility metadata
	BOMMetadataCompatibilityKey = "compatibility"
)
