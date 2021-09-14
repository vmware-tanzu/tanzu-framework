// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/swag"
	"github.com/juju/fslock"
	"github.com/pkg/errors"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/kind"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/yamlprocessor"
)

const (
	regionalClusterNamePrefix = "tkg-mgmt"
	defaultTkgNamespace       = "tkg-system"
)

const (
	statusRunning       = "running"
	statusFailed        = "failed"
	statusSuccessful    = "successful"
	vsphereVersionError = "the minimum vSphere version supported by Tanzu Kubernetes Grid is vSphere 6.7u3, please upgrade vSphere and try again"
)

// management cluster init step constants
const (
	StepConfigPrerequisite                 = "Configure prerequisite"
	StepValidateConfiguration              = "Validate configuration"
	StepGenerateClusterConfiguration       = "Generate cluster configuration"
	StepSetupBootstrapCluster              = "Setup bootstrap cluster"
	StepInstallProvidersOnBootstrapCluster = "Install providers on bootstrap cluster"
	StepCreateManagementCluster            = "Create management cluster"
	StepInstallProvidersOnRegionalCluster  = "Install providers on management cluster"
	StepMoveClusterAPIObjects              = "Move cluster-api objects from bootstrap cluster to management cluster"
	StepRegisterWithTMC                    = "Register management cluster with Tanzu Mission Control"
)

const (
	tmcFeatureFlag     = "tmcRegistration"
	cniFeatureFlag     = "cni"
	editionFeatureFlag = "edition"
)

// InitRegionSteps management cluster init step sequence
var InitRegionSteps []string = []string{
	StepConfigPrerequisite,
	StepValidateConfiguration,
	StepGenerateClusterConfiguration,
	StepSetupBootstrapCluster,
	StepInstallProvidersOnBootstrapCluster,
	StepCreateManagementCluster,
	StepInstallProvidersOnRegionalCluster,
	StepMoveClusterAPIObjects,
}

// InitRegionDryRun run management cluster dry-run
func (c *TkgClient) InitRegionDryRun(options *InitRegionOptions) ([]byte, error) {
	var (
		regionalConfigBytes []byte
		err                 error
	)
	// Obtain management cluster configuration of a provided flavor
	if regionalConfigBytes, options.ClusterName, err = c.BuildRegionalClusterConfiguration(options); err != nil {
		return nil, errors.Wrap(err, "unable to build management cluster configuration")
	}

	return regionalConfigBytes, nil
}

// InitRegion create management cluster
func (c *TkgClient) InitRegion(options *InitRegionOptions) error { //nolint:funlen,gocyclo
	var err error
	var regionalConfigBytes []byte
	var isSuccessful bool = false
	var isStartedRegionalClusterCreation bool = false
	var isBootstrapClusterCreated bool = false
	var bootstrapClusterName string
	var regionContext region.RegionContext
	var filelock *fslock.Lock

	bootstrapClusterKubeconfigPath, err := getTKGKubeConfigPath(false)
	if err != nil {
		return err
	}
	if options.TmcRegistrationURL != "" {
		InitRegionSteps = append(InitRegionSteps, StepRegisterWithTMC)
	}
	log.SendProgressUpdate(statusRunning, StepValidateConfiguration, InitRegionSteps)

	log.Info("Validating configuration...")
	defer func() {
		if regionContext != (region.RegionContext{}) {
			filelock, err = utils.GetFileLockWithTimeOut(filepath.Join(c.tkgConfigDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
			if err != nil {
				log.Warningf("cannot acquire lock for updating management cluster configuration, %s", err.Error())
			}
			err := c.regionManager.SaveRegionContext(regionContext)
			if err != nil {
				log.Warningf("Unable to persist management cluster %s info to tkg config", regionContext.ClusterName)
			}

			err = c.regionManager.SetCurrentContext(regionContext.ClusterName, regionContext.ContextName)
			if err != nil {
				log.Warningf("Unable to use context %s as current tkg context", regionContext.ContextName)
			}
			if err := filelock.Unlock(); err != nil {
				log.Warningf("unable to release lock for updating management cluster configuration, %s", err.Error())
			}
		}

		if isSuccessful {
			log.SendProgressUpdate(statusSuccessful, "", InitRegionSteps)
		} else {
			log.SendProgressUpdate(statusFailed, "", InitRegionSteps)
		}

		// if management cluster creation failed after bootstrap kind cluster was successfully created
		if !isSuccessful && isStartedRegionalClusterCreation {
			c.displayHelpTextOnFailure(options, isBootstrapClusterCreated, bootstrapClusterKubeconfigPath)
			return
		}

		if isBootstrapClusterCreated {
			if err := c.teardownKindCluster(bootstrapClusterName, bootstrapClusterKubeconfigPath, options.UseExistingCluster); err != nil {
				log.Warning(err.Error())
			}
		}
		_ = utils.DeleteFile(bootstrapClusterKubeconfigPath)
	}()

	if customImageRepo, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository); err != nil && customImageRepo != "" && tkgconfighelper.IsCustomRepository(customImageRepo) {
		log.Infof("Using custom image repository: %s", customImageRepo)
	}

	// validate docker only if user is not using an existing cluster
	// Note: Validating in client code as well to cover the usecase where users use client code instead of command line.
	if err := c.ValidatePrerequisites(!options.UseExistingCluster, true); err != nil {
		return err
	}

	log.Infof("Using infrastructure provider %s", options.InfrastructureProvider)
	log.SendProgressUpdate(statusRunning, StepGenerateClusterConfiguration, InitRegionSteps)
	log.Info("Generating cluster configuration...")

	// Obtain management cluster configuration of a provided flavor
	if regionalConfigBytes, options.ClusterName, err = c.BuildRegionalClusterConfiguration(options); err != nil {
		return errors.Wrap(err, "unable to build management cluster configuration")
	}

	log.SendProgressUpdate(statusRunning, StepSetupBootstrapCluster, InitRegionSteps)
	log.Info("Setting up bootstrapper...")
	// Ensure bootstrap cluster and copy boostrap cluster kubeconfig to ~/kube-tkg directory
	if bootstrapClusterName, err = c.ensureKindCluster(options.Kubeconfig, options.UseExistingCluster, bootstrapClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to create bootstrap cluster")
	}

	isBootstrapClusterCreated = true
	log.Infof("Bootstrapper created. Kubeconfig: %s", bootstrapClusterKubeconfigPath)
	bootStrapClusterClient, err := clusterclient.NewClient(bootstrapClusterKubeconfigPath, "", clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get bootstrap cluster client")
	}

	// configure variables required to deploy providers
	if err := c.configureVariablesForProvidersInstallation(nil); err != nil {
		return errors.Wrap(err, "unable to configure variables for provider installation")
	}

	log.SendProgressUpdate(statusRunning, StepInstallProvidersOnBootstrapCluster, InitRegionSteps)
	log.Info("Installing providers on bootstrapper...")
	// Initialize bootstrap cluster with providers
	if err = c.InitializeProviders(options, bootStrapClusterClient, bootstrapClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to initialize providers")
	}

	isStartedRegionalClusterCreation = true

	targetClusterNamespace := defaultTkgNamespace
	if options.Namespace != "" {
		targetClusterNamespace = options.Namespace
	}

	log.SendProgressUpdate(statusRunning, StepCreateManagementCluster, InitRegionSteps)
	log.Info("Start creating management cluster...")
	err = c.DoCreateCluster(bootStrapClusterClient, options.ClusterName, targetClusterNamespace, string(regionalConfigBytes))
	if err != nil {
		return errors.Wrap(err, "unable to create management cluster")
	}

	// save this context to tkg config incase the management cluster creation fails
	bootstrapClusterContext := "kind-" + bootstrapClusterName
	if options.UseExistingCluster {
		bootstrapClusterContext, err = getCurrentContextFromDefaultKubeConfig()
		if err != nil {
			return err
		}
	}
	regionContext = region.RegionContext{ClusterName: options.ClusterName, ContextName: bootstrapClusterContext, SourceFilePath: bootstrapClusterKubeconfigPath, Status: region.Failed}

	kubeConfigBytes, err := c.WaitForClusterInitializedAndGetKubeConfig(bootStrapClusterClient, options.ClusterName, targetClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to wait for cluster and get the cluster kubeconfig")
	}

	regionalClusterKubeconfigPath, err := getTKGKubeConfigPath(true)
	if err != nil {
		return err
	}
	// put a filelock to ensure mutual exclusion on updating kubeconfig
	filelock, err = utils.GetFileLockWithTimeOut(filepath.Join(c.tkgConfigDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return errors.Wrap(err, "cannot acquire lock for updating management cluster kubeconfig")
	}

	mergeFile := getDefaultKubeConfigFile()
	log.Infof("Saving management cluster kubeconfig into %s", mergeFile)
	// merge the management cluster kubeconfig into user input kubeconfig path/default kubeconfig path
	err = MergeKubeConfigWithoutSwitchContext(kubeConfigBytes, mergeFile)
	if err != nil {
		return errors.Wrap(err, "unable to merge management cluster kubeconfig")
	}

	// merge the management cluster kubeconfig into tkg managed kubeconfig
	kubeContext, err := MergeKubeConfigAndSwitchContext(kubeConfigBytes, regionalClusterKubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to save management cluster kubeconfig to TKG managed kubeconfig")
	}

	if err := filelock.Unlock(); err != nil {
		log.Warningf("cannot acquire lock for updating management cluster kubeconfigconfig, reason: %v", err)
	}

	regionalClusterClient, err := clusterclient.NewClient(regionalClusterKubeconfigPath, kubeContext, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get management cluster client")
	}

	log.SendProgressUpdate(statusRunning, StepInstallProvidersOnRegionalCluster, InitRegionSteps)
	log.Info("Installing providers on management cluster...")
	if err = c.InitializeProviders(options, regionalClusterClient, regionalClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to initialize providers on management cluster")
	}

	if err := regionalClusterClient.PatchClusterAPIAWSControllersToUseEC2Credentials(); err != nil {
		return err
	}

	log.Info("Waiting for the management cluster to get ready for move...")
	if err := c.WaitForClusterReadyForMove(bootStrapClusterClient, options.ClusterName, targetClusterNamespace); err != nil {
		return errors.Wrap(err, "unable to wait for cluster getting ready for move")
	}

	log.Info("Waiting for addons installation...")
	if err := c.WaitForAddons(waitForAddonsOptions{
		regionalClusterClient: bootStrapClusterClient,
		workloadClusterClient: regionalClusterClient,
		clusterName:           options.ClusterName,
		namespace:             options.Namespace,
		waitForCNI:            true,
	}); err != nil {
		return errors.Wrap(err, "error waiting for addons to get installed")
	}

	log.SendProgressUpdate(statusRunning, StepMoveClusterAPIObjects, InitRegionSteps)
	log.Info("Moving all Cluster API objects from bootstrap cluster to management cluster...")
	// Move all Cluster API objects from bootstrap cluster to created to management cluster for all namespaces
	if err = c.MoveObjects(bootstrapClusterKubeconfigPath, regionalClusterKubeconfigPath, targetClusterNamespace); err != nil {
		return errors.Wrap(err, "unable to move Cluster API objects from bootstrap cluster to management cluster")
	}

	regionContext = region.RegionContext{ClusterName: options.ClusterName, ContextName: kubeContext, SourceFilePath: regionalClusterKubeconfigPath, Status: region.Success}

	err = c.PatchClusterInitOperations(regionalClusterClient, options, targetClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to patch cluster object")
	}

	// install TMC agent workloads on the management cluster
	if options.TmcRegistrationURL != "" {
		if err = registerWithTmc(options.TmcRegistrationURL, regionalClusterClient); err != nil {
			log.Error(err, "Failed to register management cluster to Tanzu Mission Control")

			log.Warningf("\nTo attach the management cluster to Tanzu Mission Control:")
			log.Warningf("\ttanzu management-cluster register --tmc-registration-url %s", options.TmcRegistrationURL)
		}
	}

	providerName, _, err := ParseProviderName(options.InfrastructureProvider)
	if err != nil {
		return errors.Wrap(err, "unable to parse provider name")
	}

	// start CEIP telemetry cronjob if cluster is opt-in
	if options.CeipOptIn {
		bomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
		if err != nil {
			return errors.Wrapf(err, "failed to get default bom configuration")
		}

		httpProxy, httpsProxy, noProxy := "", "", ""
		if httpProxy, err = c.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy); err == nil && httpProxy != "" {
			httpsProxy, _ = c.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
			noProxy, err = c.getFullTKGNoProxy(providerName)
			if err != nil {
				return err
			}
		}

		if err = regionalClusterClient.AddCEIPTelemetryJob(options.ClusterName, providerName, bomConfig, "", "", httpProxy, httpsProxy, noProxy); err != nil {
			log.Error(err, "Failed to start CEIP telemetry job on management cluster")

			log.Warningf("\nTo have this cluster participate in VMware CEIP:")
			log.Warningf("\ttanzu management-cluster ceip-participation set true")
		}
	}

	log.Info("Waiting for additional components to be up and running...")
	if err := c.WaitForAddonsDeployments(regionalClusterClient); err != nil {
		return err
	}

	log.Info("Waiting for packages to be up and running...")
	if err := c.WaitForPackages(regionalClusterClient, regionalClusterClient, options.ClusterName, targetClusterNamespace); err != nil {
		log.Warningf("Warning: Management cluster is created successfully, but some packages are failing. %v", err)
	}

	log.Infof("Context set for management cluster %s as '%s'.", options.ClusterName, kubeContext)
	isSuccessful = true
	return nil
}

func registerWithTmc(url string, regionalClusterClient clusterclient.Client) error {
	log.SendProgressUpdate(statusRunning, StepRegisterWithTMC, InitRegionSteps)
	log.Info("Registering management cluster with TMC...")

	if !utils.IsValidURL(url) {
		return errors.Errorf("TMC registration URL '%s' is not valid", url)
	}

	err := regionalClusterClient.ApplyFile(url)
	if err != nil {
		return errors.Wrap(err, "failed to register management cluster to TMC")
	}

	log.Infof("Successfully registered management cluster to TMC")
	return nil
}

// PatchClusterInitOperations Patches cluster
func (c *TkgClient) PatchClusterInitOperations(regionalClusterClient clusterclient.Client, options *InitRegionOptions, targetClusterNamespace string) error {
	// Patch management cluster with the TKG version
	err := regionalClusterClient.PatchClusterObjectWithTKGVersion(options.ClusterName, targetClusterNamespace, c.tkgBomClient.GetCurrentTKGVersion())
	if err != nil {
		return errors.Wrap(err, "unable to patch TKG Version")
	}

	// Patch management cluster with the provided optional metadata
	_, err = regionalClusterClient.PatchClusterObjectWithOptionalMetadata(options.ClusterName, targetClusterNamespace, "annotations", options.Annotations)
	if err != nil {
		return errors.Wrap(err, "unable to patch optional metadata under annotations")
	}

	// Patch management cluster with the provided optional labels
	_, err = regionalClusterClient.PatchClusterObjectWithOptionalMetadata(options.ClusterName, targetClusterNamespace, "labels", options.Labels)
	if err != nil {
		return errors.Wrap(err, "unable to patch optional metadata under labels")
	}

	return err
}

// MoveObjects moves all the Cluster API objects from all the namespaces to a target management cluster.
func (c *TkgClient) MoveObjects(fromKubeconfigPath, toKubeconfigPath, namespace string) error {
	moveOptions := clusterctl.MoveOptions{
		FromKubeconfig: clusterctl.Kubeconfig{Path: fromKubeconfigPath},
		ToKubeconfig:   clusterctl.Kubeconfig{Path: toKubeconfigPath},
		Namespace:      namespace,
	}
	return c.clusterctlClient.Move(moveOptions)
}

func (c *TkgClient) ensureKindCluster(kubeconfig string, useExistingCluster bool, backupPath string) (string, error) {
	// skip if using existing cluster
	if useExistingCluster {
		if kubeconfig == "" {
			kubeconfig = getDefaultKubeConfigFile()
		}

		content, err := os.ReadFile(kubeconfig)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(backupPath, content, constants.ConfigFilePermissions)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	bomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", err
	}

	c.kindClient = kind.New(&kind.KindClusterOptions{
		KubeConfigPath:   backupPath,
		TKGConfigDir:     c.tkgConfigDir,
		Readerwriter:     c.TKGConfigReaderWriter(),
		DefaultImageRepo: bomConfig.ImageConfig.ImageRepository,
	})

	// Create kind cluster which will be used to deploy management cluster
	clusterName, err := c.kindClient.CreateKindCluster()
	if err != nil {
		return "", err
	}
	return clusterName, nil
}

func (c *TkgClient) teardownKindCluster(clusterName, kubeconfig string, useExistingCluster bool) error {
	// skip if using existing cluster
	if useExistingCluster {
		err := c.clusterctlClient.Delete(clusterctl.DeleteOptions{
			Kubeconfig: clusterctl.Kubeconfig{
				Path: kubeconfig,
			},
			DeleteAll:        true,
			IncludeNamespace: true,
			IncludeCRDs:      true,
		})
		log.V(3).Error(err, "Failed to delete resources from bootstrap cluster")
		return nil
	}

	bomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return err
	}

	if c.kindClient == nil {
		c.kindClient = kind.New(&kind.KindClusterOptions{
			KubeConfigPath:   kubeconfig,
			ClusterName:      clusterName,
			TKGConfigDir:     c.tkgConfigDir,
			DefaultImageRepo: bomConfig.ImageConfig.ImageRepository,
		})
	}

	// Delete kind cluster
	err = c.kindClient.DeleteKindCluster()
	if err != nil {
		return err
	}
	return nil
}

// InitializeProviders initializes providers
func (c *TkgClient) InitializeProviders(options *InitRegionOptions, clusterClient clusterclient.Client, kubeconfigPath string) error {
	clusterctlClientInitOptions := clusterctl.InitOptions{
		Kubeconfig:              clusterctl.Kubeconfig{Path: kubeconfigPath},
		InfrastructureProviders: []string{options.InfrastructureProvider},
		ControlPlaneProviders:   []string{options.ControlPlaneProvider},
		BootstrapProviders:      []string{options.BootstrapProvider},
		CoreProvider:            options.CoreProvider,
		TargetNamespace:         options.Namespace,
		WatchingNamespace:       options.WatchingNamespace,
	}

	componentsList, err := c.clusterctlClient.Init(clusterctlClientInitOptions)
	if err != nil {
		return errors.Errorf("%s, this can be possible because of the outbound connectivity issue. Please check deployed nodes for outbound connectivity.", err.Error())
	}

	for _, components := range componentsList {
		log.V(3).Info("installed", " Component=", components.Name(), " Type=", components.Type(), " Version=", components.Version())
	}

	if err = clusterClient.CreateNamespace(defaultTkgNamespace); err != nil {
		return errors.Wrapf(err, "Unable to create default namespace %s", defaultTkgNamespace)
	}

	// Wait for installed providers to get up and running
	waitOptions := waitForProvidersOptions{
		Kubeconfig:        options.Kubeconfig,
		TargetNamespace:   options.Namespace,
		WatchingNamespace: options.WatchingNamespace,
	}
	err = c.WaitForProviders(clusterClient, waitOptions)
	if err != nil {
		return errors.Wrap(err, "error waiting for provider components to be up and running")
	}

	return nil
}

func (c *TkgClient) configureImageTagsForProviderInstallation() error {
	// configure image tags by reading BoM file
	tkgBoMConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return errors.Wrap(err, "unable to get default BoM configuration")
	}
	configImageTag := func(configVariable, componentName, imageName string) {
		component, exists := tkgBoMConfig.Components[componentName]
		if !exists {
			log.Warningf("Warning: unable to find component '%s' under BoM", componentName)
			return
		}
		if len(component) == 0 {
			log.Warningf("Warning: component '%s' is empty", componentName)
			return
		}

		image, exists := component[0].Images[imageName]
		if !exists {
			log.Warningf("Warning: component '%s' does not have image '%s'", componentName, imageName)
			return
		}
		c.TKGConfigReaderWriter().Set(configVariable, image.Tag)
	}

	configImageTag(constants.ConfigVariableInternalKubeRBACProxyImageTag, "kube_rbac_proxy", "kubeRbacProxyControllerImageCapi")
	configImageTag(constants.ConfigVariableInternalCABPKControllerImageTag, "cluster_api", "cabpkControllerImage")
	configImageTag(constants.ConfigVariableInternalCAPIControllerImageTag, "cluster_api", "capiControllerImage")
	configImageTag(constants.ConfigVariableInternalKCPControllerImageTag, "cluster_api", "kcpControllerImage")
	configImageTag(constants.ConfigVariableInternalCAPDManagerImageTag, "cluster_api", "capdManagerImage")
	configImageTag(constants.ConfigVariableInternalCAPAManagerImageTag, "cluster_api_aws", "capaControllerImage")
	configImageTag(constants.ConfigVariableInternalCAPVManagerImageTag, "cluster_api_vsphere", "capvControllerImage")
	configImageTag(constants.ConfigVariableInternalCAPZManagerImageTag, "cluster-api-provider-azure", "capzControllerImage")
	configImageTag(constants.ConfigVariableInternalNMIImageTag, "aad-pod-identity", "nmiImage")

	return nil
}

func generateRegionalClusterName(infrastructureProvider, kubeContext string) string {
	infra := strings.Split(infrastructureProvider, ":")[0]

	if kubeContext != "" {
		return fmt.Sprintf("%s-%s-%s", regionalClusterNamePrefix, infra, kubeContext)
	}

	t := time.Now()
	return fmt.Sprintf("%s-%s-%s", regionalClusterNamePrefix, infra, t.Format("20060102150405"))
}

// BuildRegionalClusterConfiguration build management cluster configuration
func (c *TkgClient) BuildRegionalClusterConfiguration(options *InitRegionOptions) ([]byte, string, error) {
	var bytes []byte
	var err error

	if options.ClusterName == "" {
		options.ClusterName = generateRegionalClusterName(options.InfrastructureProvider, "")
	}

	namespace := options.Namespace
	if namespace == "" {
		namespace = defaultTkgNamespace
	}

	controlPlaneMachineCount, workerMachineCount := c.getMachineCountForMC(options.Plan)

	providerRepositorySource := &clusterctl.ProviderRepositorySourceOptions{
		InfrastructureProvider: options.InfrastructureProvider,
		Flavor:                 options.Plan,
	}

	clusterConfigOptions := ClusterConfigOptions{
		Kubeconfig:               clusterctl.Kubeconfig{Path: options.Kubeconfig},
		ProviderRepositorySource: providerRepositorySource,
		TargetNamespace:          namespace,
		ClusterName:              options.ClusterName,
		ControlPlaneMachineCount: swag.Int64(int64(controlPlaneMachineCount)),
		WorkerMachineCount:       swag.Int64(int64(workerMachineCount)),
	}

	if !options.DisableYTT {
		clusterConfigOptions.YamlProcessor = yamlprocessor.NewYttProcessorWithConfigDir(c.tkgConfigDir)
	}

	bytes, err = c.getClusterConfiguration(&clusterConfigOptions, true, clusterConfigOptions.ProviderRepositorySource.InfrastructureProvider)

	return bytes, options.ClusterName, err
}

func (c *TkgClient) getMachineCountForMC(plan string) (int, int) {
	// set controlplane and worker counts to default initially
	controlPlaneMachineCount := constants.DefaultDevControlPlaneMachineCount
	workerMachineCount := constants.DefaultWorkerMachineCountForManagementCluster

	switch plan {
	case constants.PlanDev:
		// use the defaults already set above
	case constants.PlanProd:
		// update controlplane count for prod plan
		controlPlaneMachineCount = constants.DefaultProdControlPlaneMachineCount
	default:
		// For custom plan use config variables to determine the count
		// Verify there is no error in retrieving this and controlplane count is odd number
		// If not provided then continue to use default values
		if cpc, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableControlPlaneMachineCount, c.TKGConfigReaderWriter()); err == nil && cpc%2 == 1 {
			controlPlaneMachineCount = cpc
		} else {
			log.Info("Using default value for CONTROL_PLANE_MACHINE_COUNT= %v. Reason: Either provided value is even or %s", err.Error())
		}
		if wc, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableWorkerMachineCount, c.TKGConfigReaderWriter()); err == nil {
			workerMachineCount = wc
		} else {
			log.Info("Using default value for WORKER_MACHINE_COUNT= %v. Reason: %s", err.Error())
		}
	}
	return controlPlaneMachineCount, workerMachineCount
}

type waitForProvidersOptions struct {
	Kubeconfig        string
	TargetNamespace   string
	WatchingNamespace string
}

// WaitForProviders checks and waits for each provider components to be up and running
func (c *TkgClient) WaitForProviders(clusterClient clusterclient.Client, options waitForProvidersOptions) error {
	var err error

	if clusterClient == nil {
		clusterClient, err = clusterclient.NewClient(options.Kubeconfig, "", clusterclient.Options{OperationTimeout: c.timeout})
		if err != nil {
			return errors.Wrap(err, "unable to get deletion cluster client")
		}
	}

	// Get the all the installed provider info
	providers := &clusterctlv1.ProviderList{}
	err = clusterClient.ListResources(providers, &crtclient.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "cannot get installed provider config")
	}
	// Wait for each provider(core-provider, bootstrap-provider, infrastructure-providers) to be up and running
	var wg sync.WaitGroup
	results := make(chan error, len(providers.Items))
	for i := range providers.Items {
		wg.Add(1)
		go func(wg *sync.WaitGroup, provider clusterctlv1.Provider) {
			defer wg.Done()
			t, err := TimedExecution(func() error {
				log.V(3).Infof("Waiting for provider %s", provider.Name)
				providerNameVersion := provider.ProviderName + ":" + provider.Version
				return c.waitForProvider(clusterClient, providerNameVersion, provider.Type, options.TargetNamespace, options.WatchingNamespace)
			})
			if err != nil {
				log.V(3).Warningf("Failed waiting for provider %v after %v", provider.Name, t)
				results <- err
			} else {
				log.V(3).Infof("Passed waiting on provider %s after %v", provider.Name, t)
			}
		}(&wg, providers.Items[i])
	}

	wg.Wait()
	close(results)
	for err := range results {
		log.Warning("Failed waiting for at least one provider, check logs for more detail.")
		return err
	}
	log.V(3).Info("Success waiting on all providers.")
	return nil
}

func (c *TkgClient) waitForProvider(clusterClient clusterclient.Client, name, providerType, targetNamespace, watchingNamespace string) error {
	providerOptions := clusterctl.ComponentsOptions{TargetNamespace: targetNamespace, WatchingNamespace: watchingNamespace}
	// get the provider component from clusterctl
	providerComponents, err := c.clusterctlClient.GetProviderComponents(name, clusterctlv1.ProviderType(providerType), providerOptions)
	if err != nil {
		return errors.Wrap(err, "Error getting provider component")
	}

	var controllerName string
	var namespace string

	objs := append(providerComponents.InstanceObjs(), providerComponents.SharedObjs()...)

	// get the deployment name and namespace from the provider
	for _, object := range objs {
		if object.GetKind() == "Deployment" {
			controllerName = object.GetName()
			namespace = object.GetNamespace()

			err := clusterClient.WaitForDeployment(controllerName, namespace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *TkgClient) displayHelpTextOnFailure(options *InitRegionOptions,
	isBootstrapClusterCreated bool, bootstrapClusterKubeconfigPath string) {

	log.Warningf("\n\nFailure while deploying management cluster, Here are some steps to investigate the cause:\n")
	log.Warningf("\nDebug:")
	log.Warningf("    kubectl get po,deploy,cluster,kubeadmcontrolplane,machine,machinedeployment -A --kubeconfig %s", bootstrapClusterKubeconfigPath)
	log.Warningf("    kubectl logs deployment.apps/<deployment-name> -n <deployment-namespace> manager --kubeconfig %s", bootstrapClusterKubeconfigPath)

	if !options.UseExistingCluster && isBootstrapClusterCreated {
		log.Warningf("\nTo clean up the resources created by the management cluster:")
		log.Warningf("	  tanzu management-cluster delete")
	}
}

// ParseHiddenArgsAsFeatureFlags adds the hidden flags from InitRegionOptions as enabled feature flags
func (c *TkgClient) ParseHiddenArgsAsFeatureFlags(options *InitRegionOptions) {
	if options.TmcRegistrationURL != "" {
		options.FeatureFlags = c.safelyAddFeatureFlag(options.FeatureFlags, tmcFeatureFlag, options.TmcRegistrationURL)
	}
	if options.CniType != "" {
		options.FeatureFlags = c.safelyAddFeatureFlag(options.FeatureFlags, cniFeatureFlag, options.CniType)
	}
	if options.Edition != "" {
		options.FeatureFlags = c.safelyAddFeatureFlag(options.FeatureFlags, editionFeatureFlag, options.Edition)
	}
}

// SaveFeatureFlags saves the feature flags to the config file via featuresClient
func (c *TkgClient) SaveFeatureFlags(featureFlags map[string]string) error {
	err := c.featuresClient.WriteFeatureFlags(featureFlags)
	if err != nil {
		return errors.Wrap(err, "failed to write feature flags")
	}
	return nil
}

// safelyAddFeatureFlag adds an entry to the feature flag map and handles if map is nil
func (c *TkgClient) safelyAddFeatureFlag(featureFlags map[string]string, feature, value string) map[string]string {
	if featureFlags != nil {
		featureFlags[feature] = value
	} else {
		featureFlags = map[string]string{feature: value}
	}
	return featureFlags
}
