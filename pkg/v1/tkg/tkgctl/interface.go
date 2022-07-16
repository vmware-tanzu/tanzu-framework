// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

// TKGClient implements TKG client
type TKGClient interface {
	// AddRegion adds region
	AddRegion(options AddRegionOptions) error
	// ConfigCluster prints cluster template to stdout
	ConfigCluster(configClusterOption CreateClusterOptions) error
	// CreateAWSCloudFormationStack create aws cloud formation stack
	CreateAWSCloudFormationStack(clusterConfigFile string) error
	// CreateCluster create tkg cluster
	CreateCluster(cc CreateClusterOptions) error
	// DeleteCluster deletes workload cluster
	DeleteCluster(options DeleteClustersOptions) error
	// DeleteMachineHealthCheck deletes MHC on cluster
	DeleteMachineHealthCheck(options DeleteMachineHealthCheckOptions) error
	// DeleteRegion deletes management cluster
	DeleteRegion(options DeleteRegionOptions) error
	// GetCEIP returns CEIP status set on management cluster
	GetCEIP() (client.ClusterCeipInfo, error)
	// GetClusters returns list of cluster
	GetClusters(options ListTKGClustersOptions) ([]client.ClusterInfo, error)
	// DescribeCluster describes all the objects in the Cluster
	DescribeCluster(options DescribeTKGClustersOptions) (DescribeClusterResult, error)
	// DescribeProviders describes all the installed providers
	DescribeProviders() (*clusterctlv1.ProviderList, error)
	// GenerateAWSCloudFormationTemplate generates a YAML template for AWS CloudFormation
	GenerateAWSCloudFormationTemplate(clusterConfigFile string) (string, error)
	// GetCredentials saves cluster credentials to a file
	GetCredentials(options GetWorkloadClusterCredentialsOptions) error
	// GetKubernetesVersions returns supported k8s versions
	GetKubernetesVersions() (*client.KubernetesVersionsInfo, error)
	// GetMachineHealthCheck return machinehealthcheck configuration for the cluster
	GetMachineHealthCheck(options GetMachineHealthCheckOptions) ([]client.MachineHealthCheck, error)
	// GetRegions return list of management clusters
	GetRegions(managementClusterName string) ([]region.RegionContext, error)
	// Init initializes tkg management cluster
	Init(options InitRegionOptions) error
	// ScaleCluster scales cluster
	ScaleCluster(options ScaleClusterOptions) error
	// SetCeip sets CEIP to the management cluster
	SetCeip(ceipOptIn, isProd, labels string) error
	// SetMachineHealthCheck apply machine health check to the cluster
	SetMachineHealthCheck(options SetMachineHealthCheckOptions) error
	// GetMachineDeployments gets machine deployments from a cluster
	GetMachineDeployments(options client.GetMachineDeploymentOptions) ([]capi.MachineDeployment, error)
	// SetMachineDeployment applies a machine deployment to the cluster
	SetMachineDeployment(options *client.SetMachineDeploymentOptions) error
	// DeleteMachineDeployment deletes a machine deployment from the cluster
	DeleteMachineDeployment(options client.DeleteMachineDeploymentOptions) error
	// SetRegion sets active management cluster
	SetRegion(options SetRegionOptions) error
	// UpgradeCluster upgrade tkg workload cluster
	UpgradeCluster(options UpgradeClusterOptions) error
	// UpgradeRegion upgrades management cluster
	UpgradeRegion(options UpgradeRegionOptions) error
	// Updates management cluster
	UpdateCredentialsRegion(options UpdateCredentialsRegionOptions) error
	// Updates workload cluster
	UpdateCredentialsCluster(options UpdateCredentialsClusterOptions) error
	// GetClusterPinnipedInfo returns the cluster and pinniped info
	GetClusterPinnipedInfo(options GetClusterPinnipedInfoOptions) (*client.ClusterPinnipedInfo, error)
	// GetTanzuKubernetesReleases returns the available TanzuKubernetesReleases
	// Deprecated: This would not be supported from TKR API version v1alpha3,
	// user can use go client to get TKR
	GetTanzuKubernetesReleases(tkrName string) ([]runv1alpha1.TanzuKubernetesRelease, error)
	// ActivateTanzuKubernetesReleases activates TanzuKubernetesRelease
	// Deprecated: This would not be supported from TKR API version v1alpha3,
	// user can use go client to set the labels to activate/deactivate the TKR
	ActivateTanzuKubernetesReleases(tkrName string) error
	// DeactivateTanzuKubernetesReleases deactivates TanzuKubernetesRelease
	// Deprecated: This would not be supported from TKR API version v1alpha3,
	// user can use go client to set the labels to activate/deactivate the TKR
	DeactivateTanzuKubernetesReleases(tkrName string) error
	// IsPacificRegionalCluster checks if the cluster pointed to by kubeconfig is Pacific management cluster(supervisor)
	IsPacificRegionalCluster() (bool, error)
	// GetPacificClusterObject gets Pacific cluster object
	GetPacificClusterObject(clusterName, namespace string) (*tkgsv1alpha2.TanzuKubernetesCluster, error)
	// GetPacificMachineDeployments gets machine deployments from a Pacific cluster
	// Note: This would be soon deprecated after TKGS and TKGm adopt the clusterclass
	GetPacificMachineDeployments(options client.GetMachineDeploymentOptions) ([]capiv1alpha3.MachineDeployment, error)
	// FeatureGateHelper returns feature gate helper to query feature gate
	FeatureGateHelper() FeatureGateHelper
}
