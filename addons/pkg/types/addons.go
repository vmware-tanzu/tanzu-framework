// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types defines type constants.
package types

const (
	// AddonSecretType is the add on Secret type
	AddonSecretType = "tkg.tanzu.vmware.com/addon" // nolint:gosec

	// AddonNameLabel is the label on the Secret to indicate the name of addon to be installed
	AddonNameLabel = "tkg.tanzu.vmware.com/addon-name"

	// ClusterNameLabel is the label on the resource to indicate the cluster on which addon is to be installed
	ClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"

	// PackageNameLabel is the label on the cloned objects namely Secrets and Providers by "TanzuClusterBootstrap" Reconciler to indicate the package name
	PackageNameLabel = "tkg.tanzu.vmware.com/package-name"

	// YttMarkerAnnotation is the key for an annotation that indicates that data secret has Ytt markers
	YttMarkerAnnotation = "ext.packaging.carvel.dev/ytt-data-values-overlays"

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

	// ClusterNameAnnotation is the cluster's name annotation
	ClusterNameAnnotation = "tkg.tanzu.vmware.com/cluster-name"

	// ClusterNamespaceAnnotation is the cluster's namespace annotation
	ClusterNamespaceAnnotation = "tkg.tanzu.vmware.com/cluster-namespace"
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

// TKGSDataValues contains the package nodeSelector and update strategy information required by TKGS clusters
type TKGSDataValues struct {
	NodeSelector NodeSelector         `yaml:"nodeSelector"`
	Deployment   DeploymentUpdateInfo `yaml:"deployment,omitempty"`
	Daemonset    DaemonsetUpdateInfo  `yaml:"daemonset,omitempty"`
}

// NodeSelector contains the nodeSelector information
type NodeSelector struct {
	TanzuKubernetesRelease string `yaml:"run.tanzu.vmware.com/tkr"`
}

// DeploymentUpdateInfo contains the deployment update strategy information
type DeploymentUpdateInfo struct {
	UpdateStrategy string             `yaml:"updateStrategy,omitempty"`
	RollingUpdate  *RollingUpdateInfo `yaml:"rollingUpdate,omitempty"`
}

// RollingUpdateInfo contains the rolling update settings
type RollingUpdateInfo struct {
	MaxUnavailable int `yaml:"maxUnavailable"`
	MaxSurge       int `yaml:"maxSurge"`
}

// DaemonsetUpdateInfo contains the daumonset update strategy information
type DaemonsetUpdateInfo struct {
	UpdateStrategy string `yaml:"updateStrategy,omitempty"`
}
