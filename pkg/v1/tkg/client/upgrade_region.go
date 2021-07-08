// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"golang.org/x/sync/errgroup"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providersupgradeclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
)

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

	// ManagementGroup that should be upgraded (e.g. capi-system/cluster-api).
	ManagementGroup string

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
	providers       []clusterctlv1.Provider
	managementGroup string
}

// UpgradeManagementCluster upgrades management clusters providers and k8s version
// Steps:
// 1. Upgrade providers
// 	a) Get the Upgrade configuration by reading BOM file to get the providers versions
// 	b) Get the providers information from the management cluster
//  c) Prepare the providers upgrade information
// 	d) Call the clusterctl ApplyUpgrade() to upgrade providers
//  e) Wait for providers to be up and running
// 2. call the UpgradeCluster() for upgrading the k8s version of the Management cluster
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

	if err := c.configureVariablesForProvidersInstallation(regionalClusterClient); err != nil {
		return errors.Wrap(err, "unable to configure variables for provider installation")
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
		Kubeconfig:        options.Kubeconfig,
		TargetNamespace:   "",
		WatchingNamespace: "",
	}
	err = c.WaitForProviders(regionalClusterClient, waitOptions)
	if err != nil {
		return errors.Wrap(err, "error waiting for provider components to be up and running after upgrading them")
	}
	log.Info("Management cluster providers upgraded successfully...")

	log.Info("Upgrading management cluster kubernetes version...")
	err = c.UpgradeCluster(options)
	if err != nil {
		return errors.Wrap(err, "unable to upgrade management cluster")
	}
	// Patch management cluster with the TKG version
	err = regionalClusterClient.PatchClusterObjectWithTKGVersion(options.ClusterName, options.Namespace, c.tkgBomClient.GetCurrentTKGVersion())
	if err != nil {
		return err
	}

	// Upgrade/Add certain addons to the old clusters during upgrade
	// This is done after we patch the management cluster object with new TKG version
	// so, while generating cluster template with new tkg and k8s version, it does not
	// throw version incompatibility validation error.
	if !options.SkipAddonUpgrade {
		err = c.upgradeAddons(regionalClusterClient, regionalClusterClient, options.ClusterName, options.Namespace, true)
		if err != nil {
			return err
		}
	}

	log.Info("Waiting for additional components to be up and running...")
	if err := c.WaitForAddonsDeployments(regionalClusterClient); err != nil {
		return err
	}
	return nil
}

func (c *TkgClient) configureVariablesForProvidersInstallation(regionalClusterClient clusterclient.Client) error {
	infraProvider, err := regionalClusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster provider information.")
	}
	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return errors.Wrap(err, "failed to parse provider name")
	}
	// set default values for variables required for infrastructure component spec rendering
	err = c.setConfigurationForUpgrade(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "failed to set configurations for upgrade")
	}

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
func (c *TkgClient) GenerateProvidersUpgradeOptions(pUpgradeInfo *providersUpgradeInfo) (*ApplyProvidersUpgradeOptions, error) {
	puo := &ApplyProvidersUpgradeOptions{}

	puo.ManagementGroup = pUpgradeInfo.managementGroup
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

	// get the management group
	pUpgradeInfo.managementGroup, err = parseManagementGroup(installedProviders)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse the management group")
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

func parseManagementGroup(installedProviders *clusterctlv1.ProviderList) (string, error) {
	for i := range installedProviders.Items {
		if clusterctlv1.ProviderType(installedProviders.Items[i].Type) == clusterctlv1.CoreProviderType {
			mgmtGroupName := installedProviders.Items[i].InstanceName()
			return mgmtGroupName, nil
		}
	}
	return "", errors.New("failed to find core provider from the current providers")
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
