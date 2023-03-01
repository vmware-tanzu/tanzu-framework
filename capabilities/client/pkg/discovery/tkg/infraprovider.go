// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"
	"fmt"
	"strings"

	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

// InfrastructureProvider represents the CAPI infrastructure provider of the cluster.
type InfrastructureProvider string

const (
	// InfrastructureProviderAWS is the AWS infrastructure provider.
	InfrastructureProviderAWS = InfrastructureProvider("aws")
	// InfrastructureProviderAzure is the Azure infrastructure provider.
	InfrastructureProviderAzure = InfrastructureProvider("azure")
	// InfrastructureProviderVsphere is the Vsphere infrastructure provider.
	InfrastructureProviderVsphere = InfrastructureProvider("vsphere")
)

// HasInfrastructureProvider checks the cluster's CAPI infrastructure provider.
// Deprecated: This function will be removed in a future release.
func (dc *DiscoveryClient) HasInfrastructureProvider(ctx context.Context, infraProvider InfrastructureProvider) (bool, error) {
	if infraProvider != InfrastructureProviderAWS && infraProvider != InfrastructureProviderAzure && infraProvider != InfrastructureProviderVsphere {
		return false, fmt.Errorf("unsupported infrastructure provider: %v", infraProvider)
	}

	var providerList clusterctl.ProviderList
	err := dc.k8sClient.List(ctx, &providerList)
	if err != nil {
		return false, err
	}

	for i := range providerList.Items {
		if providerList.Items[i].GetProviderType() != clusterctl.InfrastructureProviderType {
			continue
		}
		provider := strings.ToLower(providerList.Items[i].ProviderName)
		return string(infraProvider) == provider, nil
	}
	return false, fmt.Errorf("could not find infrastructure provider")
}
