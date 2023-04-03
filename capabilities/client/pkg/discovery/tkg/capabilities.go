// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
)

const (
	// clusterTypeManagement is the management cluster type.
	clusterTypeManagement = "management"
	// clusterTypeWorkload is the workload cluster type.
	clusterTypeWorkload = "workload"

	// metadataConfigMapNamespace is the namespace of the cluster metadata configmap.
	metadataConfigMapNamespace = "tkg-system-public"
	// metadataConfigMapName is the name of the cluster metadata configmap.
	metadataConfigMapName = "tkg-metadata"

	// namespaceNSX is the name of NSX namespace.
	namespaceNSX = "vmware-system-nsx"
)

// IsTKGm returns true if the cluster is a TKGm cluster.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) IsTKGm(ctx context.Context) (bool, error) {
	return dc.HasInfrastructureProvider(ctx, InfrastructureProviderVsphere)
}

// IsTKGS returns true if the cluster is a TKGS cluster. Checks for the existence of any TKC API version.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) IsTKGS(ctx context.Context) (bool, error) {
	query := discovery.Group("tkc", runv1alpha1.GroupVersion.Group).
		WithResource("tanzukubernetesclusters")
	return dc.clusterQueryClient.PreparedQuery(query)()
}

// IsManagementCluster returns true if the cluster is a TKG management cluster.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) IsManagementCluster(ctx context.Context) (bool, error) {
	s, err := clusterTypeFromMetadataConfigMap(ctx, dc.k8sClient)
	if err != nil {
		return false, err
	}
	return s == clusterTypeManagement, nil
}

// IsWorkloadCluster returns true if the cluster is a TKG workload cluster.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) IsWorkloadCluster(ctx context.Context) (bool, error) {
	s, err := clusterTypeFromMetadataConfigMap(ctx, dc.k8sClient)
	if err != nil {
		return false, err
	}
	return s == clusterTypeWorkload, nil
}

// clusterTypeFromMetadataConfigMap fetches cluster type from tkg-metadata configmap.
func clusterTypeFromMetadataConfigMap(ctx context.Context, c client.Client) (string, error) {
	cm := &corev1.ConfigMap{}
	key := client.ObjectKey{Namespace: metadataConfigMapNamespace, Name: metadataConfigMapName}
	if err := c.Get(ctx, key, cm); err != nil {
		return "", err
	}

	data, ok := cm.Data["metadata.yaml"]
	if !ok {
		return "", fmt.Errorf("failed to get cluster type: metadata.yaml key not found in configmap %s", key.String())
	}

	metadata := &ClusterMetadata{}
	if err := yaml.Unmarshal([]byte(data), metadata); err != nil {
		return "", fmt.Errorf("failed to get cluster type: %w", err)
	}
	switch strings.ToLower(metadata.Cluster.Type) {
	case clusterTypeManagement:
		return clusterTypeManagement, nil
	case clusterTypeWorkload:
		return clusterTypeWorkload, nil
	}
	return "", fmt.Errorf("unknown cluster type: %v", metadata.Cluster.Type)
}

// HasNSX indicates if a cluster has NSX capabilities.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) HasNSX(ctx context.Context) (bool, error) {
	nsx := &corev1.ObjectReference{
		Kind:       "Namespace",
		Name:       namespaceNSX,
		APIVersion: corev1.SchemeGroupVersion.Version,
	}
	query := discovery.Object("nsx", nsx)
	return dc.clusterQueryClient.PreparedQuery(query)()
}
