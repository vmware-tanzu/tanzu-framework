// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server"
)

// InitRegionOptions init region options
type InitRegionOptions struct {
	ClusterConfigFile           string
	Plan                        string
	ClusterName                 string
	CoreProvider                string
	BootstrapProvider           string
	InfrastructureProvider      string
	ControlPlaneProvider        string
	Namespace                   string
	WatchingNamespace           string
	Size                        string
	ControlPlaneSize            string
	WorkerSize                  string
	CeipOptIn                   string
	CniType                     string
	Bind                        string
	Browser                     string
	VsphereControlPlaneEndpoint string
	Edition                     string
	FeatureFlags                map[string]string
	Timeout                     time.Duration
	UI                          bool
	UseExistingCluster          bool
	EnableTKGSOnVsphere7        bool
	DeployTKGonVsphere7         bool
	SkipPrompt                  bool
	GenerateOnly                bool
}

const (
	TCEBuildEdition = "tce"
)

// Init initializes tkg management cluster
//
//nolint:gocritic,gocyclo,funlen
func (t *tkgctl) Init(options InitRegionOptions) error {
	var err error
	options.ClusterConfigFile, err = t.ensureClusterConfigFile(options.ClusterConfigFile)
	if err != nil {
		return err
	}

	err = ensureConfigImages(t.configDir, t.tkgConfigUpdaterClient)
	if err != nil {
		return err
	}

	err = t.configureInitManagementClusterOptionsFromConfigFile(&options)
	if err != nil {
		return err
	}

	ceipOptIn, err := strconv.ParseBool(options.CeipOptIn)
	if err != nil {
		if options.Edition == TCEBuildEdition {
			ceipOptIn = false
		} else {
			ceipOptIn = true
		}
	}

	if logPath, err := t.getAuditLogPath(options.ClusterName); err == nil {
		log.SetAuditLog(logPath)
	}

	options.CoreProvider, options.BootstrapProvider, options.ControlPlaneProvider, err = t.tkgBomClient.GetDefaultClusterAPIProviders()
	if err != nil {
		return err
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

	isInputFileClusterClassBased, err := t.processManagementClusterInputFile(&options)
	if err == nil && isInputFileClusterClassBased {
		return errors.New("creating management cluster using ClusterClass based configuration file is not yet supported. Please use legacy configuration file when creating management cluster")
	}

	if err := t.tkgClient.ValidatePrerequisites(!options.UseExistingCluster, true); err != nil {
		return err
	}

	nodeSizeOptions := client.NodeSizeOptions{
		Size:             options.Size,
		ControlPlaneSize: options.ControlPlaneSize,
		WorkerSize:       options.WorkerSize,
	}

	optionsIR := t.populateClientInitRegionOptions(&options, nodeSizeOptions, ceipOptIn)
	optionsIR.IsInputFileClusterClassBased = isInputFileClusterClassBased

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
			err := t.verifyThumbprint(options.SkipPrompt)
			if err != nil {
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
					shouldContinueDeployment, err := t.validationActionForVSphereVersion(true, &options)
					if !shouldContinueDeployment {
						return err
					}
				case client.PacificNotInVC7ErrorCode:
					shouldContinueDeployment, err := t.validationActionForVSphereVersion(false, &options)
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

		if options.GenerateOnly {
			if manifest, err := t.tkgClient.InitRegionDryRun(&optionsIR); err != nil {
				return errors.Wrap(err, "failed generating management cluster manifest")
			} else if _, err = os.Stdout.Write(manifest); err != nil {
				return errors.Wrap(err, "Failed writing management cluster manifest to stdout")
			}
		} else {
			log.Infof("\nSetting up management cluster...\n")
			err = t.tkgClient.InitRegion(&optionsIR)
			if err != nil {
				return errors.Wrap(err, "unable to set up management cluster")
			}
			logManagementCreationSuccess()
		}
	}
	return nil
}

// validationActionForVSphereVersion acts on vSphere version validations
func (t *tkgctl) validationActionForVSphereVersion(isVsphereWithKubernetes bool, options *InitRegionOptions) (bool, error) {
	shouldContinueDeployment := false
	vcHost, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereServer)
	if err != nil {
		return shouldContinueDeployment, errors.Errorf("failed to get %s", constants.ConfigVariableVsphereServer)
	}
	url := fmt.Sprintf("https://%s/ui/app/workload-platform/", vcHost)

	// display the warning message
	if isVsphereWithKubernetes {
		log.Warning(Warningvsphere7WithPacific)
	} else {
		log.Warning(Warningvsphere7WithoutPacific)
	}

	// action based on command line option with out prompt
	if options.EnableTKGSOnVsphere7 {
		shouldContinueDeployment = false
		if err = open.Run(url); err != nil {
			log.Error(err, "unable to open browser")
			return shouldContinueDeployment, errors.Wrap(err, "unable to open browser")
		}
		log.Info("Continue configuring Tanzu Kubernetes Grid Service on vSphere 7.0 using browser...")
		return shouldContinueDeployment, nil
	}

	if options.DeployTKGonVsphere7 {
		log.Info("Deploying TKG management cluster on vSphere 7.0 ...")
		shouldContinueDeployment = true
		return shouldContinueDeployment, nil
	}
	log.Warning("Note: To skip the prompts and directly deploy a non-integrated Tanzu Kubernetes Grid instance on vSphere 7.0, you can set the 'DEPLOY_TKG_ON_VSPHERE7' configuration variable to 'true' \n\n")
	// prompt user and ask for confirmation
	err = askForConfirmation("Do you want to configure vSphere with Tanzu?")
	if err == nil {
		shouldContinueDeployment = false
		if err := open.Run(url); err != nil {
			log.Error(err, "unable to open browser")
			return shouldContinueDeployment, errors.Wrap(err, "unable to open browser")
		}
		log.Info("Continue configuring Tanzu Kubernetes Grid Service on vSphere 7.0 using browser...")
		return shouldContinueDeployment, nil
	}

	err = askForConfirmation("Would you like to deploy a non-integrated Tanzu Kubernetes Grid management cluster on vSphere 7.0?")
	if err != nil {
		shouldContinueDeployment = false
		return shouldContinueDeployment, nil
	}

	log.Info("Deploying TKG management cluster on vSphere 7.0 ...")
	shouldContinueDeployment = true
	return shouldContinueDeployment, nil
}

//nolint:gocyclo
func (t *tkgctl) configureInitManagementClusterOptionsFromConfigFile(iro *InitRegionOptions) error {
	// set ClusterName from config variable
	if iro.ClusterName == "" {
		clusterName, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
		if err == nil {
			iro.ClusterName = clusterName
		}
	}

	// set InfrastructureProvider from config variable
	if iro.InfrastructureProvider == "" {
		infraProvider, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider)
		if err == nil {
			iro.InfrastructureProvider = infraProvider
		}
	}

	// set Size variable from config File
	if iro.Size == "" {
		size, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize)
		if err == nil {
			iro.Size = size
		}
	}

	// set ControlPlaneSize variable from config File
	if iro.ControlPlaneSize == "" {
		controlPlaneSize, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize)
		if err == nil {
			iro.ControlPlaneSize = controlPlaneSize
		}
	}

	// set WorkerSize variable from config File
	if iro.WorkerSize == "" {
		workerSize, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize)
		if err == nil {
			iro.WorkerSize = workerSize
		}
	}

	// set Plan from config variable
	if iro.Plan == "" {
		plan, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan)
		if err == nil {
			iro.Plan = plan
		}
	}

	// set BuildEdition from config variable
	if iro.Edition == "" {
		edition, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableBuildEdition)
		if err == nil {
			iro.Edition = edition
		}
	}

	// set vSphereControlPlaneEndpoint from config variable
	if iro.VsphereControlPlaneEndpoint == "" {
		vSphereControlPlaneEndpoint, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
		if err == nil {
			iro.VsphereControlPlaneEndpoint = vSphereControlPlaneEndpoint
		}
	}

	// set DeployTKGonVsphere7 from config variable
	if !iro.DeployTKGonVsphere7 {
		deployTKGonVsphere7, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableDeployTKGOnVsphere7)
		if err == nil {
			iro.DeployTKGonVsphere7, _ = strconv.ParseBool(deployTKGonVsphere7)
		}
	}

	// set EnableTKGSOnVsphere7 from config variable
	if !iro.EnableTKGSOnVsphere7 {
		enableTKGSOnVsphere7, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableTKGSonVsphere7)
		if err == nil {
			iro.EnableTKGSOnVsphere7, _ = strconv.ParseBool(enableTKGSOnVsphere7)
		}
	}

	// set ceip participation from config variable
	if iro.CeipOptIn == "" {
		iro.CeipOptIn = t.setCEIPOptinBasedOnConfigAndBuildEdition(iro.Edition)
	}

	log.V(5).Infof("CEIP Opt-in status: %s", iro.CeipOptIn)

	return nil
}

func (t *tkgctl) populateClientInitRegionOptions(options *InitRegionOptions, nodeSizeOptions client.NodeSizeOptions, ceipOptIn bool) client.InitRegionOptions {
	return client.InitRegionOptions{
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
		NodeSizeOptions:             nodeSizeOptions,
		CeipOptIn:                   ceipOptIn,
		CniType:                     options.CniType,
		FeatureFlags:                options.FeatureFlags,
		VsphereControlPlaneEndpoint: options.VsphereControlPlaneEndpoint,
		Edition:                     options.Edition,
	}
}

func logManagementCreationSuccess() {
	log.Infof("\nManagement cluster created!\n\n")
	log.Info("\nYou can now create your first workload cluster by running the following:\n\n")
	log.Info("  tanzu cluster create [name] -f [file]\n\n")
	log.Info("\nSome addons might be getting installed! Check their status by running the following:\n\n")
	log.Info("  kubectl get apps -A\n\n")
}
