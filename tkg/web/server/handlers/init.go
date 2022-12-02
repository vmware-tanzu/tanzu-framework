// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"os"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clientcreator"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/azure"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/docker"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/vsphere"
)

// infrastructure name constants
const (
	InfrastructureProviderVSphere = "vsphere"
	InfrastructureProviderAWS     = "aws"
	InfrastructureProviderAzure   = "azure"
	InfrastructureProviderDocker  = "docker"
)

const sleepTimeForLogsPropogation = 2 * time.Second

// CreateVSphereRegionalCluster creates vSphere management cluster
func (app *App) CreateVSphereRegionalCluster(params vsphere.CreateVSphereRegionalClusterParams) middleware.Responder {
	vsphereConfig, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewVSphereConfig(params.Params)
	if err != nil {
		return vsphere.NewCreateVSphereRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, vsphereConfig)
	if err != nil {
		return vsphere.NewCreateVSphereRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	allClients, err := clientcreator.CreateAllClients(app.AppConfig, app.TKGConfigReaderWriter)
	if err != nil {
		return vsphere.NewCreateVSphereRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	c, err := client.New(client.Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             app.AppConfig.TKGConfigDir,
		Timeout:                  app.TKGTimeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		ClusterClientFactory:     clusterclient.NewClusterClientFactory(),
		FeatureFlagClient:        getFeatureFlagClient(),
	})
	if err != nil {
		return vsphere.NewCreateVSphereRegionalClusterInternalServerError().WithPayload(Err(err))
	}
	app.InitOptions.InfrastructureProvider = InfrastructureProviderVSphere
	app.InitOptions.ClusterName = params.Params.ClusterName
	app.InitOptions.Plan = params.Params.ControlPlaneFlavor
	app.InitOptions.Annotations = params.Params.Annotations
	app.InitOptions.Labels = params.Params.Labels
	app.InitOptions.CeipOptIn = *params.Params.CeipOptIn
	app.InitOptions.CniType = params.Params.Networking.CniType
	app.InitOptions.VsphereControlPlaneEndpoint = params.Params.ControlPlaneEndpoint
	app.InitOptions.ClusterConfigFile = app.getFilePathForSavingConfig()

	if err := c.ConfigureAndValidateManagementClusterConfiguration(&app.InitOptions, false); err != nil {
		return vsphere.NewCreateVSphereRegionalClusterInternalServerError().WithPayload(Err(errors.New(err.Message)))
	}
	go app.StartSendingLogsToUI()
	go func() {
		err := c.InitRegion(&app.InitOptions)
		if err != nil {
			log.Error(err, "unable to set up management cluster, ")
		} else {
			log.Infof("\nManagement cluster created!\n\n")
			log.Info("\nYou can now create your first workload cluster by running the following:\n\n")
			log.Info("  tanzu cluster create [name] -f [file]\n\n")
			err = pluginmanager.SyncPlugins()
			if err != nil {
				log.Warningf("unable to sync plugins after management cluster create. Please run `tanzu plugin sync` command manually to install/update plugins")
			}
			// wait for the logs to be dispatched to UI before exit
			time.Sleep(sleepTimeForLogsPropogation)
			// exit the BE server on success
			os.Exit(0)
		}
	}()

	return vsphere.NewCreateVSphereRegionalClusterOK().WithPayload("started creating regional cluster")
}

// CreateAWSRegionalCluster creates aws management cluster
func (app *App) CreateAWSRegionalCluster(params aws.CreateAWSRegionalClusterParams) middleware.Responder {
	if app.awsClient == nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(errors.New("aws client is not initialized properly")))
	}
	encodedCreds, err := app.awsClient.EncodeCredentials()
	if err != nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	awsConfig, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewAWSConfig(params.Params, encodedCreds)
	if err != nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, awsConfig)
	if err != nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	allClients, err := clientcreator.CreateAllClients(app.AppConfig, app.TKGConfigReaderWriter)
	if err != nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	c, err := client.New(client.Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             app.AppConfig.TKGConfigDir,
		Timeout:                  app.TKGTimeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		ClusterClientFactory:     clusterclient.NewClusterClientFactory(),
		FeatureFlagClient:        getFeatureFlagClient(),
	})
	if err != nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	app.InitOptions.InfrastructureProvider = InfrastructureProviderAWS
	app.InitOptions.ClusterName = params.Params.ClusterName
	app.InitOptions.Plan = params.Params.ControlPlaneFlavor
	app.InitOptions.CeipOptIn = *params.Params.CeipOptIn
	app.InitOptions.CniType = params.Params.Networking.CniType
	app.InitOptions.Annotations = params.Params.Annotations
	app.InitOptions.Labels = params.Params.Labels
	app.InitOptions.ClusterConfigFile = app.getFilePathForSavingConfig()
	if err := c.ConfigureAndValidateManagementClusterConfiguration(&app.InitOptions, false); err != nil {
		return aws.NewCreateAWSRegionalClusterInternalServerError().WithPayload(Err(errors.New(err.Message)))
	}
	go app.StartSendingLogsToUI()

	go func() {
		if params.Params.CreateCloudFormationStack {
			err = c.CreateAWSCloudFormationStack()
			if err != nil {
				log.Error(err, "unable to create AWS CloudFormationStack")
				return
			}
		}
		err := c.InitRegion(&app.InitOptions)
		if err != nil {
			log.Error(err, "unable to set up management cluster, ")
		} else {
			log.Infof("\nManagement cluster created!\n\n")
			log.Info("\nYou can now create your first workload cluster by running the following:\n\n")
			log.Info("  tanzu cluster create [name] -f [file]\n\n")
			err = pluginmanager.SyncPlugins()
			if err != nil {
				log.Warningf("unable to sync plugins after management cluster create. Please run `tanzu plugin sync` command manually to install/update plugins")
			}
			// wait for the logs to be dispatched to UI before exit
			time.Sleep(sleepTimeForLogsPropogation)
			// exit the BE server on success
			os.Exit(0)
		}
	}()

	return aws.NewCreateAWSRegionalClusterOK().WithPayload("started creating regional cluster")
}

// CreateAzureRegionalCluster creates azure management cluster
func (app *App) CreateAzureRegionalCluster(params azure.CreateAzureRegionalClusterParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	azureConfig, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewAzureConfig(params.Params)
	if err != nil {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, azureConfig)
	if err != nil {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	allClients, err := clientcreator.CreateAllClients(app.AppConfig, app.TKGConfigReaderWriter)
	if err != nil {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	c, err := client.New(client.Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             app.AppConfig.TKGConfigDir,
		Timeout:                  app.TKGTimeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		ClusterClientFactory:     clusterclient.NewClusterClientFactory(),
		FeatureFlagClient:        getFeatureFlagClient(),
	})
	if err != nil {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	// setting the below configuration to tkgClient to be used during Azure mc creation but not saving them to tkg config
	if params.Params.ResourceGroup == "" {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New("azure resource group name cannot be empty")))
	}
	app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureResourceGroup, params.Params.ResourceGroup)

	if params.Params.VnetResourceGroup == "" {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New("azure vnet resource group name cannot be empty")))
	}
	app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureVnetResourceGroup, params.Params.VnetResourceGroup)

	if params.Params.VnetName == "" {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New("azure vnet name cannot be empty")))
	}
	app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureVnetName, params.Params.VnetName)

	if params.Params.ControlPlaneSubnet == "" {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New("azure controlplane subnet name cannot be empty")))
	}
	app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureControlPlaneSubnet, params.Params.ControlPlaneSubnet)

	if params.Params.WorkerNodeSubnet == "" {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New("azure node subnet name cannot be empty")))
	}
	app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureWorkerSubnet, params.Params.WorkerNodeSubnet)

	if params.Params.VnetCidr != "" { // create new vnet
		app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureVnetCidr, params.Params.VnetCidr)
		app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureControlPlaneSubnetCidr, params.Params.ControlPlaneSubnetCidr)
		app.TKGConfigReaderWriter.Set(constants.ConfigVariableAzureWorkerNodeSubnetCidr, params.Params.WorkerNodeSubnetCidr)
	}

	app.InitOptions.InfrastructureProvider = InfrastructureProviderAzure
	app.InitOptions.ClusterName = params.Params.ClusterName
	app.InitOptions.Plan = params.Params.ControlPlaneFlavor
	app.InitOptions.Annotations = params.Params.Annotations
	app.InitOptions.Labels = params.Params.Labels
	app.InitOptions.CeipOptIn = *params.Params.CeipOptIn
	app.InitOptions.ClusterConfigFile = app.getFilePathForSavingConfig()
	if err := c.ConfigureAndValidateManagementClusterConfiguration(&app.InitOptions, false); err != nil {
		return azure.NewCreateAzureRegionalClusterInternalServerError().WithPayload(Err(errors.New(err.Message)))
	}
	go app.StartSendingLogsToUI()
	go func() {
		err := c.InitRegion(&app.InitOptions)
		if err != nil {
			log.Error(err, "unable to set up management cluster, ")
		} else {
			log.Infof("\nManagement cluster created!\n\n")
			log.Info("\nYou can now create your first workload cluster by running the following:\n\n")
			log.Info("  tanzu cluster create [name] -f [file]\n\n")
			err = pluginmanager.SyncPlugins()
			if err != nil {
				log.Warningf("unable to sync plugins after management cluster create. Please run `tanzu plugin sync` command manually to install/update plugins")
			}
			// wait for the logs to be dispatched to UI before exit
			time.Sleep(sleepTimeForLogsPropogation)
			// exit the BE server on success
			os.Exit(0)
		}
	}()

	return azure.NewCreateAzureRegionalClusterOK().WithPayload("started creating regional cluster")
}

// CreateDockerRegionalCluster creates docker management cluster
func (app *App) CreateDockerRegionalCluster(params docker.CreateDockerRegionalClusterParams) middleware.Responder {
	dockerConfig, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewDockerConfig(params.Params)
	if err != nil {
		return docker.NewCreateDockerRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, dockerConfig)
	if err != nil {
		return docker.NewCreateDockerRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	allClients, err := clientcreator.CreateAllClients(app.AppConfig, app.TKGConfigReaderWriter)
	if err != nil {
		return docker.NewCreateDockerRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	c, err := client.New(client.Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             app.AppConfig.TKGConfigDir,
		Timeout:                  app.TKGTimeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		ClusterClientFactory:     clusterclient.NewClusterClientFactory(),
		FeatureFlagClient:        getFeatureFlagClient(),
	})
	if err != nil {
		return docker.NewCreateDockerRegionalClusterInternalServerError().WithPayload(Err(err))
	}

	app.InitOptions.InfrastructureProvider = InfrastructureProviderDocker
	app.InitOptions.ClusterName = params.Params.ClusterName
	app.InitOptions.Plan = "dev"
	app.InitOptions.Annotations = params.Params.Annotations
	app.InitOptions.Labels = params.Params.Labels
	app.InitOptions.ClusterConfigFile = app.getFilePathForSavingConfig()

	if err := c.ConfigureAndValidateManagementClusterConfiguration(&app.InitOptions, false); err != nil {
		return docker.NewCreateDockerRegionalClusterInternalServerError().WithPayload(Err(errors.New(err.Message)))
	}

	go app.StartSendingLogsToUI()
	go func() {
		err := c.InitRegion(&app.InitOptions)
		if err != nil {
			log.Error(err, "unable to set up management cluster, ")
		} else {
			log.Infof("\nManagement cluster created!\n\n")
			log.Info("\nYou can now create your first workload cluster by running the following:\n\n")
			log.Info("  tanzu cluster create [name] -f [file]\n\n")
			err = pluginmanager.SyncPlugins()
			if err != nil {
				log.Warningf("unable to sync plugins after management cluster create. Please run `tanzu plugin sync` command manually to install/update plugins")
			}
			// wait for the logs to be dispatched to UI before exit
			time.Sleep(sleepTimeForLogsPropogation)
			// exit the BE server on success
			os.Exit(0)
		}
	}()

	return docker.NewCreateDockerRegionalClusterOK().WithPayload("started creating regional cluster")
}

func getFeatureFlagClient() client.FeatureFlagClient {
	featureFlagClient, err := config.GetClientConfig()
	if err != nil {
		featureFlagClient = &configapi.ClientConfig{}
	}
	return featureFlagClient
}
