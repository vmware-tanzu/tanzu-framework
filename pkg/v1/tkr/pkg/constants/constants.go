// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

const (
	BomConfigMapTKRLabel           = "tanzuKubernetesRelease"
	BomConfigMapImageTagAnnotation = "bomImageTag"
	BomConfigMapContentKey         = "bomContent"

	BOMKubernetesComponentKey = "kubernetes"

	ManagememtClusterRoleLabel = "cluster-role.tkg.tanzu.vmware.com/management"
	TKGVersionKey              = "TKGVERSION"

	TKRNamespace = "tkr-system"
	TKGNamespace = "tkg-system"
)
