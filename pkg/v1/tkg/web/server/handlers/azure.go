// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/azure"
	azureclient "github.com/vmware-tanzu/tanzu-framework/tkg/azure"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

// GetAzureEndpoint gets azure account info set in environment variables
func (app *App) GetAzureEndpoint(params azure.GetAzureEndpointParams) middleware.Responder {
	res := models.AzureAccountParams{
		SubscriptionID: os.Getenv(constants.ConfigVariableAzureSubscriptionID),
		TenantID:       os.Getenv(constants.ConfigVariableAzureTenantID),
		ClientID:       os.Getenv(constants.ConfigVariableAzureClientID),
		ClientSecret:   os.Getenv(constants.ConfigVariableAzureClientSecret),
	}

	return azure.NewGetAzureEndpointOK().WithPayload(&res)
}

// SetAzureEndPoint verify and sets Azure account
func (app *App) SetAzureEndPoint(params azure.SetAzureEndpointParams) middleware.Responder {
	creds := azureclient.Credentials{
		SubscriptionID: params.AccountParams.SubscriptionID,
		ClientID:       params.AccountParams.ClientID,
		ClientSecret:   params.AccountParams.ClientSecret,
		TenantID:       params.AccountParams.TenantID,
		AzureCloud:     params.AccountParams.AzureCloud,
	}

	client, err := azureclient.New(&creds)
	if err != nil {
		return azure.NewSetAzureEndpointInternalServerError().WithPayload(Err(err))
	}

	err = client.VerifyAccount(params.HTTPRequest.Context())
	if err != nil {
		return azure.NewSetAzureEndpointInternalServerError().WithPayload(Err(err))
	}

	app.azureClient = client
	return azure.NewSetAzureEndpointCreated()
}

// GetAzureResourceGroups gets the list of all Azure resource groups
func (app *App) GetAzureResourceGroups(params azure.GetAzureResourceGroupsParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewGetAzureResourceGroupsInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	resourceGroups, err := app.azureClient.ListResourceGroups(params.HTTPRequest.Context(), params.Location)
	if err != nil {
		return azure.NewGetAzureResourceGroupsInternalServerError().WithPayload(Err(err))
	}

	return azure.NewGetAzureResourceGroupsOK().WithPayload(resourceGroups)
}

// CreateAzureResourceGroup creates a new Azure resource group
func (app *App) CreateAzureResourceGroup(params azure.CreateAzureResourceGroupParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewCreateAzureResourceGroupInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	err := app.azureClient.CreateResourceGroup(params.HTTPRequest.Context(), *params.Params.Name, *params.Params.Location)
	if err != nil {
		return azure.NewCreateAzureResourceGroupInternalServerError().WithPayload(Err(err))
	}

	return azure.NewCreateAzureResourceGroupCreated()
}

// GetAzureVirtualNetworks gets the list of all Azure virtual networks for a resource group
func (app *App) GetAzureVirtualNetworks(params azure.GetAzureVnetsParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewGetAzureVnetsInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	vnets, err := app.azureClient.ListVirtualNetworks(params.HTTPRequest.Context(), params.ResourceGroupName, params.Location)
	if err != nil {
		return azure.NewGetAzureVnetsInternalServerError().WithPayload(Err(err))
	}

	return azure.NewGetAzureVnetsOK().WithPayload(vnets)
}

// CreateAzureVirtualNetwork creates a new Azure Virtual Network
func (app *App) CreateAzureVirtualNetwork(params azure.CreateAzureVirtualNetworkParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewCreateAzureVirtualNetworkInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	err := app.azureClient.CreateVirtualNetwork(params.HTTPRequest.Context(), params.ResourceGroupName, *params.Params.Name, *params.Params.CidrBlock, *params.Params.Location)
	if err != nil {
		return azure.NewCreateAzureVirtualNetworkInternalServerError().WithPayload(Err(err))
	}

	return azure.NewCreateAzureVirtualNetworkCreated()
}

// GetAzureRegions gets a list of all available regions
func (app *App) GetAzureRegions(params azure.GetAzureRegionsParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewGetAzureRegionsInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	regions, err := app.azureClient.GetAzureRegions(params.HTTPRequest.Context())
	if err != nil {
		return azure.NewGetAzureRegionsInternalServerError().WithPayload(Err(err))
	}

	return azure.NewGetAzureRegionsOK().WithPayload(regions)
}

// GetAzureInstanceTypes lists the available instance types for a given region
func (app *App) GetAzureInstanceTypes(params azure.GetAzureInstanceTypesParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewGetAzureInstanceTypesInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	instanceTypes, err := app.azureClient.GetAzureInstanceTypesForRegion(params.HTTPRequest.Context(), params.Location)
	if err != nil {
		return azure.NewGetAzureInstanceTypesInternalServerError().WithPayload(Err(err))
	}

	return azure.NewGetAzureInstanceTypesOK().WithPayload(instanceTypes)
}

// GetAzureOSImages gets os information for Azure
func (app *App) GetAzureOSImages(params azure.GetAzureOSImagesParams) middleware.Responder {
	bomConfig, err := tkgconfigbom.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).GetDefaultTkrBOMConfiguration()
	if err != nil {
		return azure.NewGetAzureOSImagesInternalServerError().WithPayload(Err(err))
	}

	results := []*models.AzureVirtualMachine{}
	for i := range bomConfig.Azure {
		displayName := fmt.Sprintf("%s-%s-%s (%s)", bomConfig.Azure[i].OSInfo.Name, bomConfig.Azure[i].OSInfo.Version, bomConfig.Azure[i].OSInfo.Arch, bomConfig.Azure[i].Version)
		results = append(results, &models.AzureVirtualMachine{
			Name: displayName,
			OsInfo: &models.OSInfo{
				Name:    bomConfig.Azure[i].OSInfo.Name,
				Version: bomConfig.Azure[i].OSInfo.Version,
				Arch:    bomConfig.Azure[i].OSInfo.Arch,
			},
		})
	}

	return azure.NewGetAzureOSImagesOK().WithPayload(results)
}
