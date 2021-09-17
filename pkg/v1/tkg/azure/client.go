// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute/computeapi"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network/networkapi"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources/resourcesapi"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions/subscriptionsapi"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

const (
	// ResourceTypeVirtualMachine defines virtualMachines resource type
	ResourceTypeVirtualMachine = "virtualMachines"
)

const (
	// ChinaCloud defines China cloud
	ChinaCloud = "AzureChinaCloud"
	// GermanCloud defines German cloud
	GermanCloud = "AzureGermanCloud"
	// PublicCloud defines Public cloud
	PublicCloud = "AzurePublicCloud"
	// USGovernmentCloud defines US Government cloud
	USGovernmentCloud = "AzureUSGovernmentCloud"
)

// Supported Azure VM family types
var supportedVMFamilyTypes = map[string]bool{
	"standardDSv3Family": true,
	"standardFSv2Family": true,
}

type client struct {
	SubscriptionID        string
	Authorizer            autorest.Authorizer
	ResourceGroupsClient  resourcesapi.GroupsClientAPI
	VirtualNetworksClient networkapi.VirtualNetworksClientAPI
	ResourceSkusClient    computeapi.ResourceSkusClientAPI
	SubscriptionsClient   subscriptionsapi.ClientAPI
}

// Credentials defines azure credentials
type Credentials struct {
	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string
	AzureCloud     string
}

// New creates an Azure client
func New(creds *Credentials) (Client, error) {
	if creds.AzureCloud == "" {
		creds.AzureCloud = PublicCloud
	}

	clientCredentialsConfig := auth.NewClientCredentialsConfig(creds.ClientID, creds.ClientSecret, creds.TenantID)
	if err := setActiveDirectoryEndpoint(&clientCredentialsConfig, creds.AzureCloud); err != nil {
		return nil, err
	}

	authorizer, err := clientCredentialsConfig.Authorizer()
	if err != nil {
		return nil, err
	}

	// Initialize resourceGroup client
	resourceGroupsClient := resources.NewGroupsClientWithBaseURI(clientCredentialsConfig.Resource, creds.SubscriptionID)
	resourceGroupsClient.Authorizer = authorizer

	// Initialize virtualNetwork client
	virtualNetworkClient := network.NewVirtualNetworksClientWithBaseURI(clientCredentialsConfig.Resource, creds.SubscriptionID)
	virtualNetworkClient.Authorizer = authorizer

	// Initialize resourceSkus client
	skuClient := compute.NewResourceSkusClientWithBaseURI(clientCredentialsConfig.Resource, creds.SubscriptionID)
	skuClient.Authorizer = authorizer

	// Initialize subscription client
	subscriptionClient := subscriptions.NewClientWithBaseURI(clientCredentialsConfig.Resource)
	subscriptionClient.Authorizer = authorizer

	return &client{
		SubscriptionID:        creds.SubscriptionID,
		Authorizer:            authorizer,
		ResourceGroupsClient:  resourceGroupsClient,
		VirtualNetworksClient: virtualNetworkClient,
		ResourceSkusClient:    skuClient,
		SubscriptionsClient:   subscriptionClient,
	}, nil
}

func setActiveDirectoryEndpoint(config *auth.ClientCredentialsConfig, azureCloud string) error {
	switch azureCloud {
	case USGovernmentCloud:
		config.Resource = azure.USGovernmentCloud.ResourceManagerEndpoint
		config.AADEndpoint = azure.USGovernmentCloud.ActiveDirectoryEndpoint
	case ChinaCloud:
		config.Resource = azure.ChinaCloud.ResourceManagerEndpoint
		config.AADEndpoint = azure.ChinaCloud.ActiveDirectoryEndpoint
	case GermanCloud:
		config.Resource = azure.GermanCloud.ResourceManagerEndpoint
		config.AADEndpoint = azure.GermanCloud.ActiveDirectoryEndpoint
	case PublicCloud:
		config.Resource = azure.PublicCloud.ResourceManagerEndpoint
		config.AADEndpoint = azure.PublicCloud.ActiveDirectoryEndpoint
	default:
		return errors.Errorf("%q is not a supported cloud in Azure. Supported clouds are AzurePublicCloud, AzureUSGovernmentCloud, AzureGermanCloud, AzureChinaCloud", azureCloud)
	}
	return nil
}

// verifies azure credentials by fetching the list of resource groups available
func (c *client) VerifyAccount(ctx context.Context) error {
	// fetching just one resource group for efficiency
	var resultCount int32 = 1
	_, err := c.ResourceGroupsClient.ListComplete(ctx, "", &resultCount)
	return err
}

// List all the resource groups
func (c *client) ListResourceGroups(ctx context.Context, location string) ([]*models.AzureResourceGroup, error) {
	var resourceGroups []*models.AzureResourceGroup

	rg, err := c.ResourceGroupsClient.ListComplete(ctx, "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch resource groups")
	}

	for rg.NotDone() {
		resourceGroup := &models.AzureResourceGroup{
			ID:       *rg.Value().ID,
			Location: rg.Value().Location,
			Name:     rg.Value().Name,
		}

		// Filter resource groups based on location if it not empty
		if location == "" || *resourceGroup.Location == location {
			resourceGroups = append(resourceGroups, resourceGroup)
		}
		if err := rg.NextWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "unable to fetch resource groups")
		}
	}
	return resourceGroups, nil
}

// Create a resource group
func (c *client) CreateResourceGroup(ctx context.Context, resourceGroupName, location string) error {
	rgParameters := resources.Group{
		Location: &location,
	}

	_, err := c.ResourceGroupsClient.CreateOrUpdate(ctx, resourceGroupName, rgParameters)
	return err
}

// List all virtual networks in a resource group
func (c *client) ListVirtualNetworks(ctx context.Context, resourceGroup, location string) ([]*models.AzureVirtualNetwork, error) {
	var virtualNetworks []*models.AzureVirtualNetwork
	vnet, err := c.VirtualNetworksClient.ListComplete(ctx, resourceGroup)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch virtual networks")
	}

	for ; vnet.NotDone(); err = vnet.NextWithContext(ctx) {
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch virtual networks")
		}

		// Filter virtual networks based on location if it not empty
		if location != "" && *vnet.Value().Location != location {
			continue
		}

		virtualNetwork := &models.AzureVirtualNetwork{
			ID:       *vnet.Value().ID,
			Location: vnet.Value().Location,
			Name:     vnet.Value().Name,
		}

		if vnet.Value().Subnets != nil {
			var azureSubnets []*models.AzureSubnet
			for _, s := range *vnet.Value().Subnets {
				azureSubnet := models.AzureSubnet{
					Name: *s.Name,
					Cidr: *s.AddressPrefix,
				}
				azureSubnets = append(azureSubnets, &azureSubnet)
			}
			virtualNetwork.Subnets = azureSubnets
		}

		virtualNetworks = append(virtualNetworks, virtualNetwork)
	}
	return virtualNetworks, nil
}

// Create a virtual network
func (c *client) CreateVirtualNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, cidrBlock, location string) error {
	vnetClient := network.NewVirtualNetworksClient(c.SubscriptionID)
	vnetClient.Authorizer = c.Authorizer

	addressPrefixes := []string{cidrBlock}
	vnetParams := network.VirtualNetwork{
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &addressPrefixes,
			},
		},
		Location: &location,
	}
	future, err := vnetClient.CreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, vnetParams)
	if err != nil {
		return err
	}

	err = future.WaitForCompletionRef(ctx, vnetClient.Client)
	return err
}

func (c *client) GetAzureRegions(ctx context.Context) ([]*models.AzureLocation, error) {
	regions := make(map[string]string)
	result := make([]*models.AzureLocation, 0)

	// list all available locations
	locationsResult, err := c.SubscriptionsClient.ListLocations(ctx, c.SubscriptionID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch Azure regions")
	}

	for _, location := range *locationsResult.Value {
		regions[strings.ToLower(*location.Name)] = *location.DisplayName
	}

	sku, err := c.ResourceSkusClient.ListComplete(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch Azure regions")
	}

	for sku.NotDone() {
		if *sku.Value().ResourceType == ResourceTypeVirtualMachine {
			if _, ok := supportedVMFamilyTypes[*sku.Value().Family]; ok {
				for _, locationInfo := range *sku.Value().LocationInfo {
					includeRegion := true
					if *sku.Value().Restrictions != nil {
						for _, restriction := range *sku.Value().Restrictions {
							if restriction.Type == compute.Location {
								includeRegion = false
								break
							}
						}
					}

					// don't include restricted regions
					if !includeRegion {
						continue
					}

					location := strings.ToLower(*locationInfo.Location)
					if displayName, ok := regions[location]; ok {
						result = append(result, &models.AzureLocation{
							Name:        location,
							DisplayName: displayName,
						})
						delete(regions, location)
					}
				}
			}
		}

		if err := sku.NextWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "unable to fetch Azure regions")
		}
	}

	return result, nil
}

// List all instance types for a region
func (c *client) GetAzureInstanceTypesForRegion(ctx context.Context, region string) ([]*models.AzureInstanceType, error) {
	filter := "location eq '" + region + "'"
	sku, err := c.ResourceSkusClient.ListComplete(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch Azure regions")
	}

	var instanceTypes []*models.AzureInstanceType

	for sku.NotDone() {
		if *sku.Value().ResourceType == ResourceTypeVirtualMachine {
			if _, ok := supportedVMFamilyTypes[*sku.Value().Family]; ok {
				for _, locationInfo := range *sku.Value().LocationInfo {
					instanceType := &models.AzureInstanceType{
						Family: *sku.Value().Family,
						Name:   *sku.Value().Name,
						Size:   *sku.Value().Size,
						Tier:   *sku.Value().Tier,
						Zones:  *locationInfo.Zones,
					}

					includeInstanceType := true
					if *sku.Value().Restrictions != nil {
						for _, restriction := range *sku.Value().Restrictions {
							if restriction.Type == compute.Location {
								includeInstanceType = false
								break
							}
						}
					}

					// don't include restricted instance types
					if includeInstanceType {
						instanceTypes = append(instanceTypes, instanceType)
					}
				}
			}
		}

		if err := sku.NextWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "unable to fetch Azure regions")
		}
	}
	return instanceTypes, nil
}
