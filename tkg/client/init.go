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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	capav1beta2 "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	capzv1beta1 "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/kind"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/yamlprocessor"
)

const (
	regionalClusterNamePrefix = "tkg-mgmt"
	defaultTkgNamespace       = "tkg-system"
)

const (
	editionAnnotationKey = "edition"
	statusRunning        = "running"
	statusFailed         = "failed"
	statusSuccessful     = "successful"
	vsphereVersionError  = "the minimum vSphere version supported by Tanzu Kubernetes Grid is vSphere 6.7u3, please upgrade vSphere and try again"
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
)

const (
	cniFeatureFlag     = "cni"
	editionFeatureFlag = "edition"
)

// InitRegionSteps management cluster init step sequence
var InitRegionSteps = []string{
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
	if regionalConfigBytes, options.ClusterName, _, err = c.BuildRegionalClusterConfiguration(options); err != nil {
		return nil, errors.Wrap(err, "unable to build management cluster configuration")
	}

	return regionalConfigBytes, nil
}

// InitRegion create management cluster
func (c *TkgClient) InitRegion(options *InitRegionOptions) error { //nolint:funlen,gocyclo
	var err error
	var regionalConfigBytes []byte
	var isSuccessful = false
	var isStartedRegionalClusterCreation = false
	var isBootstrapClusterCreated = false
	var bootstrapClusterName string
	var regionContext region.RegionContext
	var filelock *fslock.Lock
	var configFilePath string

	bootstrapClusterKubeconfigPath, err := getTKGKubeConfigPath(false)
	if err != nil {
		return err
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

	c.ensureClusterTopologyConfiguration()

	providerName, _, err := ParseProviderName(options.InfrastructureProvider)
	if err != nil {
		return errors.Wrap(err, "unable to parse provider name")
	}

	// validate docker only if user is not using an existing cluster
	// Note: Validating in client code as well to cover the usecase where users use client code instead of command line.

	if err := c.ValidatePrerequisites(!options.UseExistingCluster, true); err != nil {
		return err
	}

	// validate docker resources if provider is docker
	if providerName == "docker" {
		if err := c.ValidateDockerResourcePrerequisites(); err != nil {
			return err
		}
	}

	log.Infof("Using infrastructure provider %s", options.InfrastructureProvider)
	log.SendProgressUpdate(statusRunning, StepGenerateClusterConfiguration, InitRegionSteps)
	log.Info("Generating cluster configuration...")

	log.SendProgressUpdate(statusRunning, StepSetupBootstrapCluster, InitRegionSteps)
	log.Info("Setting up bootstrapper...")
	// Ensure bootstrap cluster and copy boostrap cluster kubeconfig to ~/kube-tkg directory
	if bootstrapClusterName, err = c.ensureKindCluster(options.Kubeconfig, options.UseExistingCluster, bootstrapClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to create bootstrap cluster")
	}

	// Configure kubeconfig as part of options as bootstrap cluster kubeconfig
	options.Kubeconfig = bootstrapClusterKubeconfigPath

	isBootstrapClusterCreated = true
	log.Infof("Bootstrapper created. Kubeconfig: %s", bootstrapClusterKubeconfigPath)
	bootStrapClusterClient, err := clusterclient.NewClient(bootstrapClusterKubeconfigPath, "", clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get bootstrap cluster client")
	}
	bootstrapPkgClient, err := packageclient.NewPackageClientForContext(bootstrapClusterKubeconfigPath, "")
	if err != nil {
		return errors.Wrap(err, "unable to get a PackageClient")
	}

	// configure variables required to deploy providers
	if err := c.configureVariablesForProvidersInstallation(nil); err != nil {
		return errors.Wrap(err, "unable to configure variables for provider installation")
	}

	// If clusterclass feature flag is enabled then deploy kapp-controller
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Info("Installing kapp-controller on bootstrap cluster...")
		if err = c.InstallOrUpgradeKappController(bootStrapClusterClient, constants.OperationTypeInstall); err != nil {
			return errors.Wrap(err, "unable to install kapp-controller to bootstrap cluster")
		}
	}

	log.SendProgressUpdate(statusRunning, StepInstallProvidersOnBootstrapCluster, InitRegionSteps)
	log.Info("Installing providers on bootstrapper...")
	// Initialize bootstrap cluster with providers
	if err = c.InitializeProviders(options, bootStrapClusterClient, bootstrapClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to initialize providers")
	}

	// If clusterclass feature flag is enabled then deploy management components
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		if err = c.InstallOrUpgradeManagementComponents(bootStrapClusterClient, bootstrapPkgClient, "", false); err != nil {
			return errors.Wrap(err, "unable to install management components to bootstrap cluster")
		}

		akoRequired, err := c.isAKORequiredInBootstrapCluster()
		if err != nil {
			return errors.Wrap(err, "unable to check whether avi ha is enabled")
		}
		if akoRequired {
			log.Info("Installing AKO on bootstrapper...")
			if err = c.InstallAKO(bootStrapClusterClient); err != nil {
				return errors.Wrap(err, "unable to install ako")
			}
		}
	}

	if options.AdditionalTKGManifests != "" {
		log.Infof("Apply additional manifests %s for the bootstrap cluster in tkg-system", options.AdditionalTKGManifests)
		if err = bootStrapClusterClient.ApplyFileRecursively(options.AdditionalTKGManifests, "tkg-system"); err != nil {
			return errors.Wrap(err, "unable to apply additional manifests")
		}
	}

	// Obtain management cluster configuration of a provided flavor
	if regionalConfigBytes, options.ClusterName, configFilePath, err = c.BuildRegionalClusterConfiguration(options); err != nil {
		return errors.Wrap(err, "unable to build management cluster configuration")
	}
	log.Infof("Management cluster config file has been generated and stored at: '%v'", configFilePath)

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

	err = bootStrapClusterClient.WaitForControlPlaneAvailable(options.ClusterName, targetClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to wait for cluster control plane available")
	}
	log.Info("Management cluster control plane is available, means API server is ready to receive requests")

	kubeConfigBytes, err := bootStrapClusterClient.GetKubeConfigForCluster(options.ClusterName, targetClusterNamespace, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to extract kube config for cluster %s", options.ClusterName)
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
	regionalPkgClient, err := packageclient.NewPackageClientForContext(regionalClusterKubeconfigPath, kubeContext)
	if err != nil {
		return errors.Wrap(err, "unable to get a PackageClient")
	}

	// If clusterclass feature flag is enabled then deploy kapp-controller
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Info("Installing kapp-controller on management cluster...")
		if err = c.InstallOrUpgradeKappController(regionalClusterClient, constants.OperationTypeInstall); err != nil {
			return errors.Wrap(err, "unable to install kapp-controller to management cluster")
		}
	}

	err = bootStrapClusterClient.WaitForClusterInitialized(options.ClusterName, targetClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "error waiting for cluster to be provisioned (this may take a few minutes)")
	}

	log.SendProgressUpdate(statusRunning, StepInstallProvidersOnRegionalCluster, InitRegionSteps)
	log.Info("Installing providers on management cluster...")
	if err = c.InitializeProviders(options, regionalClusterClient, regionalClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to initialize providers on management cluster")
	}

	if err := regionalClusterClient.PatchClusterAPIAWSControllersToUseEC2Credentials(); err != nil {
		return err
	}

	// If clusterclass feature flag is enabled then deploy management components to the cluster
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		if err = c.InstallOrUpgradeManagementComponents(regionalClusterClient, regionalPkgClient, kubeContext, false); err != nil {
			return errors.Wrap(err, "unable to install management components to management cluster")
		}
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

	if options.AdditionalTKGManifests != "" {
		log.Infof("Apply additional manifests %s for the management cluster in %s", options.AdditionalTKGManifests, defaultTkgNamespace)
		if err = regionalClusterClient.ApplyFileRecursively(options.AdditionalTKGManifests, defaultTkgNamespace); err != nil {
			return errors.Wrap(err, "unable to apply additional manifests")
		}
	}

	// Applying ClusterBootstrap and its associated resources on the management cluster
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Infof("Applying ClusterBootstrap and its associated resources on management cluster")
		if err := c.ApplyClusterBootstrapObjects(bootStrapClusterClient, regionalClusterClient); err != nil {
			return errors.Wrap(err, "Unable to apply ClusterBootstarp and its associated resources on management cluster")
		}
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

	if !config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Info("Waiting for additional components to be up and running...")
		if err := c.WaitForAddonsDeployments(regionalClusterClient); err != nil {
			return err
		}
	}

	// Wait for packages if the feature-flag is disabled
	// We do not need to wait for packages as we have already installed and waited for all
	// packages to be deployed during tkg package installation
	if !config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Info("Waiting for packages to be up and running...")
		if err := c.WaitForPackages(regionalClusterClient, regionalClusterClient, options.ClusterName, targetClusterNamespace, true); err != nil {
			log.Warningf("Warning: Management cluster is created successfully, but some packages are failing. %v", err)
		}
	}

	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Info("Creating tkg-bom versioned ConfigMaps...")
		if err := c.CreateOrUpdateVerisionedTKGBom(regionalClusterClient); err != nil {
			log.Warningf("Warning: Management cluster is created successfully, but the tkg-bom versioned ConfigMaps creation is failing. %v", err)
		}
	}

	log.Infof("You can now access the management cluster %s by running 'kubectl config use-context %s'", options.ClusterName, kubeContext)
	isSuccessful = true
	return nil
}

// PatchClusterInitOperations Patches cluster
func (c *TkgClient) PatchClusterInitOperations(regionalClusterClient clusterclient.Client, options *InitRegionOptions, targetClusterNamespace string) error {
	// Patch management cluster with the TKG version
	err := regionalClusterClient.PatchClusterObjectWithTKGVersion(options.ClusterName, targetClusterNamespace, c.tkgBomClient.GetCurrentTKGVersion())
	if err != nil {
		return errors.Wrap(err, "unable to patch TKG Version")
	}

	// Patch management cluster with the edition used to bootstrap it (e.g. tce or community-edition)
	_, err = regionalClusterClient.PatchClusterObjectWithOptionalMetadata(options.ClusterName, targetClusterNamespace, "annotations", map[string]string{editionAnnotationKey: options.Edition})
	if err != nil {
		return errors.Wrap(err, "unable to patch edition under annotations")
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

	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		// Patch and remove kapp-controller labels from clusterclass resources
		err = c.removeKappControllerLabelsFromClusterClassResources(regionalClusterClient)
		if err != nil {
			return errors.Wrap(err, "unable to remove kapp-controller labels from the clusterclass resources")
		}
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

// ApplyClusterBootstrapObjects renders the ClusterBootstrap and its associated objects from templates and applies them to a target management cluster.
func (c *TkgClient) ApplyClusterBootstrapObjects(fromClusterClient, toClusterClient clusterclient.Client) error {
	path, err := c.tkgConfigPathsClient.GetTKGProvidersDirectory()
	if err != nil {
		return err
	}
	clusterBootstrapTemplatesDirPath := filepath.Join(path, "yttcb")
	configDefaultPath := filepath.Join(path, "config_default.yaml")

	userConfigValuesFile, err := c.getUserConfigVariableValueMapFile()
	if err != nil {
		return err
	}

	defer func() {
		// Remove intermediate config files if err is empty
		if err == nil {
			os.Remove(userConfigValuesFile)
		}
	}()

	log.V(6).Infof("User ConfigValues File: %v", userConfigValuesFile)

	clusterBootstrapBytes, err := carvelhelpers.ProcessYTTPackage(clusterBootstrapTemplatesDirPath, configDefaultPath, userConfigValuesFile)
	if err != nil {
		return err
	}

	clusterBootstrapString := string(clusterBootstrapBytes)
	if len(clusterBootstrapString) > 0 {

		// Before applying ClusterBootstrap need to check if TKR is available. Sometimes it takes some time for the
		// TKr to be available. Hence checking for TKr availability before applying ClusterBootstrap
		tkr, err := fromClusterClient.GetClusterResolvedTanzuKubernetesRelease()
		if tkr == nil || err != nil {
			return err // error occurs or management cluster does not have resolved TKR
		}

		var toClusterTKR v1alpha3.TanzuKubernetesRelease
		pollOptions := &clusterclient.PollOptions{Interval: clusterclient.CheckResourceInterval, Timeout: clusterclient.PackageInstallTimeout}
		log.V(6).Infof("Checking if TKr %s is created on management cluster", tkr.Name)
		if err := toClusterClient.GetResource(&toClusterTKR, tkr.Name, tkr.Namespace, nil, pollOptions); err != nil {
			return err
		}

		log.V(6).Infof("Applying ClusterBootstrap: %v", clusterBootstrapString)
		if err := toClusterClient.Apply(clusterBootstrapString); err != nil {
			return err
		}
	}

	return nil
}

// CopyNeededTKRAndOSImages moves the resolved TKR and associated OSImage CR for a Cluster resource
// from namespaceto a target management
func (c *TkgClient) CopyNeededTKRAndOSImages(fromClusterClient, toClusterClient clusterclient.Client) error {
	tkr, err := fromClusterClient.GetClusterResolvedTanzuKubernetesRelease()
	if tkr == nil {
		return err // error occurs or management cluster does not have resolved TKR
	}
	osImages, err := fromClusterClient.GetClusterResolvedOSImagesFromTKR(tkr)
	if err != nil {
		return err
	}

	// copy the solved TKR if not presented in the cleanup cluster
	var toClusterTKR v1alpha3.TanzuKubernetesRelease
	if err = toClusterClient.GetResource(&toClusterTKR, tkr.Name, tkr.Namespace, nil, nil); apierrors.IsNotFound(err) {
		tkr.SetResourceVersion("")
		if err := toClusterClient.CreateResource(tkr, tkr.Name, tkr.Namespace); err != nil {
			return err
		}
	}

	// copy the solved OSImage(s) if not presented in the cleanup cluster
	for _, osImage := range osImages {
		var toClusterOSImage v1alpha3.OSImage
		if err = toClusterClient.GetResource(&toClusterOSImage, osImage.Name, osImage.Namespace, nil, nil); apierrors.IsNotFound(err) {
			osImage.SetResourceVersion("")
			if err := toClusterClient.CreateResource(&osImage, osImage.Name, osImage.Namespace); err != nil {
				return err
			}
		}
	}
	return nil
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
		Kubeconfig:      options.Kubeconfig,
		TargetNamespace: options.Namespace,
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
	configImageTag(constants.ConfigVariableInternalCAPOCIManagerImageTag, "cluster-api-provider-oci", "capociControllerImage")
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
// returns cluster-configuration bytes, clustername, configFilePath, error if present
func (c *TkgClient) BuildRegionalClusterConfiguration(options *InitRegionOptions) ([]byte, string, string, error) {
	var bytes []byte
	var err error
	var configFilePath string

	if options.ClusterName == "" {
		options.ClusterName = generateRegionalClusterName(options.InfrastructureProvider, "")
	}

	namespace := options.Namespace
	if namespace == "" {
		namespace = defaultTkgNamespace
	}

	SetClusterClass(c.TKGConfigReaderWriter())

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
		YamlProcessor:            yamlprocessor.NewYttProcessorWithConfigDir(c.tkgConfigDir),
	}

	if options.IsInputFileClusterClassBased {
		bytes, err = getContentFromInputFile(options.ClusterConfigFile)
	} else {
		bytes, err = c.getClusterConfigurationBytes(&clusterConfigOptions, clusterConfigOptions.ProviderRepositorySource.InfrastructureProvider, true, false)
		if err != nil {
			return bytes, options.ClusterName, "", err
		}
		clusterConfigDir, err := c.tkgConfigPathsClient.GetClusterConfigurationDirectory()
		if err != nil {
			return bytes, options.ClusterName, "", err
		}
		configFilePath = filepath.Join(clusterConfigDir, fmt.Sprintf("%s.yaml", options.ClusterName))
		err = utils.SaveFile(configFilePath, bytes)
		if err != nil {
			return bytes, options.ClusterName, "", err
		}
	}

	return bytes, options.ClusterName, configFilePath, err
}

type waitForProvidersOptions struct {
	Kubeconfig      string
	TargetNamespace string
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
				return c.waitForProvider(clusterClient, providerNameVersion, provider.Type, options.TargetNamespace)
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

func (c *TkgClient) waitForProvider(clusterClient clusterclient.Client, name, providerType, targetNamespace string) error {
	providerOptions := clusterctl.ComponentsOptions{TargetNamespace: targetNamespace}
	// get the provider component from clusterctl
	providerComponents, err := c.clusterctlClient.GetProviderComponents(name, clusterctlv1.ProviderType(providerType), providerOptions)
	if err != nil {
		return errors.Wrap(err, "Error getting provider component")
	}

	var controllerName string
	var namespace string

	// get the deployment name and namespace from the provider
	for _, object := range providerComponents.Objs() {
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

// removeKappControllerLabelsFromClusterClassResources removes kapp-controller labels from clusterclass resources
func (c *TkgClient) removeKappControllerLabelsFromClusterClassResources(regionalClusterClient clusterclient.Client) error {
	errList := []error{}
	labelsToBeDeleted := []string{constants.KappControllerAppLabel, constants.KappControllerAssociationLabel}
	gvkToResourcesMap := map[schema.GroupVersionKind]string{
		capi.GroupVersion.WithKind(constants.KindClusterClass):                          constants.ResourceClusterClass,
		controlplanev1.GroupVersion.WithKind(constants.KindKubeadmControlPlaneTemplate): constants.ResourceKubeadmControlPlaneTemplate,
		bootstrapv1.GroupVersion.WithKind(constants.KindKubeadmConfigTemplate):          constants.ResourceKubeadmConfigTemplate,
		capav1beta2.GroupVersion.WithKind(constants.KindAWSClusterTemplate):             constants.ResourceAWSClusterTemplate,
		capav1beta2.GroupVersion.WithKind(constants.KindAWSMachineTemplate):             constants.ResourceAWSMachineTemplate,
		capzv1beta1.GroupVersion.WithKind(constants.KindAzureClusterTemplate):           constants.ResourceAzureClusterTemplate,
		capzv1beta1.GroupVersion.WithKind(constants.KindAzureMachineTemplate):           constants.ResourceAzureMachineTemplate,
		capvv1beta1.GroupVersion.WithKind(constants.KindVSphereClusterTemplate):         constants.ResourceVSphereClusterTemplate,
		capvv1beta1.GroupVersion.WithKind(constants.KindVSphereMachineTemplate):         constants.ResourceVSphereMachineTemplate,
	}

	for gvk, resourceName := range gvkToResourcesMap {
		if exists, err := regionalClusterClient.VerifyExistenceOfCRD(resourceName, gvk.Group); err != nil || !exists {
			continue
		}
		errList = append(errList, regionalClusterClient.RemoveMatchingMetadataFromResources(gvk, constants.TkgNamespace, "labels", labelsToBeDeleted))
	}

	return kerrors.NewAggregate(errList)
}

// isAKORequiredInBootstrapCluster return whether AVI_CONTROL_PLANE_HA_PROVIDER is enabled
func (c *TkgClient) isAKORequiredInBootstrapCluster() (bool, error) {
	log.V(5).Info("Get AVI_CONTROL_PLANE_HA_PROVIDER from user config ")

	if p, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereHaProvider); err == nil && p == trueString {
		return true, nil
	}
	return false, nil
}
