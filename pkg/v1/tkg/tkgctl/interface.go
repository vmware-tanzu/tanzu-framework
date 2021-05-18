// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	runv1alpha1 "github.com/vmware-tanzu-private/core/pkg/v1/tkg/api/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/region"
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
	// DeregisterFromTmc deregister management cluster from TMC
	DeregisterFromTmc(options DeregisterFromTMCOptions) error
	// GetCEIP returns CEIP status set on management cluster
	GetCEIP() (client.ClusterCeipInfo, error)
	// GetClusters returns list of cluster
	GetClusters(options ListTKGClustersOptions) ([]client.ClusterInfo, error)
	// DescribeCluster describes all the objects in the Cluster
	DescribeCluster(options DescribeTKGClustersOptions) (DescribeClusterResult, error)
	// DescribeProviders describes all the installed providers
	DescribeProviders() (*clusterctlv1.ProviderList, error)
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
	// RegisterWithTmc registers management cluster with TMC
	RegisterWithTmc(options RegisterOptions) error
	// ScaleCluster scales cluster
	ScaleCluster(options ScaleClusterOptions) error
	// SetCeip sets CEIP to the management cluster
	SetCeip(ceipOptIn, isProd, labels string) error
	// SetMachineHealthCheck apply machine health check to the cluster
	SetMachineHealthCheck(options SetMachineHealthCheckOptions) error
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
	GetTanzuKubernetesReleases(tkrName string) ([]runv1alpha1.TanzuKubernetesRelease, error)
}
