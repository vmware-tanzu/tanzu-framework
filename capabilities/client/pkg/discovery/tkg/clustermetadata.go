// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

// ClusterMetadata is currently needed by one of the pre-defined queries in capabilities SDK
// to tell if a cluster is a management or a workload cluster
// Deprecated: This struct type will be removed in a future release.
type ClusterMetadata struct {
	Cluster Cluster `json:"cluster" yaml:"cluster"`
}

// Deprecated: This struct type will be removed in a future release.
type Cluster struct {
	Name               string         `json:"name" yaml:"name"`
	Type               string         `json:"type" yaml:"type"`
	Plan               string         `json:"plan" yaml:"plan"`
	KubernetesProvider string         `json:"kubernetesProvider" yaml:"kubernetesProvider"`
	TkgVersion         string         `json:"tkgVersion" yaml:"tkgVersion"`
	Infrastructure     Infrastructure `json:"infrastructure" yaml:"infrastructure"`
}

// Deprecated: This struct type will be removed in a future release.
type Infrastructure struct {
	Provider string `json:"provider" yaml:"provider"`
}
