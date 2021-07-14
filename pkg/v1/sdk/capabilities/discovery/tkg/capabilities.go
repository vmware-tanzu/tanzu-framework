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

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
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
func (dc *DiscoveryClient) IsTKGm(ctx context.Context) (bool, error) {
	return dc.HasInfrastructureProvider(ctx, InfrastructureProviderVsphere)
}

// IsTKGS returns true if the cluster is a TKGS cluster.
func (dc *DiscoveryClient) IsTKGS(ctx context.Context) (bool, error) {
	return dc.HasTanzuKubernetesClusterV1alpha1(ctx)
}

// IsManagementCluster returns true if the cluster is a TKG management cluster.
func (dc *DiscoveryClient) IsManagementCluster(ctx context.Context) (bool, error) {
	s, err := clusterTypeFromMetadataConfigMap(ctx, dc.k8sClient)
	if err != nil {
		return false, err
	}
	return s == clusterTypeManagement, nil
}

// IsWorkloadCluster returns true if the cluster is a TKG workload cluster.
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

	// ClusterMetadata is not defined in a common package, so re-use this.
	metadata := &shared.ClusterMetadata{}
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
func (dc *DiscoveryClient) HasNSX(ctx context.Context) (bool, error) {
	nsx := &corev1.ObjectReference{
		Kind:       "Namespace",
		Name:       namespaceNSX,
		APIVersion: corev1.SchemeGroupVersion.Version,
	}
	query := discovery.Object(nsx)
	return dc.clusterQueryClient.PreparedQuery(query)()
}
