// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types defines type constants.
package types

const (
	// AddonSecretType is the add on Secret type
	AddonSecretType = "tkg.tanzu.vmware.com/addon" // nolint:gosec

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

	// AddonExtYttPathsFromSecretNameAnnotation is the annotation that specifies a data secret has annotations
	AddonExtYttPathsFromSecretNameAnnotation = "ext.packaging.carvel.dev/ytt-data-values-overlays"

	// AddonNameAnnotation is the add on name annotation
	AddonNameAnnotation = AddonNameLabel

	// AddonNamespaceAnnotation is the add on's namespace annotation
	AddonNamespaceAnnotation = "tkg.tanzu.vmware.com/addon-namespace"

	// AddonPausedAnnotation is the add on's "paused" annotation
	AddonPausedAnnotation = "tkg.tanzu.vmware.com/addon-paused"
)

// AddonImageInfo contains addon image info
type AddonImageInfo struct {
	Info ImageInfo `yaml:"imageInfo"`
}

// ImageInfo contains addon image repository and URLs
type ImageInfo struct {
	ImageRepository string `yaml:"imageRepository"`
	ImagePullPolicy string `yaml:"imagePullPolicy"`
	// Each component can optionally have container images associated with it
	Images map[string]Image `yaml:"images,omitempty"`
}

// Image contains the image information
type Image struct {
	ImagePath string `yaml:"imagePath"`
	Tag       string `yaml:"tag"`
}
