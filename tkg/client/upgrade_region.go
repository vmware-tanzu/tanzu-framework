// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	betav1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/providersupgradeclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

// ErrorBlockUpgradeTMCIncompatible defines the error message to display during upgrade when cluster is registred to TMC and TMC does not support latest version of TKG
const ErrorBlockUpgradeTMCIncompatible = "The management cluster cannot be upgraded to Tanzu Kubernetes Grid '%v' while registered to Tanzu Mission Control. See https://via.vmw.com/tkg-14-tmc-compat for more information."

// UpgradeManagementClusterOptions upgrade management cluster options
type UpgradeManagementClusterOptions struct {
	ClusterName         string
	Namespace           string
	KubernetesVersion   string
	Kubeconfig          string
	IsRegionalCluster   bool
	VSphereTemplateName string
	BOMFilePath         string
}

// ApplyProvidersUpgradeOptions carries the options supported by upgrade apply.
type ApplyProvidersUpgradeOptions struct {
	// Kubeconfig file to use for accessing the management cluster. If empty, default discovery rules apply.
	Kubeconfig clusterctl.Kubeconfig

	// Contract defines the API Version of Cluster API (contract e.g. v1alpha3) the management group should upgrade to.
	// When upgrading by contract, the latest versions available will be used for all the providers; if you want
	// a more granular control on upgrade, use CoreProvider, BootstrapProviders, ControlPlaneProviders, InfrastructureProviders.
	// Note: For tkg we will ignore this option as tkg management cluster is opinionated and it controls version of the providers to be upgraded
	Contract string

	// CoreProvider instance and version (e.g. capi-system/cluster-api:v0.3.0) to upgrade to. This field can be used as alternative to Contract.
	CoreProvider string

	// BootstrapProviders instance and versions (e.g. capi-kubeadm-bootstrap-system/kubeadm:v0.3.0) to upgrade to. This field can be used as alternative to Contract.
	BootstrapProviders []string

	// ControlPlaneProviders instance and versions (e.g. capi-kubeadm-control-plane-system/kubeadm:v0.3.0) to upgrade to. This field can be used as alternative to Contract.
	ControlPlaneProviders []string

	// InfrastructureProviders instance and versions (e.g. capa-system/aws:v0.5.0) to upgrade to. This field can be used as alternative to Contract.
	InfrastructureProviders []string
}

type providersUpgradeInfo struct {
	providers []clusterctlv1.Provider
}

// UpgradeManagementCluster upgrades management clusters providers and k8s version
// Steps:
//  1. Upgrade providers
//     a) Get the Upgrade configuration by reading BOM file to get the providers versions
//     b) Get the providers information from the management cluster
//     c) Prepare the providers upgrade information
//     d) Call the clusterctl ApplyUpgrade() to upgrade providers
//     e) Wait for providers to be up and running
//  2. call the UpgradeCluster() for upgrading the k8s version of the Management cluster
//
//nolint:gocyclo
func (c *TkgClient) UpgradeManagementCluster(options *UpgradeClusterOptions) error {
	contexts, err := c.GetRegionContexts(options.ClusterName)
	if err != nil || len(contexts) == 0 {
		return errors.Errorf("management cluster %s not found", options.ClusterName)
	}
	currentRegion := contexts[0]
	options.Kubeconfig = currentRegion.SourceFilePath

	if currentRegion.Status == region.Failed {
		return errors.Errorf("cannot upgrade since deployment failed for management cluster %s", currentRegion.ClusterName)
	}

	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while upgrading management cluster")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return errors.New("upgrading 'Tanzu Kubernetes Cluster service for vSphere' management cluster is not yet supported")
	}

	// Validate the upgrade and preconfigure required variables
	err = c.validateAndconfigure(options, regionalClusterClient)
	if err != nil {
		return err
	}

	log.Info("Upgrading management cluster providers...")
	providersUpgradeClient := providersupgradeclient.New(c.clusterctlClient)
	if err = c.DoProvidersUpgrade(regionalClusterClient, currentRegion.ContextName, providersUpgradeClient, options); err != nil {
		return errors.Wrap(err, "failed to upgrade management cluster providers")
	}

	// Wait for installed providers to get up and running
	// TODO: Currently tkg doesn't support TargetNamespace and WatchingNamespace as it's not supporting multi-tenency of providers
	// If we support it in future we need to make these namespaces as command line options and use here
	waitOptions := waitForProvidersOptions{
		Kubeconfig:      options.Kubeconfig,
		TargetNamespace: "",
	}
	err = c.WaitForProviders(regionalClusterClient, waitOptions)
	if err != nil {
		return errors.Wrap(err, "error waiting for provider components to be up and running after upgrading them")
	}
	log.Info("Management cluster providers upgraded successfully...")

	// If clusterclass feature flag is enabled then deploy management components
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		log.Info("Upgrading kapp-controller...")
		if err = c.InstallOrUpgradeKappController(currentRegion.SourceFilePath, currentRegion.ContextName, constants.OperationTypeUpgrade); err != nil {
			return errors.Wrap(err, "unable to upgrade kapp-controller")
		}
		log.Info("Upgrading management components...")
		if err = c.InstallOrUpgradeManagementComponents(currentRegion.SourceFilePath, currentRegion.ContextName, true); err != nil {
			return errors.Wrap(err, "unable to upgrade management components")
		}
	}

	log.Info("Upgrading management cluster kubernetes version...")
	err = c.UpgradeCluster(options)
	if err != nil {
		return errors.Wrap(err, "unable to upgrade management cluster")
	}

	err = c.upgradeTelemetryImageIfExists(regionalClusterClient, currentRegion)
	if err != nil {
		return err
	}

	// Patch management cluster with the TKG version
	err = regionalClusterClient.PatchClusterObjectWithTKGVersion(options.ClusterName, options.Namespace, c.tkgBomClient.GetCurrentTKGVersion())
	if err != nil {
		return err
	}

	return nil
}

func (c *TkgClient) validateAndconfigure(options *UpgradeClusterOptions, regionalClusterClient clusterclient.Client) error {
	// Validate the compatibility before upgrading management cluster
	log.Infof("Validating the compatibility before management cluster upgrade")
	err := c.validateCompatibilityBeforeManagementClusterUpgrade(options, regionalClusterClient)
	if err != nil {
		return err
	}

	// Validate required environment variables are set
	log.Infof("Validating for the required environment variables to be set")
	if err := c.validateEnvVariables(regionalClusterClient); err != nil {
		return errors.Wrap(err, "required env variables are not set")
	}

	if err := c.configureVariablesForProvidersInstallation(regionalClusterClient); err != nil {
		return errors.Wrap(err, "unable to configure variables for provider installation")
	}

	return nil
}

func (c *TkgClient) configureVariablesForProvidersInstallation(regionalClusterClient clusterclient.Client) error {
	err := c.configureImageTagsForProviderInstallation()
	if err != nil {
		return errors.Wrap(err, "failed to configure image tags for provider installation")
	}

	// If region client is not specified nothing to configure based on existing management cluster
	if regionalClusterClient == nil {
		return nil
	}

	infraProvider, err := regionalClusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster provider information.")
	}
	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return errors.Wrap(err, "failed to parse provider name")
	}
	// retrieve required variables required for infrastructure component spec rendering
	// set them to default values if they don't exist.
	err = c.RetrieveRegionalClusterConfiguration(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "failed to set configurations for upgrade")
	}

	// Configure provider name to readerwriter config
	c.SetProviderType(infraProviderName)

	switch infraProviderName {
	case AzureProviderName:
		// since the templates needs Base64 values of credentials, encode them
		if _, err := c.EncodeAzureCredentialsAndGetClient(regionalClusterClient); err != nil {
			return errors.Wrap(err, "failed to encode azure credentials")
		}
	case AWSProviderName:
		if _, err := c.EncodeAWSCredentialsAndGetClient(regionalClusterClient); err != nil {
			return errors.Wrap(err, "failed to encode AWS credentials")
		}
	case VSphereProviderName:
		if err := c.configureVsphereCredentialsFromCluster(regionalClusterClient); err != nil {
			return errors.Wrap(err, "failed to configure Vsphere credentials")
		}
	case DockerProviderName:
		// no variable configuration is needed to deploy Docker provider as
		// infrastructure-components.yaml for docker does not require any variable
	}
	return nil
}

// DoProvidersUpgrade upgrades the providers of the management cluster
func (c *TkgClient) DoProvidersUpgrade(regionalClusterClient clusterclient.Client, ctx string,
	providersUpgradeClient providersupgradeclient.Client, options *UpgradeClusterOptions) error {
	// read the BOM file for latest providers version information to upgrade to
	bomConfiguration, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return errors.Wrap(err, "unable to read in configuration from BOM file")
	}

	pUpgradeInfo, err := c.getProvidersUpgradeInfo(regionalClusterClient, bomConfiguration)
	if err != nil {
		return errors.Wrap(err, "failed to get providers upgrade information")
	}
	if len(pUpgradeInfo.providers) == 0 {
		log.Infof("All providers are up to date...")
		return nil
	}

	pUpgradeApplyOptions, err := c.GenerateProvidersUpgradeOptions(pUpgradeInfo)
	if err != nil {
		return errors.Wrap(err, "failed to generate providers upgrade apply options")
	}

	// update the kubeconfig
	pUpgradeApplyOptions.Kubeconfig.Path = options.Kubeconfig
	pUpgradeApplyOptions.Kubeconfig.Context = ctx

	log.V(6).Infof("clusterctl upgrade apply options: %+v", *pUpgradeApplyOptions)
	clusterctlUpgradeOptions := clusterctl.ApplyUpgradeOptions(*pUpgradeApplyOptions)
	err = providersUpgradeClient.ApplyUpgrade(&clusterctlUpgradeOptions)
	if err != nil {
		return errors.Wrap(err, "failed to apply providers upgrade")
	}

	return nil
}

// GenerateProvidersUpgradeOptions generates provider upgrade options
func (c *TkgClient) GenerateProvidersUpgradeOptions(pUpgradeInfo *providersUpgradeInfo) (*clusterctl.ApplyUpgradeOptions, error) {
	puo := &clusterctl.ApplyUpgradeOptions{}
	puo.Contract = "v1beta1"

	for i := range pUpgradeInfo.providers {
		instanceVersion := pUpgradeInfo.providers[i].Namespace + "/" + pUpgradeInfo.providers[i].ProviderName + ":" + pUpgradeInfo.providers[i].Version
		switch clusterctlv1.ProviderType(pUpgradeInfo.providers[i].Type) {
		case clusterctlv1.CoreProviderType:
			puo.CoreProvider = instanceVersion
		case clusterctlv1.BootstrapProviderType:
			puo.BootstrapProviders = append(puo.BootstrapProviders, instanceVersion)
		case clusterctlv1.ControlPlaneProviderType:
			puo.ControlPlaneProviders = append(puo.ControlPlaneProviders, instanceVersion)
		case clusterctlv1.InfrastructureProviderType:
			puo.InfrastructureProviders = append(puo.InfrastructureProviders, instanceVersion)
		default:
			return nil, errors.Errorf("unknown provider type: %s", pUpgradeInfo.providers[i].Type)
		}
	}
	return puo, nil
}

// getProvidersUpgradeInfo prepares the upgrade information by comparing the provider current version with and the upgradable version
// obtained from the BOM file.
func (c *TkgClient) getProvidersUpgradeInfo(regionalClusterClient clusterclient.Client, bomConfig *tkgconfigbom.BOMConfiguration) (*providersUpgradeInfo, error) {
	pUpgradeInfo := &providersUpgradeInfo{}

	// Get all the installed providers info
	installedProviders := &clusterctlv1.ProviderList{}
	err := regionalClusterClient.ListResources(installedProviders, &crtclient.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get installed provider config")
	}

	// get the providers Info with the version updated with the upgrade version obtained from BOM file map
	upgradeProviderVersionMap := bomConfig.ProvidersVersionMap
	// make a list of providers eligible for upgrade
	for i := range installedProviders.Items {
		// Note: provider.Name has the manifest label (eg:control-plane-kubeadm) and provider.ProviderName would not be ideal(eg:kubeadm)
		// here as both bootstrap-kubeadm and control-plane-kubeadm has the same ProviderName as 'kubeadm'
		latestVersion, ok := upgradeProviderVersionMap[installedProviders.Items[i].Name]
		if !ok || latestVersion == "" {
			log.Warningf(" %s provider's version is missing in BOM file, so it would not be upgraded ", installedProviders.Items[i].Name)
			continue
		}
		latestSemVersion, err := version.ParseSemantic(latestVersion)
		if err != nil {
			log.Warningf("failed to parse %s provider's upgrade version, so it would not be upgraded ", installedProviders.Items[i].Name)
			continue
		}
		currentSemVersion, err := version.ParseSemantic(installedProviders.Items[i].Version)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s provider's current version", installedProviders.Items[i].Name)
		}
		if latestSemVersion.LessThan(currentSemVersion) {
			log.V(1).Infof("%s provider's upgrade version %s is less than current version %s, so skipping it for upgrade ",
				installedProviders.Items[i].ProviderName, latestVersion, installedProviders.Items[i].Version)
			continue
		}
		// update the provider to the latest version to be upgraded
		installedProviders.Items[i].Version = fmt.Sprintf("v%v.%v.%v", latestSemVersion.Major(), latestSemVersion.Minor(), latestSemVersion.Patch())
		pUpgradeInfo.providers = append(pUpgradeInfo.providers, installedProviders.Items[i])
	}

	return pUpgradeInfo, nil
}

// WaitForAddonsDeployments wait for addons deployments
func (c *TkgClient) WaitForAddonsDeployments(clusterClient clusterclient.Client) error {
	group, _ := errgroup.WithContext(context.Background())

	group.Go(
		func() error {
			err := clusterClient.WaitForDeployment(constants.TkrControllerDeploymentName, constants.TkrNamespace)
			if err != nil {
				log.V(3).Warningf("Failed waiting for deployment %s", constants.TkrControllerDeploymentName)
			}
			return err
		})

	group.Go(
		func() error {
			err := clusterClient.WaitForDeployment(constants.KappControllerDeploymentName, constants.KappControllerNamespace)
			if err != nil {
				log.V(3).Warningf("Failed waiting for deployment %s", constants.KappControllerDeploymentName)
			}
			return err
		})

	group.Go(
		func() error {
			err := clusterClient.WaitForDeployment(constants.AddonsManagerDeploymentName, constants.KappControllerNamespace)
			if err != nil {
				log.V(3).Warningf("Failed waiting for deployment %s", constants.AddonsManagerDeploymentName)
			}
			return err
		})

	err := group.Wait()
	if err != nil {
		return errors.Wrap(err, "Failed waiting for at least one CRS deployment, check logs for more detail.")
	}
	return nil
}

// WaitForPackages wait for packages to be up and running
func (c *TkgClient) WaitForPackages(regionalClusterClient, currentClusterClient clusterclient.Client, clusterName, namespace string, isRegionalCluster bool) error {
	// Adding kapp-controller package to the exclude list
	// For management cluster, kapp-controller is deployed using CRS and addon secret does not exist
	// For workload cluster, kapp-controller is deployed by addons manager. Even though the
	// addon secret for kapp-controller exists, it is not deployed using PackageInstall.
	// Hence skipping it while waiting for packages.
	ListExcludePackageInstallsFromWait := []string{constants.KappControllerPackageName}

	// Get the list of addons secrets
	secretList := &corev1.SecretList{}
	err := regionalClusterClient.ListResources(secretList, &crtclient.ListOptions{Namespace: namespace})
	if err != nil {
		return errors.Wrap(err, "unable to get list of secrets")
	}

	// From the addons secret get the names of package installs for each addon secret
	// This is determined from the "tkg.tanzu.vmware.com/addon-name" label on the secret
	var packagesInstalled []kapppkgv1alpha1.Package

	// Add tanzu-core-management-plugins packages to the list of packages to wait for management-cluster
	if isRegionalCluster {
		packagesInstalled = append(packagesInstalled, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: constants.CoreManagementPluginsPackageName, Namespace: constants.TkgNamespace}})
	}

	for i := range secretList.Items {
		if secretList.Items[i].Type == constants.AddonSecretType {
			if cn, exists := secretList.Items[i].Labels[constants.ClusterNameLabel]; exists && cn == clusterName {
				if addonName, exists := secretList.Items[i].Labels[constants.AddonNameLabel]; exists {
					if !utils.ContainsString(ListExcludePackageInstallsFromWait, addonName) {
						packagesInstalled = append(packagesInstalled, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: addonName, Namespace: constants.TkgNamespace}})
					}
				}
			}
		}
	}
	return WaitForPackagesInstallation(currentClusterClient, packagesInstalled, c.getPackageInstallTimeoutFromConfig())
}

func (c *TkgClient) getPackageInstallTimeoutFromConfig() time.Duration {
	var err error
	packageInstallTimeoutStr, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariablePackageInstallTimeout)
	packageInstallTimeout := time.Duration(0)
	if packageInstallTimeoutStr != "" {
		packageInstallTimeout, err = time.ParseDuration(packageInstallTimeoutStr)
		if err != nil {
			log.Infof("Unable to parse '%s: %v'", constants.ConfigVariablePackageInstallTimeout, packageInstallTimeoutStr)
		}
	}
	return packageInstallTimeout
}

func (c *TkgClient) validateCompatibilityBeforeManagementClusterUpgrade(options *UpgradeClusterOptions, regionalClusterClient clusterclient.Client) error {
	err := c.ValidateManagementClusterUpgradeVersionCompatibility(options, regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "upgrade version compatibility validation failed")
	}
	return c.validateCompatibilityWithTMC(regionalClusterClient, options.SkipPrompt)
}

// ValidateManagementClusterUpgradeVersionCompatibility validates the upgrade version compatibility for a management cluster
func (c *TkgClient) ValidateManagementClusterUpgradeVersionCompatibility(options *UpgradeClusterOptions, regionalClusterClient clusterclient.Client) error {
	mgmtClusterSemVersion, currentTKGSemVersion, err := c.getManagementClusterAndCurrentTKGSemVersions(options, regionalClusterClient)
	if err != nil {
		return err
	}
	log.V(9).Infof("Management cluster SemVersion is %q and current TKG SemVersion(to be upgraded to) is %q",
		mgmtClusterSemVersion.String(), currentTKGSemVersion.String())

	// TODO: Update this condition when TKG major version is changed
	if mgmtClusterSemVersion.Major() != currentTKGSemVersion.Major() {
		return errors.Errorf("major version mismatch detected")
	}
	warningMsg := ""
	minorGap := int(currentTKGSemVersion.Minor()) - int(mgmtClusterSemVersion.Minor())
	if minorGap < 0 || (minorGap == 0 && currentTKGSemVersion.Patch() < mgmtClusterSemVersion.Patch()) {
		return errors.Errorf("TKG version downgrade is not supported")
	} else if minorGap > 1 {
		warningMsg = "Upgrade skipping minor version is detected and is not recommended. It could leave cluster unmanageable"
	}
	if warningMsg != "" {
		if !options.SkipPrompt {
			log.Warning(warningMsg)
			if err := component.AskForConfirmation("Do you want to continue with the upgrade?"); err != nil {
				return err
			}
		} else {
			log.Infof("Warning: %v", warningMsg)
		}
	}
	return nil
}

func (c *TkgClient) getManagementClusterAndCurrentTKGSemVersions(options *UpgradeClusterOptions,
	regionalClusterClient clusterclient.Client) (mcSemversion, currentTKGSemVersion *version.Version, err error) {

	namespace := options.Namespace
	if namespace == "" {
		namespace = TKGsystemNamespace
	}
	mgmtClusterTkgVersion, err := regionalClusterClient.GetManagementClusterTKGVersion(options.ClusterName, namespace)
	if err != nil {
		return mcSemversion, currentTKGSemVersion, errors.Wrapf(err, "unable to get tkg version of management cluster %q in namespace %q", options.ClusterName, options.Namespace)
	}
	mcSemversion, err = version.ParseSemantic(mgmtClusterTkgVersion)
	if err != nil {
		return mcSemversion, currentTKGSemVersion, errors.Wrapf(err, "unable to parse management cluster version %s", mgmtClusterTkgVersion)
	}
	currentTKGVersion := c.tkgBomClient.GetCurrentTKGVersion()
	currentTKGSemVersion, err = version.ParseSemantic(currentTKGVersion)
	if err != nil {
		return mcSemversion, currentTKGSemVersion, errors.Wrapf(err, "unable to parse management cluster version %s", currentTKGVersion)
	}
	return
}

// validateCompatibilityWithTMC validate compatibility of new TKG version with TMC if management cluster is registered with TMC
func (c *TkgClient) validateCompatibilityWithTMC(regionalClusterClient clusterclient.Client, skipPrompt bool) error {
	registered, err := regionalClusterClient.IsClusterRegisteredToTMC()
	if err != nil || !registered {
		return nil
	}

	log.Info("Management Cluster is registered with Tanzu Mission Control. Validating upgrade compatibility...")

	tkgVersion, err := c.tkgBomClient.GetDefaultTKGReleaseVersion()
	if err != nil {
		return err
	}

	tmcInteropConfigMap := &corev1.ConfigMap{}
	if err := regionalClusterClient.GetResource(tmcInteropConfigMap, "interop", constants.TmcNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			// Compatibility has to be explicitly communicated by an existing iterop ConfigMap, hence failing.
			// This is because old management cluster will not have this ConfigMap created
			// and TMC will add this functionality in future.
			return errors.Errorf(ErrorBlockUpgradeTMCIncompatible, tkgVersion)
		}

		if !skipPrompt {
			log.Infof("error occurred while validating compatibility with Tanzu Mission Control, %v", err)
			log.Warning("Unable to validate compatibility of new version with Tanzu Mission Control. See https://via.vmw.com/tkg-14-tmc-compat for more information.")
			if err := component.AskForConfirmation("Do you want to continue with the upgrade?"); err != nil {
				return err
			}
		} else {
			log.Infof("Warning: Unable to validate compatibility of new version with Tanzu Mission Control: %v", err)
		}

		return nil
	}

	// Get the supported versions by TMC. Below is the 'interop' ConfigMap sample.
	// Also verify the `interop-schema-version` is of version 'v1'
	//
	// apiVersion: v1
	// kind: ConfigMap
	// metadata:
	//   name: interop
	//   namespace: vmware-system-tmc
	// data:
	//   interop-schema-version: "v1.0"
	//   supported-versions: "v1.4.0;v1.4.1"
	interopSchemaVersion, found := tmcInteropConfigMap.Data["interop-schema-version"]
	if !found || interopSchemaVersion != "v1.0" {
		return errors.Errorf(ErrorBlockUpgradeTMCIncompatible, tkgVersion)
	}
	supportedVersions, found := tmcInteropConfigMap.Data["supported-versions"]
	if !found {
		return errors.Errorf(ErrorBlockUpgradeTMCIncompatible, tkgVersion)
	}

	lstSupportedVersions := strings.Split(supportedVersions, ";")
	for _, sv := range lstSupportedVersions {
		if utils.CompareMajorMinorPatchVersion(sv, tkgVersion) {
			// If compatible version is found return nil which will allow
			// user to continue the upgrade as TMC supports new version of TKG
			return nil
		}
	}

	return errors.Errorf(ErrorBlockUpgradeTMCIncompatible, tkgVersion)
}

// upgradeTelemetryImage will upgrade the telemetry image if the management-cluster is opted into telemetry
// additional labels set by the user using the CLI will be preserved (such as: isProd or customer identification)
func (c *TkgClient) upgradeTelemetryImageIfExists(regionalClusterClient clusterclient.Client, ctx region.RegionContext) error {
	CEIPJobExists, err := regionalClusterClient.HasCEIPTelemetryJob(ctx.ClusterName)
	if err != nil {
		return errors.Wrap(err, "unable to determine if telemetry cron job is installed already")
	}

	if CEIPJobExists {
		telemetryJob := &betav1.CronJob{}
		err := regionalClusterClient.GetResource(telemetryJob, "tkg-telemetry", "tkg-system-telemetry", nil, nil)
		if err != nil {
			return errors.Wrap(err, "Failed to determine if telemetry job is installed or not")
		}

		log.Info("Upgrading management cluster telemetry image...")
		// telemetry command: /tkg-telemetry <sonobuoy bin path> <datastore-url> <labels>
		command := telemetryJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command
		dsURL := command[2]
		isProd := !strings.Contains(dsURL, "stg")
		labels := strings.Split(command[3], "--labels=")[1]
		err = c.DoSetCEIPParticipation(regionalClusterClient, ctx, true, strconv.FormatBool(isProd), labels)
		if err != nil {
			return errors.Wrap(err, "Failed to upgrade telemetry image")
		}
	} else {
		log.Info("management cluster is opted out of telemetry - skipping telemetry image upgrade")
	}

	return nil
}
