// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// AddonSecretType is the add on Secret type
	AddonSecretType = "tkg.tanzu.vmware.com/addon"

	// AddonNameLabel is the label on the Secret to indicate the name of addon to be installed
	AddonNameLabel = "tkg.tanzu.vmware.com/addon-name"

	// ClusterNameLabel is the label on the Secret to indicate the cluster on which addon is to be installed
	ClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"

	// AddonFinalizer is the finalizer for the add on.
	AddonFinalizer = "tkg.tanzu.vmware.com/addon"

	// AddonTypeAnnotation is the add on type annotation
	AddonTypeAnnotation = "tkg.tanzu.vmware.com/addon-type"

	// AddonRemoteAppAnnotation is the add on remote app annotation
	AddonRemoteAppAnnotation = "tkg.tanzu.vmware.com/remote-app"

	// AddonNameAnnotation is the add on name annotation
	AddonNameAnnotation = AddonNameLabel

	// AddonNamespaceAnnotation is the add on's namespace annotation
	AddonNamespaceAnnotation = "tkg.tanzu.vmware.com/addon-namespace"

	// AddonPausedAnnotation is the add on's "paused" annotation
	AddonPausedAnnotation = "tkg.tanzu.vmware.com/addon-paused"
)
