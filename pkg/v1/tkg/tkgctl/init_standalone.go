/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tkgctl

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server"
)

// InitRegionOptions init region options
//type InitRegionOptions struct {
//	ClusterConfigFile string
//
//	Plan                        string
//	UI                          bool
//	ClusterName                 string
//	UseExistingCluster          bool
//	CoreProvider                string
//	BootstrapProvider           string
//	InfrastructureProvider      string
//	ControlPlaneProvider        string
//	Namespace                   string
//	WatchingNamespace           string
//	Size                        string
//	ControlPlaneSize            string
//	WorkerSize                  string
//	TmcRegistrationURL          string
//	CeipOptIn                   string
//	CniType                     string
//	FeatureFlags                map[string]string
//	EnableTKGSOnVsphere7        bool
//	DeployTKGonVsphere7         bool
//	Bind                        string
//	Browser                     string
//	VsphereControlPlaneEndpoint string
//	SkipPrompt                  bool
//	Timeout                     time.Duration
//}

// InitStandalone initializes standalone cluster
func (t *tkgctl) InitStandalone(options InitRegionOptions) error {
	log.Infof("\nSTANDALONE INIT YO YO YO...")
	var err error

	log.Infof("\nloading cluster config file at %s", options.ClusterConfigFile)
	options.ClusterConfigFile, err = t.ensureClusterConfigFile(options.ClusterConfigFile)
	if err != nil {
		return err
	}

	options.CoreProvider, options.BootstrapProvider, options.ControlPlaneProvider, err = t.tkgBomClient.GetDefaultClusterAPIProviders()
	if err != nil {
		return err
	}
	log.Infof("\nloaded coreprovider: %s, bootstrapprovider: %s, and cp-provider: %s", options.CoreProvider, options.BootstrapProvider, options.ControlPlaneProvider)

	err = t.configureInitManagementClusterOptionsFromConfigFile(&options)
	if err != nil {
		return err
	}

	ceipOptIn, err := strconv.ParseBool(options.CeipOptIn)
	if err != nil {
		ceipOptIn = true
	}

	// init requires minimum 15 minutes timeout
	minTimeoutReq := 15 * time.Minute
	if options.Timeout < minTimeoutReq {
		log.V(6).Infof("timeout duration of at least 15 minutes is required, using default timeout %v", constants.DefaultLongRunningOperationTimeout)
		options.Timeout = constants.DefaultLongRunningOperationTimeout
	}
	defer t.restoreAfterSettingTimeout(options.Timeout)()

	// validate docker only if user is not using an existing cluster
	log.Infof("\nValidating the pre-requisites...")
	if err := t.tkgClient.ValidatePrerequisites(!options.UseExistingCluster, true); err != nil {
		return err
	}

	nodeSizeOptions := client.NodeSizeOptions{
		Size:             options.Size,
		ControlPlaneSize: options.ControlPlaneSize,
		WorkerSize:       options.WorkerSize,
	}

	// DYV
	//optionsIR := t.populateClientInitRegionOptions(&options, nodeSizeOptions, ceipOptIn)
	optionsIR := client.InitRegionOptions{
		ClusterConfigFile:           options.ClusterConfigFile,
		Kubeconfig:                  t.kubeconfig,
		Plan:                        options.Plan,
		LaunchUI:                    options.UI,
		ClusterName:                 options.ClusterName,
		UseExistingCluster:          options.UseExistingCluster,
		InfrastructureProvider:      options.InfrastructureProvider,
		ControlPlaneProvider:        options.ControlPlaneProvider,
		BootstrapProvider:           options.BootstrapProvider,
		CoreProvider:                options.CoreProvider,
		Namespace:                   options.Namespace,
		WatchingNamespace:           options.WatchingNamespace,
		NodeSizeOptions:             nodeSizeOptions,
		TmcRegistrationURL:          options.TmcRegistrationURL,
		CeipOptIn:                   ceipOptIn,
		CniType:                     options.CniType,
		FeatureFlags:                options.FeatureFlags,
		VsphereControlPlaneEndpoint: options.VsphereControlPlaneEndpoint,
		Edition:                     "tce-standalone",
	}

	// take the provided hidden flags and enable the related feature flags
	t.tkgClient.ParseHiddenArgsAsFeatureFlags(&optionsIR)

	// save the feature flags to the config file before launching UI
	err = t.tkgClient.SaveFeatureFlags(optionsIR.FeatureFlags)
	if err != nil {
		log.Error(err, "Failed to save feature flag options.")
	}

	if optionsIR.LaunchUI {
		err := server.Serve(optionsIR, t.appConfig, t.TKGConfigReaderWriter(), options.Timeout, options.Bind, options.Browser)
		if err != nil {
			return errors.Wrap(err, "failed to start Kickstart UI")
		}
	} else {
		providerName, _, err := client.ParseProviderName(optionsIR.InfrastructureProvider)
		if err != nil {
			return errors.Wrap(err, "unable to parse provider name")
		}
		if providerName == "vsphere" {
			if err = t.verifyThumbprint(options.SkipPrompt); err != nil {
				return err
			}
			vcClient, err := t.tkgClient.GetVSphereEndpoint(nil)
			if err != nil {
				return errors.Wrap(err, "unable to verify vSphere credentials")
			}
			validateErr := client.ValidateVSphereVersion(vcClient)
			if validateErr != nil {
				switch validateErr.Code {
				case client.PacificInVC7ErrorCode:

					isVsphereWithKubernetes := true
					shouldContinueDeployment, err := t.validationActionForVSphereVersion(isVsphereWithKubernetes, &options)
					if !shouldContinueDeployment {
						return err
					}
				case client.PacificNotInVC7ErrorCode:

					isVsphereWithKubernetes := false
					shouldContinueDeployment, err := t.validationActionForVSphereVersion(isVsphereWithKubernetes, &options)
					if !shouldContinueDeployment {
						return err
					}
				default:
					return errors.Wrap(validateErr, "configuration validation failed")
				}
			}

		}

		validateErr := t.tkgClient.ConfigureAndValidateManagementClusterConfiguration(&optionsIR, false)
		if validateErr != nil {
			return errors.Wrap(validateErr, "configuration validation failed")
		}

		log.Infof("\nSetting up standalone cluster...\n")
		err = t.tkgClient.InitStandaloneRegion(&optionsIR)
		if err != nil {
			return errors.Wrap(err, "unable to set up management cluster")
		}

		log.Infof("\nStandalone cluster created!\n\n")
	}

	return nil
}
