// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// CloudProvider represents the cloud provider of the cluster.
type CloudProvider string

const (
	// CloudProviderAWS is the AWS cloud provider.
	CloudProviderAWS = CloudProvider("aws")
	// CloudProviderAzure is the Azure cloud provider.
	CloudProviderAzure = CloudProvider("azure")
	// CloudProviderVsphere is the Vsphere cloud provider.
	CloudProviderVsphere = CloudProvider("vsphere")
)

// HasCloudProvider checks if the cluster is configured with the given cloud provider.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) HasCloudProvider(ctx context.Context, cloudProvider CloudProvider) (bool, error) {
	if cloudProvider != CloudProviderAWS && cloudProvider != CloudProviderAzure && cloudProvider != CloudProviderVsphere {
		return false, fmt.Errorf("unsupported cloud provider: %v", cloudProvider)
	}

	nodeList := &corev1.NodeList{}
	if err := dc.k8sClient.List(ctx, nodeList); err != nil {
		return false, err
	}

	if len(nodeList.Items) < 1 {
		return false, fmt.Errorf("failed to identify cloud provider: node list is empty")
	}

	node := nodeList.Items[0]
	providerID := strings.Split(node.Spec.ProviderID, ":")
	if len(providerID) < 2 {
		return false, fmt.Errorf("unknown cloud provider")
	}
	provider := strings.ToLower(providerID[0])

	return string(cloudProvider) == provider, nil
}
