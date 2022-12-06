// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// GetClusterConfiguration gets cluster configuration
func (c *TkgClient) GetClusterConfiguration(options *CreateClusterOptions) ([]byte, error) { // nolint:gocyclo
	// check if user provided both infra provider name and version, so that user doesn't have to
	// have a management cluster created before he generates work load cluster config, else follow the usual path
	if options.ProviderRepositorySource.InfrastructureProvider != "" {
		provider, version, err := ParseProviderName(options.ProviderRepositorySource.InfrastructureProvider)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse the provider name")
		}

		// clusterctl doesn't connect to Management cluster if user provides BOTH infrastructure provider name and version
		// given that tkg stores the providers templates locally,
		// Note: If namespace is not provided in options, it defaults to constants.DefaultNamespace namespace
		if provider != "" && version != "" {
			if provider == PacificProviderName {
				if options.NodeSizeOptions.Size != "" || options.NodeSizeOptions.ControlPlaneSize != "" || options.NodeSizeOptions.WorkerSize != "" {
					return nil, errors.New("creating Tanzu Kubernetes Cluster is not compatible with the node size options: --size, --controlplane-size, and --worker-size")
				}
				return c.getPacificClusterConfiguration(options)
			}

			// Skip the validation when infrastructure name and version is passed
			// As '-i' is hidden option, while creating workload cluster template
			// this option is passed for quick verification and testing purpose only
			if err := c.configureAndValidateConfiguration(options, nil, true); err != nil {
				return nil, err
			}
			return c.getClusterConfiguration(&options.ClusterConfigOptions, false, provider, options.IsWindowsWorkloadCluster)
		}
	}
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get current management cluster context")
	}

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster client while getting cluster config")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return nil, errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		if options.NodeSizeOptions.Size != "" || options.NodeSizeOptions.ControlPlaneSize != "" || options.NodeSizeOptions.WorkerSize != "" {
			return nil, errors.New("creating Tanzu Kubernetes Cluster is not compatible with the node size options: --size, --controlplane-size, and --worker-size")
		}
		return c.getPacificClusterConfiguration(options)
	}

	options.Kubeconfig = clusterctlclient.Kubeconfig{Path: currentRegion.SourceFilePath, Context: currentRegion.ContextName}

	if err := c.configureAndValidateConfiguration(options, regionalClusterClient, options.SkipValidation); err != nil {
		return nil, err
	}

	infraProvider, err := regionalClusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster provider information.")
	}
	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return nil, err
	}

	return c.getClusterConfigurationBytes(&options.ClusterConfigOptions, infraProviderName, false, options.IsWindowsWorkloadCluster)
}

func (c *TkgClient) configureAndValidateConfiguration(options *CreateClusterOptions, regionalClusterClient clusterclient.Client, skipValidation bool) error {
	var err error
	if options.KubernetesVersion, options.TKRVersion, err = c.ConfigureAndValidateTkrVersion(options.TKRVersion); err != nil {
		return err
	}

	if err := c.ConfigureAndValidateWorkloadClusterConfiguration(options, regionalClusterClient, skipValidation); err != nil {
		return errors.Wrap(err, "workload cluster configuration validation failed")
	}
	return nil
}

func (c *TkgClient) getClusterConfiguration(options *ClusterConfigOptions, isManagementCluster bool, infraProvider string, isWindowsWorkloadCluster bool) ([]byte, error) {
	// Set CLUSTER_PLAN to viper configuration
	c.SetPlan(options.ProviderRepositorySource.Flavor)

	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedCC) {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableFeatureFlagPackageBasedCC, "true")
	}

	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return nil, err
	}

	// Sets cluster class value.
	SetClusterClass(c.TKGConfigReaderWriter())

	// need to provide clusterctl the worker count for md0 and not the full worker-machine-count value.
	workerCounts, err := c.DistributeMachineDeploymentWorkers(*options.WorkerMachineCount, options.ProviderRepositorySource.Flavor == constants.PlanProd, isManagementCluster, infraProviderName, isWindowsWorkloadCluster)
	if err != nil {
		return nil, errors.Wrap(err, "failed to distribute machine deployments")
	}
	c.SetMachineDeploymentWorkerCounts(workerCounts, *options.WorkerMachineCount, options.ProviderRepositorySource.Flavor == constants.PlanProd)
	md0WorkerCount := int64(workerCounts[0])
	options.WorkerMachineCount = &md0WorkerCount

	template, err := c.clusterctlClient.GetClusterTemplate(clusterctlclient.GetClusterTemplateOptions(*options))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get template")
	}
	return template.Yaml()
}

// SetPlan saves the plan name
func (c *TkgClient) SetPlan(planName string) {
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterPlan, planName)
}

// SetProviderType saves the provider type
func (c *TkgClient) SetProviderType(providerType string) {
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableProviderType, providerType)
}

// SetVsphereVersion saves the vsphere version
func (c *TkgClient) SetVsphereVersion(vsphereVersion string) {
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereVersion, vsphereVersion)
}

// SetBuildEdition saves the build edition
func (c *TkgClient) SetBuildEdition(buildEdition string) {
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableBuildEdition, buildEdition)
}

// SetTKGVersion saves the tkg version based on Default BoM file
func (c *TkgClient) SetTKGVersion() {
	bomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		log.Info("unable to get default BoM file...")
		return
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableTKGVersion, bomConfig.Release.Version)
}

// SetPinnipedConfigForWorkloadCluster sets the pinniped configuration(concierge) for workload cluster
func (c *TkgClient) SetPinnipedConfigForWorkloadCluster(issuerURL, issuerCA string) {
	c.TKGConfigReaderWriter().Set(constants.ConfigVariablePinnipedSupervisorIssuerURL, issuerURL)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariablePinnipedSupervisorIssuerCABundleData, issuerCA)
}
