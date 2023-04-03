// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
)

// HasTanzuRunGroup checks if run.tanzu.vmware.com API group exists and optionally checks versions.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) HasTanzuRunGroup(ctx context.Context, versions ...string) (bool, error) {
	query := discovery.Group("rungroup", runv1alpha1.GroupVersion.Group).WithVersions(versions...)
	return dc.clusterQueryClient.PreparedQuery(query)()
}

// HasTanzuKubernetesClusterV1alpha1 checks if the cluster has TanzuKubernetesCluster v1alpha1 resource.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) HasTanzuKubernetesClusterV1alpha1(ctx context.Context) (bool, error) {
	query := discovery.Group("tkc", runv1alpha1.GroupVersion.Group).
		WithVersions(runv1alpha1.GroupVersion.Version).
		WithResource("tanzukubernetesclusters")
	return dc.clusterQueryClient.PreparedQuery(query)()
}

// HasTanzuKubernetesReleaseV1alpha1 checks if the cluster has TanzuKubernetesRelease v1alpha1 resource.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) HasTanzuKubernetesReleaseV1alpha1(ctx context.Context) (bool, error) {
	query := discovery.Group("tkr", runv1alpha1.GroupVersion.Group).
		WithVersions(runv1alpha1.GroupVersion.Version).
		WithResource("tanzukubernetesreleases")
	return dc.clusterQueryClient.PreparedQuery(query)()
}
