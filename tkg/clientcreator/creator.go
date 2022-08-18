// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package clientcreator defines functions to create clients.
package clientcreator

import (
	"github.com/pkg/errors"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/features"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
)

// Clients is a combination structure of clients
type Clients struct {
	ClusterCtlClient         clusterctl.Client
	ConfigClient             tkgconfigreaderwriter.Client
	RegionManager            region.Manager
	FeaturesClient           features.Client
	TKGConfigProvidersClient tkgconfigproviders.Client
	TKGBomClient             tkgconfigbom.Client
	TKGConfigUpdaterClient   tkgconfigupdater.Client
	TKGConfigPathsClient     tkgconfigpaths.Client
	FeatureFlagClient        *configv1alpha1.ClientConfig
}

// CreateAllClients creates all clients and returns Clients struct
func CreateAllClients(appConfig types.AppConfig, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (Clients, error) {
	var err error
	tkgConfigPathsClient := tkgconfigpaths.New(appConfig.TKGConfigDir)

	tkgConfigFile := appConfig.TKGSettingsFile
	if tkgConfigFile == "" {
		tkgConfigFile, err = tkgConfigPathsClient.GetTKGConfigPath()
		if err != nil {
			return Clients{}, err
		}
	}

	// Create tkg configuration client, reads tkg and cluster configuration
	// file and sets configuration values to viper store
	var configClient tkgconfigreaderwriter.Client
	if tkgConfigReaderWriter == nil {
		configClient, err = tkgconfigreaderwriter.New(tkgConfigFile)
	} else {
		configClient, err = tkgconfigreaderwriter.NewWithReaderWriter(tkgConfigReaderWriter)
	}
	if err != nil {
		return Clients{}, errors.Wrap(err, "unable to create tkg config Client")
	}

	tkgConfigUpdaterClient := tkgconfigupdater.New(appConfig.TKGConfigDir, appConfig.ProviderGetter, configClient.TKGConfigReaderWriter())
	tkgBomClient := tkgconfigbom.New(appConfig.TKGConfigDir, configClient.TKGConfigReaderWriter())
	tkgConfigProvidersClient := tkgconfigproviders.New(appConfig.TKGConfigDir, configClient.TKGConfigReaderWriter())

	// Create clusterctl client
	clusterctlClient, err := clusterctl.New("", clusterctl.InjectConfig(configClient.ClusterConfigClient()))
	if err != nil {
		return Clients{}, errors.Wrap(err, "unable to initialize clusterctl client with config path")
	}

	regionManager, err := appConfig.CustomizerOptions.RegionManagerFactory.CreateManager(tkgConfigFile)
	if err != nil {
		return Clients{}, errors.Wrap(err, "unable to initialize management cluster manager with config path")
	}

	// create new features client, defaults config file path to ~/.tkg/features.yaml
	// This client is used to activate/deactivate features based on this features.yaml file
	featuresClient, err := features.New(appConfig.TKGConfigDir, "")
	if err != nil {
		return Clients{}, errors.Wrap(err, "failed to create features client")
	}

	featureFlagClient, err := config.GetClientConfig()
	if err != nil {
		return Clients{}, errors.Wrap(err, "failed to get client config")
	}

	return Clients{
		ClusterCtlClient:         clusterctlClient,
		ConfigClient:             configClient,
		FeaturesClient:           featuresClient,
		RegionManager:            regionManager,
		TKGBomClient:             tkgBomClient,
		TKGConfigProvidersClient: tkgConfigProvidersClient,
		TKGConfigUpdaterClient:   tkgConfigUpdaterClient,
		TKGConfigPathsClient:     tkgConfigPathsClient,
		FeatureFlagClient:        featureFlagClient,
	}, nil
}
