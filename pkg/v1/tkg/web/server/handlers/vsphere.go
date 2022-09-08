// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/vsphere"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

const trueString = "true"

// GetVsphereThumbprint gets the vSphere thumbprint if insecure flag not set
func (app *App) GetVsphereThumbprint(params vsphere.GetVsphereThumbprintParams) middleware.Responder {
	insecure := false
	thumbprint := ""
	var err error

	thumbprint, err = vc.GetVCThumbprint(params.Host)
	if err != nil {
		return vsphere.NewGetVsphereThumbprintInternalServerError().WithPayload(Err(err))
	}

	res := models.VSphereThumbprint{Thumbprint: thumbprint, Insecure: &insecure}

	return vsphere.NewGetVsphereThumbprintOK().WithPayload(&res)
}

// SetVSphereEndpoint validates vsphere credentials and sets the vsphere client into web app
func (app *App) SetVSphereEndpoint(params vsphere.SetVSphereEndpointParams) middleware.Responder {
	host := strings.TrimSpace(params.Credentials.Host)

	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}
	vcURL, err := url.Parse(host)
	if err != nil {
		return vsphere.NewSetVSphereEndpointInternalServerError().WithPayload(Err(err))
	}

	vcURL.Path = "/sdk"

	vsphereInsecure := false
	vsphereInsecureString, err := app.TKGConfigReaderWriter.Get(constants.ConfigVariableVsphereInsecure)
	if err == nil {
		vsphereInsecure = (vsphereInsecureString == trueString)
	}

	if params.Credentials.Insecure != nil && *params.Credentials.Insecure {
		vsphereInsecure = true
	}

	vcClient, err := vc.NewClient(vcURL, params.Credentials.Thumbprint, vsphereInsecure)
	if err != nil {
		return vsphere.NewSetVSphereEndpointInternalServerError().WithPayload(Err(err))
	}

	_, err = vcClient.Login(params.HTTPRequest.Context(), params.Credentials.Username, params.Credentials.Password)
	if err != nil {
		return vsphere.NewSetVSphereEndpointInternalServerError().WithPayload(Err(err))
	}

	app.vcClient = vcClient

	version, build, err := vcClient.GetVSphereVersion()
	if err != nil {
		return vsphere.NewSetVSphereEndpointInternalServerError().WithPayload(Err(err))
	}

	res := models.VsphereInfo{
		Version:    fmt.Sprintf("%s:%s", version, build),
		HasPacific: "no",
	}

	if hasPP, err := vcClient.DetectPacific(params.HTTPRequest.Context()); err == nil && hasPP {
		res.HasPacific = "yes"
	}

	return vsphere.NewSetVSphereEndpointCreated().WithPayload(&res)
}

// GetVSphereDatacenters returns all the datacenters in vsphere
func (app *App) GetVSphereDatacenters(params vsphere.GetVSphereDatacentersParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereDatacentersInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	datacenters, err := app.vcClient.GetDatacenters(params.HTTPRequest.Context())
	if err != nil {
		return vsphere.NewGetVSphereDatacentersInternalServerError().WithPayload(Err(err))
	}

	return vsphere.NewGetVSphereDatacentersOK().WithPayload(datacenters)
}

// GetVSphereDatastores returns all the datastores in the datacenter
func (app *App) GetVSphereDatastores(params vsphere.GetVSphereDatastoresParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereDatastoresInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	datastores, err := app.vcClient.GetDatastores(params.HTTPRequest.Context(), params.Dc)
	if err != nil {
		return vsphere.NewGetVSphereDatastoresInternalServerError().WithPayload(Err(err))
	}

	return vsphere.NewGetVSphereDatastoresOK().WithPayload(datastores)
}

// GetVSphereNetworks gets all the  networks in the datacenter
func (app *App) GetVSphereNetworks(params vsphere.GetVSphereNetworksParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereNetworksInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	networks, err := app.vcClient.GetNetworks(params.HTTPRequest.Context(), params.Dc)
	if err != nil {
		return vsphere.NewGetVSphereNetworksInternalServerError().WithPayload(Err(err))
	}

	return vsphere.NewGetVSphereNetworksOK().WithPayload(networks)
}

// GetVSphereResourcePools gets all the resource pools in the datacenter
func (app *App) GetVSphereResourcePools(params vsphere.GetVSphereResourcePoolsParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereResourcePoolsInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	rps, err := app.vcClient.GetResourcePools(params.HTTPRequest.Context(), params.Dc)
	if err != nil {
		return vsphere.NewGetVSphereResourcePoolsInternalServerError().WithPayload(Err(err))
	}

	return vsphere.NewGetVSphereResourcePoolsOK().WithPayload(rps)
}

// GetVsphereOSImages gets vm templates for deploying kubernetes node
func (app *App) GetVsphereOSImages(params vsphere.GetVSphereOSImagesParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereOSImagesInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	vms, err := app.vcClient.GetVirtualMachineImages(params.HTTPRequest.Context(), params.Dc)
	if err != nil {
		return vsphere.NewGetVSphereOSImagesInternalServerError().WithPayload(Err(err))
	}

	defaultTKRBom, err := tkgconfigbom.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).GetDefaultTkrBOMConfiguration()
	if err != nil {
		return vsphere.NewGetVSphereOSImagesInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get the default TanzuKubernetesRelease")))
	}
	matchedTemplates, nonTemplateVms := vc.FindMatchingVirtualMachineTemplate(vms, defaultTKRBom.GetOVAVersions())

	if len(matchedTemplates) == 0 && len(nonTemplateVms) != 0 {
		log.Infof("unable to find any VM Template associated with the TanzuKubernetesRelease %s, but found these VM(s) [%s] that can be used once converted to a VM Template", defaultTKRBom.Release.Version, strings.Join(nonTemplateVms, ","))
	}

	results := []*models.VSphereVirtualMachine{}

	for _, template := range matchedTemplates {
		results = append(results, &models.VSphereVirtualMachine{
			IsTemplate: &template.IsTemplate,
			Name:       template.Name,
			Moid:       template.Moid,
			OsInfo: &models.OSInfo{
				Name:    template.DistroName,
				Version: template.DistroVersion,
				Arch:    template.DistroArch,
			},
		})
	}

	return vsphere.NewGetVSphereOSImagesOK().WithPayload(results)
}

// GetVSphereFolders gets vsphere folders
func (app *App) GetVSphereFolders(params vsphere.GetVSphereFoldersParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereFoldersInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	folders, err := app.vcClient.GetFolders(params.HTTPRequest.Context(), params.Dc)
	if err != nil {
		return vsphere.NewGetVSphereFoldersInternalServerError().WithPayload(Err(err))
	}

	return vsphere.NewGetVSphereFoldersOK().WithPayload(folders)
}

// GetVsphereComputeResources gets vsphere compute resources
func (app *App) GetVsphereComputeResources(params vsphere.GetVSphereComputeResourcesParams) middleware.Responder {
	if app.vcClient == nil {
		return vsphere.NewGetVSphereComputeResourcesInternalServerError().WithPayload(Err(errors.New("vSphere client is not initialized properly")))
	}

	results, err := app.vcClient.GetComputeResources(params.HTTPRequest.Context(), params.Dc)
	if err != nil {
		return vsphere.NewGetVSphereComputeResourcesInternalServerError().WithPayload(Err(err))
	}

	return vsphere.NewGetVSphereComputeResourcesOK().WithPayload(results)
}
