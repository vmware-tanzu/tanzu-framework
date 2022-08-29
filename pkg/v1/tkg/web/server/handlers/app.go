// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package handlers implements api handlers
package handlers

import (
	"path/filepath"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/edition"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/avi"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/docker"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/features"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/ldap"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/provider"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/vsphere"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/ws"
	aviClient "github.com/vmware-tanzu/tanzu-framework/tkg/avi"
	awsclient "github.com/vmware-tanzu/tanzu-framework/tkg/aws"
	azureclient "github.com/vmware-tanzu/tanzu-framework/tkg/azure"
	ldapClient "github.com/vmware-tanzu/tanzu-framework/tkg/ldap"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// App application structs consisting init options and clients
type App struct {
	InitOptions           client.InitRegionOptions
	AppConfig             types.AppConfig
	TKGTimeout            time.Duration
	vcClient              vc.Client
	awsClient             awsclient.Client
	azureClient           azureclient.Client
	aviClient             aviClient.Client
	ldapClient            ldapClient.Client
	TKGConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
	clusterConfigFile     string
}

// ConfigureHandlers configures API handlers func
func (app *App) ConfigureHandlers(api middleware.RoutableAPI) { // nolint:funlen
	a, ok := api.(*operations.KickstartUIAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.Logger = log.Infof

	a.JSONConsumer = runtime.JSONConsumer()

	a.JSONProducer = runtime.JSONProducer()

	a.ServerShutdown = func() {}

	a.ProviderGetProviderHandler = provider.GetProviderHandlerFunc(app.GetProvider)
	a.VsphereSetVSphereEndpointHandler = vsphere.SetVSphereEndpointHandlerFunc(app.SetVSphereEndpoint)
	a.VsphereGetVSphereDatacentersHandler = vsphere.GetVSphereDatacentersHandlerFunc(app.GetVSphereDatacenters)
	a.VsphereGetVSphereDatastoresHandler = vsphere.GetVSphereDatastoresHandlerFunc(app.GetVSphereDatastores)
	a.VsphereGetVSphereNetworksHandler = vsphere.GetVSphereNetworksHandlerFunc(app.GetVSphereNetworks)
	a.VsphereGetVSphereResourcePoolsHandler = vsphere.GetVSphereResourcePoolsHandlerFunc(app.GetVSphereResourcePools)
	a.VsphereCreateVSphereRegionalClusterHandler = vsphere.CreateVSphereRegionalClusterHandlerFunc(app.CreateVSphereRegionalCluster)
	a.VsphereGetVSphereOSImagesHandler = vsphere.GetVSphereOSImagesHandlerFunc(app.GetVsphereOSImages)
	a.VsphereGetVSphereFoldersHandler = vsphere.GetVSphereFoldersHandlerFunc(app.GetVSphereFolders)
	a.VsphereGetVSphereComputeResourcesHandler = vsphere.GetVSphereComputeResourcesHandlerFunc(app.GetVsphereComputeResources)
	a.VsphereApplyTKGConfigForVsphereHandler = vsphere.ApplyTKGConfigForVsphereHandlerFunc(app.ApplyTKGConfigForVsphere)
	a.VsphereGetVsphereThumbprintHandler = vsphere.GetVsphereThumbprintHandlerFunc(app.GetVsphereThumbprint)
	a.VsphereExportTKGConfigForVsphereHandler = vsphere.ExportTKGConfigForVsphereHandlerFunc(app.ExportVSphereConfig)
	a.VsphereImportTKGConfigForVsphereHandler = vsphere.ImportTKGConfigForVsphereHandlerFunc(app.ImportVSphereConfig)

	a.AwsSetAWSEndpointHandler = aws.SetAWSEndpointHandlerFunc(app.SetAWSEndPoint)
	a.AwsGetVPCsHandler = aws.GetVPCsHandlerFunc(app.GetVPCs)
	a.AwsGetAWSAvailabilityZonesHandler = aws.GetAWSAvailabilityZonesHandlerFunc(app.GetAWSAvailabilityZones)
	a.AwsGetAWSRegionsHandler = aws.GetAWSRegionsHandlerFunc(app.GetAWSRegions)
	a.AwsCreateAWSRegionalClusterHandler = aws.CreateAWSRegionalClusterHandlerFunc(app.CreateAWSRegionalCluster)
	a.AwsGetAWSSubnetsHandler = aws.GetAWSSubnetsHandlerFunc(app.GetAWSSubnets)
	a.AwsApplyTKGConfigForAWSHandler = aws.ApplyTKGConfigForAWSHandlerFunc(app.ApplyTKGConfigForAWS)
	a.AwsGetAWSNodeTypesHandler = aws.GetAWSNodeTypesHandlerFunc(app.GetAWSNodeTypes)
	a.AwsGetAWSCredentialProfilesHandler = aws.GetAWSCredentialProfilesHandlerFunc(app.GetAWSCredentialProfiles)
	a.AwsGetAWSOSImagesHandler = aws.GetAWSOSImagesHandlerFunc(app.GetAWSOSImages)
	a.AwsExportTKGConfigForAWSHandler = aws.ExportTKGConfigForAWSHandlerFunc(app.ExportAWSConfig)
	a.AwsImportTKGConfigForAWSHandler = aws.ImportTKGConfigForAWSHandlerFunc(app.ImportAwsConfig)

	a.AzureGetAzureEndpointHandler = azure.GetAzureEndpointHandlerFunc(app.GetAzureEndpoint)
	a.AzureSetAzureEndpointHandler = azure.SetAzureEndpointHandlerFunc(app.SetAzureEndPoint)
	a.AzureGetAzureResourceGroupsHandler = azure.GetAzureResourceGroupsHandlerFunc(app.GetAzureResourceGroups)
	a.AzureCreateAzureResourceGroupHandler = azure.CreateAzureResourceGroupHandlerFunc(app.CreateAzureResourceGroup)
	a.AzureGetAzureVnetsHandler = azure.GetAzureVnetsHandlerFunc(app.GetAzureVirtualNetworks)
	a.AzureCreateAzureVirtualNetworkHandler = azure.CreateAzureVirtualNetworkHandlerFunc(app.CreateAzureVirtualNetwork)
	a.AzureGetAzureRegionsHandler = azure.GetAzureRegionsHandlerFunc(app.GetAzureRegions)
	a.AzureGetAzureInstanceTypesHandler = azure.GetAzureInstanceTypesHandlerFunc(app.GetAzureInstanceTypes)
	a.AzureApplyTKGConfigForAzureHandler = azure.ApplyTKGConfigForAzureHandlerFunc(app.ApplyTKGConfigForAzure)
	a.AzureCreateAzureRegionalClusterHandler = azure.CreateAzureRegionalClusterHandlerFunc(app.CreateAzureRegionalCluster)
	a.AzureGetAzureOSImagesHandler = azure.GetAzureOSImagesHandlerFunc(app.GetAzureOSImages)
	a.AzureExportTKGConfigForAzureHandler = azure.ExportTKGConfigForAzureHandlerFunc(app.ExportAzureConfig)
	a.AzureImportTKGConfigForAzureHandler = azure.ImportTKGConfigForAzureHandlerFunc(app.ImportAzureConfig)

	a.DockerCheckIfDockerDaemonAvailableHandler = docker.CheckIfDockerDaemonAvailableHandlerFunc(app.IsDockerDaemonAvailable)
	a.DockerApplyTKGConfigForDockerHandler = docker.ApplyTKGConfigForDockerHandlerFunc(app.ApplyTKGConfigForDocker)
	a.DockerCreateDockerRegionalClusterHandler = docker.CreateDockerRegionalClusterHandlerFunc(app.CreateDockerRegionalCluster)
	a.DockerExportTKGConfigForDockerHandler = docker.ExportTKGConfigForDockerHandlerFunc(app.ExportDockerConfig)
	a.DockerImportTKGConfigForDockerHandler = docker.ImportTKGConfigForDockerHandlerFunc(app.ImportDockerConfig)

	a.FeaturesGetFeatureFlagsHandler = features.GetFeatureFlagsHandlerFunc(app.GetFeatureFlags)
	a.EditionGetTanzuEditionHandler = edition.GetTanzuEditionHandlerFunc(app.GetTanzuEdition)

	a.AviVerifyAccountHandler = avi.VerifyAccountHandlerFunc(app.VerifyAccount)
	a.AviGetAviCloudsHandler = avi.GetAviCloudsHandlerFunc(app.GetAviClouds)
	a.AviGetAviServiceEngineGroupsHandler = avi.GetAviServiceEngineGroupsHandlerFunc(app.GetAviServiceEngineGroups)
	a.AviGetAviVipNetworksHandler = avi.GetAviVipNetworksHandlerFunc(app.GetAviVipNetworks)

	a.LdapVerifyLdapConnectHandler = ldap.VerifyLdapConnectHandlerFunc(app.VerifyLdapConnect)
	a.LdapVerifyLdapBindHandler = ldap.VerifyLdapBindHandlerFunc(app.VerifyLdapBind)
	a.LdapVerifyLdapUserSearchHandler = ldap.VerifyLdapUserSearchHandlerFunc(app.VerifyUserSearch)
	a.LdapVerifyLdapGroupSearchHandler = ldap.VerifyLdapGroupSearchHandlerFunc(app.VerifyGroupSearch)
	a.LdapVerifyLdapCloseConnectionHandler = ldap.VerifyLdapCloseConnectionHandlerFunc(app.VerifyLdapCloseConnection)
}

// StartSendingLogsToUI creates logchannel passes it to tkg logger
// retrieves the logs through logChannel and passes it to webSocket
func (app *App) StartSendingLogsToUI() {
	logChannel := make(chan []byte)
	log.SetChannel(logChannel)
	for logMsg := range logChannel {
		ws.SendLog(logMsg)
	}
}

func (app *App) getFilePathForSavingConfig() string {
	if app.clusterConfigFile == "" {
		randomFileName := utils.GenerateRandomID(10, true) + ".yaml"
		app.clusterConfigFile = filepath.Join(app.AppConfig.TKGConfigDir, constants.TKGClusterConfigFileDirForUI, randomFileName)
	}
	return app.clusterConfigFile
}
