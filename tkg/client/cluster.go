// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	// PacificProviderName tkgs provider name
	PacificProviderName = "tkg-service-vsphere"
	// VSphereProviderName vsphere provider name
	VSphereProviderName = "vsphere"
	// AWSProviderName aws provider name
	AWSProviderName = "aws"
	// AzureProviderName azure provider name
	AzureProviderName = "azure"
	// DockerProviderName docker provider name
	DockerProviderName = "docker"

	defaultPacificProviderVersion = "v1.1.0"
)

const (
	// TkgLabelClusterRolePrefix cluster-role label prefix
	TkgLabelClusterRolePrefix = "cluster-role.tkg.tanzu.vmware.com/"
	// TkgLabelClusterRoleManagement cluster-role management label
	TkgLabelClusterRoleManagement = "management"
	// TkgLabelClusterRoleWorkload cluster-role workload label
	TkgLabelClusterRoleWorkload = "workload"
)

type waitForAddonsOptions struct {
	regionalClusterClient clusterclient.Client
	workloadClusterClient clusterclient.Client
	clusterName           string
	namespace             string
	waitForCNI            bool
	isTKGSCluster         bool
}

// TKGSupportedClusterOptions is the comma separated list of cluster options that could be enabled by user
// !!!NOTE this is set during the build time.
// Only the cluster options mentioned in "AllowedEnableOptions" could be enabled by user through command line option("enable-cluster-options"),
// if TKGSupportedClusterOptions is set to ""(for development purpose) the check would be deactivated
var TKGSupportedClusterOptions string

// CreateCluster creates a workload cluster based on a cluster template
// generated from the provided options. Returns whether cluster creation was attempted
// information along with error information
func (c *TkgClient) CreateCluster(options *CreateClusterOptions, waitForCluster bool) (bool, error) { //nolint:gocyclo,funlen
	if err := CheckClusterNameFormat(options.ClusterName, options.ProviderRepositorySource.InfrastructureProvider); err != nil {
		return false, NewValidationError(ValidationErrorCode, err.Error())
	}
	log.Info("Validating configuration...")
	// validate kubectl only since we need only kubectl for create cluster
	if err := c.ValidatePrerequisites(false, true); err != nil {
		return false, err
	}
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return false, errors.Wrap(err, "cannot get current management cluster context")
	}

	options.Kubeconfig = clusterctlclient.Kubeconfig{Path: currentRegion.SourceFilePath, Context: currentRegion.ContextName}
	if options.ProviderRepositorySource.InfrastructureProvider != "" {
		options.ProviderRepositorySource.InfrastructureProvider, err = c.tkgConfigUpdaterClient.CheckInfrastructureVersion(options.ProviderRepositorySource.InfrastructureProvider)
		if err != nil {
			return false, errors.Wrap(err, "unable to check infrastructure provider version")
		}
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
		OperationTimeout:  c.timeout,
	}
	regionalClusterClient, err := clusterclient.NewClient(options.Kubeconfig.Path, options.Kubeconfig.Context, clusterclientOptions)
	if err != nil {
		return false, errors.Wrap(err, "unable to get cluster client while creating cluster")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return false, errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		return true, c.createPacificCluster(options, waitForCluster)
	}

	// Validate management cluster version
	err = c.ValidateManagementClusterVersionWithCLI(regionalClusterClient)
	if err != nil {
		return false, errors.Wrap(err, "validation failed")
	}

	var bytes []byte
	var configFilePath string
	isManagementCluster := false
	if options.KubernetesVersion, options.TKRVersion, err = c.ConfigureAndValidateTkrVersion(options.TKRVersion); err != nil {
		return false, err
	}
	if customImageRepo, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository); err != nil && customImageRepo != "" && tkgconfighelper.IsCustomRepository(customImageRepo) {
		log.Infof("Using custom image repository: %s", customImageRepo)
	}

	if err := c.ConfigureAndValidateWorkloadClusterConfiguration(options, regionalClusterClient, false); err != nil {
		return false, errors.Wrap(err, "workload cluster configuration validation failed")
	}
	infraProvider, err := regionalClusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return false, errors.Wrap(err, "failed to get cluster provider information.")
	}
	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return false, err
	}

	if options.IsInputFileClusterClassBased {
		bytes, err = getContentFromInputFile(options.ClusterConfigFile)
		if err != nil {
			return false, errors.Wrap(err, "unable to get cluster configuration")
		}
	} else {
		bytes, err = c.getClusterConfigurationBytes(&options.ClusterConfigOptions, infraProviderName, isManagementCluster, options.IsWindowsWorkloadCluster)
		if err != nil {
			return false, errors.Wrap(err, "unable to get cluster configuration")
		}

		if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
			clusterConfigDir, err := c.tkgConfigPathsClient.GetClusterConfigurationDirectory()
			if err != nil {
				return false, err
			}
			configFilePath = filepath.Join(clusterConfigDir, fmt.Sprintf("%s.yaml", options.ClusterName))
			err = utils.SaveFile(configFilePath, bytes)
			if err != nil {
				return false, err
			}

			log.Warningf("\nLegacy configuration file detected. The inputs from said file have been converted into the new Cluster configuration as '%v'", configFilePath)

			// If `features.cluster.auto-apply-generated-clusterclass-based-configuration` feature-flag is not activated
			// log command to use to create cluster using ClusterClass based config file and return
			if !config.IsFeatureActivated(constants.FeatureFlagAutoApplyGeneratedClusterClassBasedConfiguration) {
				log.Warningf("\nTo create a cluster with it, use")
				log.Warningf("    tanzu cluster create --file %v", configFilePath)
				return false, nil
			}
			log.Warningf("\nUsing this new Cluster configuration '%v' to create the cluster.\n", configFilePath)
		}
	}

	log.Infof("creating workload cluster '%s'...", options.ClusterName)
	err = c.DoCreateCluster(regionalClusterClient, options.ClusterName, options.TargetNamespace, string(bytes))
	if err != nil {
		return false, errors.Wrap(err, "unable to create cluster")
	}
	// If user opts not to wait for the cluster to be provisioned, return
	if !waitForCluster {
		return false, nil
	}
	return true, c.waitForClusterCreation(regionalClusterClient, options)
}

// getClusterConfigurationBytes returns cluster configuration by taking into consideration of legacy vs clusterclass based cluster creation
func (c *TkgClient) getClusterConfigurationBytes(options *ClusterConfigOptions, infraProviderName string, isManagementCluster, isWindowsWorkloadCluster bool) ([]byte, error) {
	deployClusterClassBasedCluster, err := c.ShouldDeployClusterClassBasedCluster(isManagementCluster)
	if err != nil {
		return nil, err
	}

	// If ClusterClass based cluster creation is feasible update the plan to use ClusterClass based plan
	if deployClusterClassBasedCluster {
		plan, err := getCCPlanFromLegacyPlan(options.ProviderRepositorySource.Flavor)
		if err != nil {
			return nil, err
		}
		options.ProviderRepositorySource.Flavor = plan
	}

	// Get the cluster configuration yaml bytes
	return c.getClusterConfiguration(options, isManagementCluster, infraProviderName, isWindowsWorkloadCluster)
}

func getContentFromInputFile(fileName string) ([]byte, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error while reading input file %v : ", fileName))
	}
	return content, nil
}

func (c *TkgClient) waitForClusterCreation(regionalClusterClient clusterclient.Client, options *CreateClusterOptions) error {
	log.Info("waiting for cluster to be initialized...")
	kubeConfigBytes, err := c.WaitForClusterInitializedAndGetKubeConfig(regionalClusterClient, options.ClusterName, options.TargetNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to wait for cluster and get the cluster kubeconfig")
	}

	workloadClusterKubeconfigPath, err := getTKGKubeConfigPath(false)
	if err != nil {
		return errors.Wrap(err, "unable to save workload cluster kubeconfig to temporary path")
	}
	defer func() {
		_ = utils.DeleteFile(workloadClusterKubeconfigPath)
	}()

	kubeContext, err := MergeKubeConfigAndSwitchContext(kubeConfigBytes, workloadClusterKubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to save management cluster kubeconfig to TKG managed kubeconfig")
	}

	log.Info("waiting for cluster nodes to be available...")
	if err := c.WaitForClusterReadyAfterCreate(regionalClusterClient, options.ClusterName, options.TargetNamespace); err != nil {
		return errors.Wrap(err, "unable to wait for cluster nodes to be available")
	}

	isClusterClassBased, err := regionalClusterClient.IsClusterClassBased(options.ClusterName, options.TargetNamespace)
	if err != nil {
		return errors.Wrap(err, "error while checking workload cluster type")
	}

	c.WaitForAutoscalerDeployment(regionalClusterClient, options.ClusterName, options.TargetNamespace, isClusterClassBased)
	workloadClusterClient, err := clusterclient.NewClient(workloadClusterKubeconfigPath, kubeContext, clusterclient.Options{OperationTimeout: 15 * time.Minute})
	if err != nil {
		return errors.Wrap(err, "unable to create workload cluster client")
	}

	isTKGSCluster, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return err
	}
	if isClusterClassBased {
		log.Info("waiting for addons core packages installation...")
		if err := c.WaitForAddonsCorePackagesInstallation(waitForAddonsOptions{
			regionalClusterClient: regionalClusterClient,
			workloadClusterClient: workloadClusterClient,
			clusterName:           options.ClusterName,
			namespace:             options.TargetNamespace,
			waitForCNI:            true,
			isTKGSCluster:         isTKGSCluster,
		}); err != nil {
			return errors.Wrap(err, "error waiting for addons to get installed")
		}
	} else {
		log.Info("waiting for addons installation...")
		if err := c.WaitForAddons(waitForAddonsOptions{
			regionalClusterClient: regionalClusterClient,
			workloadClusterClient: workloadClusterClient,
			clusterName:           options.ClusterName,
			namespace:             options.TargetNamespace,
			waitForCNI:            true,
		}); err != nil {
			return errors.Wrap(err, "error waiting for addons to get installed")
		}
		log.Info("waiting for packages to be up and running...")
		if err := c.WaitForPackages(regionalClusterClient, workloadClusterClient, options.ClusterName, options.TargetNamespace, false); err != nil {
			log.Warningf("warning: Cluster is created successfully, but some packages are failing. %v", err)
		}
	}
	return nil
}

func (c *TkgClient) getValueForAutoscalerDeploymentConfig() bool {
	var autoscalerEnabled string
	var isEnabled bool
	var err error

	// swallowing the error when the value for config variable 'ENABLE_AUTOSCALER' is not set
	if autoscalerEnabled, err = c.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableAutoscaler); err != nil {
		return false
	}

	if isEnabled, err = strconv.ParseBool(autoscalerEnabled); err != nil {
		log.Warningf("unable to parse the value of config variable %q. reason: %v", constants.ConfigVariableEnableAutoscaler, err)
		return false
	}

	return isEnabled
}

// WaitForAutoscalerDeployment waits for autoscaler deployment if enabled
func (c *TkgClient) WaitForAutoscalerDeployment(regionalClusterClient clusterclient.Client, clusterName, targetNamespace string, isClusterClassBased bool) {
	isEnabled := false
	autoscalerDeploymentName := clusterName + constants.AutoscalerDeploymentNameSuffix
	if isClusterClassBased {
		autoscalerDeployment, err := regionalClusterClient.GetDeployment(autoscalerDeploymentName, targetNamespace)
		if autoscalerDeployment.Name == "" || err != nil {
			log.Warning("unable to get the autoscaler deployment, maybe it is not exist")
			return
		}
		isEnabled = true
	} else {
		isEnabled = c.getValueForAutoscalerDeploymentConfig()
	}
	if isEnabled {
		log.Warning("Waiting for cluster autoscaler to be available...")
		if err := regionalClusterClient.WaitForAutoscalerDeployment(autoscalerDeploymentName, targetNamespace); err != nil {
			log.Warningf("Unable to wait for autoscaler deployment to be ready. reason: %v", err)
		}
	}
}

// DoCreateCluster performs steps to create cluster
func (c *TkgClient) DoCreateCluster(clusterClient clusterclient.Client, name, namespace, manifest string) error {
	var err error

	if name == "" {
		return errors.New("invalid cluster name")
	}
	if manifest == "" {
		return errors.New("invalid cluster manifest")
	}

	err = clusterClient.Apply(manifest)
	if err != nil {
		return errors.Wrap(err, "unable to apply cluster configuration")
	}

	err = clusterClient.PatchClusterWithOperationStartedStatus(name, namespace, clusterclient.OperationTypeCreate, c.timeout)
	if err != nil {
		log.V(6).Infof("unable to patch cluster object with operation status, %s", err.Error())
	}

	return nil
}

// WaitForClusterInitializedAndGetKubeConfig wait for cluster initialization and once initialized get kubeconfig
func (c *TkgClient) WaitForClusterInitializedAndGetKubeConfig(clusterClient clusterclient.Client, name, targetNamespace string) ([]byte, error) {
	var err error
	var kubeConfigBytes []byte

	err = clusterClient.WaitForClusterInitialized(name, targetNamespace)
	if err != nil {
		return kubeConfigBytes, errors.Wrap(err, "error waiting for cluster to be provisioned (this may take a few minutes)")
	}

	kubeConfigBytes, err = clusterClient.GetKubeConfigForCluster(name, targetNamespace, nil)
	if err != nil {
		return kubeConfigBytes, errors.Wrapf(err, "unable to extract kube config for cluster %s", name)
	}

	// Wait for the cluster to actually answer. This ensure that the LB/DNS are effectively in place

	return kubeConfigBytes, nil
}

// WaitForClusterReadyForMove wait for cluster to be ready for move operation
func (c *TkgClient) WaitForClusterReadyForMove(clusterClient clusterclient.Client, name, targetNamespace string) error {
	return clusterClient.WaitForClusterReady(name, targetNamespace, true)
}

// WaitForClusterReadyAfterCreate wait for cluster to be ready after creation
func (c *TkgClient) WaitForClusterReadyAfterCreate(clusterClient clusterclient.Client, name, targetNamespace string) error {
	// For now we use the same waiting logic to wait for workload cluster creation. As an enhancement we may
	// want to be less stringent or more adaptive to the parameters used to create the cluster (wait longer when
	// worker size is big, or wait for fewer nodes, etc)
	return clusterClient.WaitForClusterReady(name, targetNamespace, true)
}

// WaitForClusterReadyAfterReverseMove Called when relocating cluster-api objects out from the management cluster to the cleanup cluster.
func (c *TkgClient) WaitForClusterReadyAfterReverseMove(clusterClient clusterclient.Client, name, targetNamespace string) error {
	// The cluster checker should not get hung up on cluster not having its full inventory of worker and
	// control plane replicas
	checkReplicas := false

	// Nb. After move we are required to wait for the ClusterAPi objects to get reconciled before doing delete.
	// this is required because delete will fail if the object does not have the status field properly set.
	// This is achieved by testing the cluster & machine status reports ready (as it was before move).
	return clusterClient.WaitForClusterReady(name, targetNamespace, checkReplicas)
}

// WaitForAddons wait for addons to be installed
func (c *TkgClient) WaitForAddons(options waitForAddonsOptions) error {
	if err := c.waitForCRS(options); err != nil {
		return err
	}
	if options.waitForCNI {
		if err := c.waitForCNI(options); err != nil {
			return err
		}
	}
	return nil
}

func (c *TkgClient) waitForCNI(options waitForAddonsOptions) error {
	cni, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)
	if err != nil {
		log.Info("Warning: unable to get CNI, skipping CNI installation verification")
	}

	if cni == "antrea" {
		if err := options.workloadClusterClient.WaitForDeployment(
			constants.AntreaDeploymentName,
			constants.AntreaDeploymentNamespace); err != nil {
			return errors.Wrap(err, "timeout waiting for antrea cni to start")
		}
	} else if cni == "calico" {
		if err := options.workloadClusterClient.WaitForDeployment(
			constants.CalicoDeploymentName,
			constants.CalicoDeploymentNamespace); err != nil {
			return errors.Wrap(err, "timeout waiting for calico cni to start")
		}
	}
	return nil
}

func (c *TkgClient) waitForCRS(options waitForAddonsOptions) error {
	crsList := &addonsv1.ClusterResourceSetList{}
	err := options.regionalClusterClient.GetResourceList(crsList,
		options.clusterName,
		options.namespace,
		clusterclient.VerifyCRSAppliedSuccessfully,
		&clusterclient.PollOptions{Interval: clusterclient.CheckClusterInterval, Timeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "error waiting for ClusterResourceSet object to be applied for the cluster")
	}

	return nil
}

// WaitForAddonsCorePackagesInstallation gets ClusterBootstrap and collects list of addons core packages, and monitors the kapp controller package installation in management cluster and rest of core packages installation in workload cluster
func (c *TkgClient) WaitForAddonsCorePackagesInstallation(options waitForAddonsOptions) error {
	clusterBootstrap, err := GetClusterBootstrap(options.regionalClusterClient, options.clusterName, options.namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error while getting ClusterBootstrap object for workload cluster: %v", options.clusterName))
	}
	var corePackagesNamespace string
	if options.isTKGSCluster {
		corePackagesNamespace = constants.CorePackagesNamespaceInTKGS
	} else {
		corePackagesNamespace = constants.CorePackagesNamespaceInTKGM
	}
	packages, err := GetCorePackagesFromClusterBootstrap(options.regionalClusterClient, options.workloadClusterClient, clusterBootstrap, corePackagesNamespace, options.clusterName)
	if err != nil {
		return err
	}
	return MonitorAddonsCorePackageInstallation(options.regionalClusterClient, options.workloadClusterClient, packages, c.getPackageInstallTimeoutFromConfig())
}

func (c *TkgClient) createPacificCluster(options *CreateClusterOptions, waitForCluster bool) (err error) {
	// get the configuration
	var configYaml []byte
	var clusterName, namespace string

	if options.NodeSizeOptions.Size != "" || options.NodeSizeOptions.ControlPlaneSize != "" || options.NodeSizeOptions.WorkerSize != "" {
		return errors.New("creating Tanzu Kubernetes Cluster is not compatible with the node size options: --size, --controlplane-size, and --worker-size")
	}

	if options.IsInputFileClusterClassBased {
		configYaml, err = getContentFromInputFile(options.ClusterConfigFile)
		if err != nil {
			return errors.Wrap(err, "unable to get cluster configuration")
		}
	} else {
		configYaml, err = c.getPacificClusterConfiguration(options)
		if err != nil {
			return errors.Wrap(err, "failed to create Tanzu Kubernetes Cluster service for vSphere workload cluster")
		}
	}

	clusterName, namespace, err = c.getClusterNameAndNameSpace()
	if err != nil {
		return errors.Wrap(err, "failed to get cluster name and namespace")
	}

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
		OperationTimeout:  c.timeout,
	}
	clusterClient, err := clusterclient.NewClient(options.Kubeconfig.Path, options.Kubeconfig.Context, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}
	// Apply the template(with variables subsitituted) to the supervisor cluster
	if err := clusterClient.Apply(string(configYaml)); err != nil {
		return errors.Wrap(err, "failed to apply the cluster configuration")
	}
	// If user opts not to wait for the cluster to be provisioned, return
	if !waitForCluster {
		return nil
	}

	log.V(3).Infof("Waiting for the Tanzu Kubernetes Cluster service for vSphere workload cluster\n")
	if options.IsInputFileClusterClassBased {
		err = c.waitForClusterCreation(clusterClient, options)
	} else {
		err = clusterClient.WaitForPacificCluster(clusterName, namespace)
	}
	if err != nil {
		return errors.Wrap(err, "failed waiting for workload cluster")
	}
	return nil
}

func (c *TkgClient) getPacificClusterConfiguration(options *CreateClusterOptions) ([]byte, error) {
	if options.ProviderRepositorySource.InfrastructureProvider == "" {
		options.ProviderRepositorySource.InfrastructureProvider = PacificProviderName
	}

	// parse the abbreviated syntax for name[:version]
	name, providerVersion, err := ParseProviderName(options.ProviderRepositorySource.InfrastructureProvider)
	if err != nil {
		return nil, err
	}

	if providerVersion == "" {
		// TODO: should be changed once we get APIs from Pacific to determine the version, for now using "1.1.0"
		providerVersion = defaultPacificProviderVersion
	}

	// Set CLUSTER_PLAN to viper configuration
	c.SetPlan(options.ProviderRepositorySource.Flavor)
	c.SetProviderType(name)
	c.SetTKGClusterRole(WorkloadCluster)
	c.SetTKGVersion()
	tkrName := utils.GetTkrNameFromTkrVersion(options.TKRVersion)
	if !strings.HasPrefix(tkrName, "v") {
		tkrName = "v" + tkrName
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableTkrName, tkrName)
	err = c.ConfigureAndValidateCNIType(options.CniType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to validate CNI")
	}

	clusterCtlConfigClient := c.readerwriterConfigClient.ClusterConfigClient()
	providerConfig, err := clusterCtlConfigClient.Providers().Get(name, clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return nil, err
	}

	kubeconfig := cluster.Kubeconfig{Path: options.Kubeconfig.Path, Context: options.Kubeconfig.Context}
	clusterClient := cluster.New(kubeconfig, clusterCtlConfigClient)

	// If the option specifying the targetNamespace is empty, try to detect it.
	if options.TargetNamespace == "" {
		currentNamespace, err := clusterClient.Proxy().CurrentNamespace()
		if err != nil {
			return nil, err
		}
		options.TargetNamespace = currentNamespace
	}

	// Inject some of the templateOptions into the configClient so they can be consumed as a variables from the template.
	if err := c.templateOptionsToVariables(&options.ClusterConfigOptions); err != nil {
		return nil, err
	}
	repo, err := repository.New(providerConfig, clusterCtlConfigClient, repository.InjectYamlProcessor(options.YamlProcessor))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create repository Client for getting Tanzu Kubernetes Cluster service for vSphere cluster plan")
	}

	template, err := repo.Templates(providerVersion).Get(options.ProviderRepositorySource.Flavor, options.TargetNamespace, false)
	if err != nil {
		return nil, err
	}
	yaml, err := template.Yaml()
	if err != nil {
		return nil, errors.Wrap(err, "error generating yaml file to create cluster")
	}
	return yaml, nil
}

// templateOptionsToVariables injects some of the templateOptions to the configClient so they can be consumed as a variables from the template.
func (c *TkgClient) templateOptionsToVariables(options *ClusterConfigOptions) error {
	if options.TargetNamespace != "" {
		// the TargetNamespace, if valid, can be used in templates using the ${ NAMESPACE } variable.
		if err := validateDNS1123Label(options.TargetNamespace); err != nil {
			return errors.Wrapf(err, "invalid target-namespace")
		}
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableNamespace, options.TargetNamespace)
	}

	if options.ClusterName != "" {
		// the ClusterName, if valid, can be used in templates using the ${ CLUSTER_NAME } variable.
		if err := validateDNS1123Label(options.ClusterName); err != nil {
			return errors.Wrapf(err, "invalid cluster name")
		}
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterName, options.ClusterName)
	}

	// the KubernetesVersion, if valid, can be used in templates using the ${ KUBERNETES_VERSION } variable.
	// NB. in case the KubernetesVersion from the templateOptions is empty, we are not setting any values so the
	// configClient is going to search into os env variables/the clusterctl config file as a fallback options.
	if options.KubernetesVersion != "" {
		if _, err := version.ParseSemantic(options.KubernetesVersion); err != nil {
			return errors.Errorf("invalid KubernetesVersion. Please use a semantic version number")
		}
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableKubernetesVersion, options.KubernetesVersion)
	}

	// the ControlPlaneMachineCount, if valid, can be used in templates using the ${ CONTROL_PLANE_MACHINE_COUNT } variable.
	if *options.ControlPlaneMachineCount < 1 {
		return errors.Errorf("invalid ControlPlaneMachineCount. Please use a number greater or equal than 1")
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableControlPlaneMachineCount, strconv.Itoa(int(*options.ControlPlaneMachineCount)))

	// the WorkerMachineCount, if valid, can be used in templates using the ${ WORKER_MACHINE_COUNT } variable.
	if *options.WorkerMachineCount < 0 {
		return errors.Errorf("invalid WorkerMachineCount. Please use a number greater or equal than 0")
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount, strconv.Itoa(int(*options.WorkerMachineCount)))
	return nil
}

func (c *TkgClient) getClusterNameAndNameSpace() (string, string, error) {
	var err error
	clusterName := ""
	if clusterName, err = c.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName); err != nil {
		return "", "", errors.Wrap(err, "failed to get the cluster name")
	}
	namespace := ""
	if namespace, err = c.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace); err != nil {
		return "", "", errors.Wrap(err, "failed to get the namespace")
	}
	return clusterName, namespace, err
}

// ConfigureAndValidateWorkloadClusterConfiguration configures and validates workload cluster configuration
func (c *TkgClient) ConfigureAndValidateWorkloadClusterConfiguration(options *CreateClusterOptions, clusterClient clusterclient.Client, skipValidation bool) error { // nolint:gocyclo
	var err error

	err = c.ValidateAndConfigureClusterOptions(options)
	if err != nil {
		return errors.Wrap(err, "unable to configure the cluster options")
	}
	infraProvider := options.ProviderRepositorySource.InfrastructureProvider
	if infraProvider == "" {
		if infraProvider, err = clusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType); err != nil {
			return err
		}
	}

	providerName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return err
	}

	if err = c.ConfigureAndValidateHTTPProxyConfiguration(providerName); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if options.ClusterType == "" {
		options.ClusterType = WorkloadCluster
	}
	// BUILD_EDITION is the Tanzu Edition, the plugin should be built for. Its value is supposed be constructed from
	// cmd/cli/plugin/managementcluster/create.go. So empty value at this point is not expected.
	if options.Edition == "" {
		return NewValidationError(ValidationErrorCode, "required config variable 'edition' is not set")
	}
	c.SetBuildEdition(options.Edition)
	c.SetTKGClusterRole(options.ClusterType)
	c.SetTKGVersion()
	if !skipValidation {
		// Get the Pinniped information required for workload cluster from management cluster
		// NOTE: Not blocking the workload cluster deployment if the pinniped information is not available on management cluster
		pinnipedIssuerURL, pinnipedIssuerCAData, err := clusterClient.GetPinnipedIssuerURLAndCA()
		if err != nil {
			log.Warningf("Warning: Pinniped configuration not found; Authentication via Pinniped will not be set up in this cluster. If you wish to set up Pinniped after the cluster is created, please refer to the documentation.")
		} else {
			c.SetPinnipedConfigForWorkloadCluster(pinnipedIssuerURL, pinnipedIssuerCAData)
		}
	}

	err = c.ConfigureAndValidateCNIType(options.CniType)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "unable to validate CNI type").Error())
	}

	if err = c.configureAndValidateIPFamilyConfiguration(TkgLabelClusterRoleWorkload); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.configureAndValidateCoreDNSIP(); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.validateServiceCIDRNetmask(); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.ConfigureAndValidateNameserverConfiguration(TkgLabelClusterRoleWorkload); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err := c.configureAndValidateProviderConfig(providerName, options, clusterClient, skipValidation); err != nil {
		return err
	}

	return c.ValidateSupportOfK8sVersionForManagmentCluster(clusterClient, options.KubernetesVersion, skipValidation)
}

// ValidateProviderConfig configure and validate based on provider
func (c *TkgClient) configureAndValidateProviderConfig(providerName string, options *CreateClusterOptions, clusterClient clusterclient.Client, skipValidation bool) error {
	isProdPlan := IsProdPlan(options.ClusterConfigOptions.ProviderRepositorySource.Flavor)
	switch providerName {
	case AWSProviderName:
		if err := c.ConfigureAndValidateAWSConfig(options.TKRVersion, options.NodeSizeOptions, skipValidation,
			isProdPlan, *options.WorkerMachineCount, clusterClient, false); err != nil {
			return errors.Wrap(err, "AWS config validation failed")
		}
	case VSphereProviderName:
		if err := c.ConfigureAndValidateVsphereConfig(options.TKRVersion, options.NodeSizeOptions, options.VsphereControlPlaneEndpoint, skipValidation, nil); err != nil {
			return errors.Wrap(err, "vSphere config validation failed")
		}
		if err := c.ValidateVsphereVipWorkloadCluster(clusterClient, options.VsphereControlPlaneEndpoint, skipValidation); err != nil {
			return NewValidationError(ValidationErrorCode, errors.Wrap(err, "vSphere control plane endpoint IP validation failed").Error())
		}
	case AzureProviderName:
		if err := c.ConfigureAndValidateAzureConfig(options.TKRVersion, options.NodeSizeOptions, skipValidation, nil); err != nil {
			return errors.Wrap(err, "Azure config validation failed")
		}
	case DockerProviderName:
		if err := c.ConfigureAndValidateDockerConfig(options.TKRVersion, options.NodeSizeOptions, skipValidation); err != nil {
			return NewValidationError(ValidationErrorCode, err.Error())
		}
	}

	workerCounts, err := c.DistributeMachineDeploymentWorkers(*options.WorkerMachineCount, isProdPlan, false, providerName, false)
	if err != nil {
		return errors.Wrap(err, "failed to distribute machine deployments")
	}
	c.SetMachineDeploymentWorkerCounts(workerCounts, *options.WorkerMachineCount, isProdPlan)

	return nil
}

// ValidateSupportOfK8sVersionForManagmentCluster validate k8s version support for management cluster
func (c *TkgClient) ValidateSupportOfK8sVersionForManagmentCluster(regionalClusterClient clusterclient.Client, kubernetesVersion string, skipValidation bool) error {
	if skipValidation {
		return nil
	}

	mgmtClusterName, mgmtClusterNamespace, err := c.getRegionalClusterNameAndNamespace(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "unable to get name and namespace of current management cluster")
	}

	mgmtClusterTkgVersion, err := regionalClusterClient.GetManagementClusterTKGVersion(mgmtClusterName, mgmtClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to get tkg version of management clusters")
	}

	return tkgconfighelper.ValidateK8sVersionSupport(mgmtClusterTkgVersion, kubernetesVersion)
}

// IsPacificManagementCluster check if management cluster is TKGS cluster
func (c *TkgClient) IsPacificManagementCluster() (bool, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return false, errors.Wrap(err, "cannot get current management cluster context")
	}

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
		OperationTimeout:  c.timeout,
	}

	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return false, errors.Wrap(err, "unable to get management cluster client")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return false, errors.Wrap(err, "error while determining infrastructure provider type")
	}
	return isPacific, nil
}

// ValidateAndConfigureClusterOptions validates and configures the cluster options user want to enable
// through command line options
func (c *TkgClient) ValidateAndConfigureClusterOptions(options *CreateClusterOptions) error {
	if options.ClusterOptionsEnableList == nil || len(options.ClusterOptionsEnableList) == 0 {
		return nil
	}
	clusterOptions, err := getClusterOptionsEnableList(options.ClusterOptionsEnableList)
	if err != nil {
		return errors.Wrap(err, "cluster options validation failed")
	}

	supportedOptionsMap := make(map[string]bool)
	supportedOptions := []string{}

	if TKGSupportedClusterOptions != "" {
		supportedOptions = strings.Split(TKGSupportedClusterOptions, ",")
		log.V(9).Infof("List of cluster options supported are :%v", supportedOptions)
		for _, option := range supportedOptions {
			supportedOptionsMap[option] = true
		}
	}
	// set the config value using viper, so that ytt can access these values
	for _, option := range clusterOptions {
		// for development, this would be empty which implies there is no restriction imposed
		if len(supportedOptions) != 0 {
			if _, exist := supportedOptionsMap[option]; !exist {
				return errors.Errorf("cluster option '%s' is not supported in this release. Supported cluster options are %v", option, supportedOptions)
			}
		}
		// convert to upper case and replace hypens with underscore and prefix with "ENABLE_"
		opt := "ENABLE_" + strings.ToUpper(strings.ReplaceAll(option, "-", "_"))
		c.TKGConfigReaderWriter().Set(opt, strconv.FormatBool(true))
	}
	return nil
}

// ValidateManagementClusterVersionWithCLI validate management cluster version with cli version
func (c *TkgClient) ValidateManagementClusterVersionWithCLI(regionalClusterClient clusterclient.Client) error {
	mgmtClusterName, mgmtClusterNamespace, err := c.getRegionalClusterNameAndNamespace(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "unable to get name and namespace of current management cluster")
	}

	mgmtClusterTkgVersion, err := regionalClusterClient.GetManagementClusterTKGVersion(mgmtClusterName, mgmtClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to get tkg version of management clusters")
	}

	defaultTKGVersion, err := c.tkgBomClient.GetDefaultTKGReleaseVersion()
	if err != nil {
		return errors.Wrap(err, "unable to get default TKG version")
	}

	curMCSemVersion, err := version.ParseSemantic(mgmtClusterTkgVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to parse management cluster version %s", mgmtClusterTkgVersion)
	}

	defaultTKGSemVersion, err := version.ParseSemantic(defaultTKGVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to parse TKG version %s", defaultTKGVersion)
	}

	if curMCSemVersion.Major() != defaultTKGSemVersion.Major() ||
		curMCSemVersion.Minor() != defaultTKGSemVersion.Minor() {
		return errors.Errorf("version mismatch between management cluster and cli version. Please upgrade your management cluster to the latest to continue")
	}

	return nil
}
