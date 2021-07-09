// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package azure defines client to connect to Azure cloud
package azure

import (
	"context"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// Client defines methods to access Azure inventory
type Client interface {
	VerifyAccount(ctx context.Context) error
	ListResourceGroups(ctx context.Context, location string) ([]*models.AzureResourceGroup, error)
	ListVirtualNetworks(ctx context.Context, resourceGroup string, location string) ([]*models.AzureVirtualNetwork, error)
	CreateResourceGroup(ctx context.Context, resourceGroupName string, location string) error
	CreateVirtualNetwork(ctx context.Context, resourceGroupName string, virtualNetworkName string, cidrBlock string, location string) error
	GetAzureRegions(ctx context.Context) ([]*models.AzureLocation, error)
	GetAzureInstanceTypesForRegion(ctx context.Context, region string) ([]*models.AzureInstanceType, error)
}
