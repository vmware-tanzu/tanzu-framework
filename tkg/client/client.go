// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package client implements core functionality of the tkg client
package client

import (
	"time"

	"github.com/pkg/errors"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	clusterctltree "sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"

	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/features"
	"github.com/vmware-tanzu/tanzu-framework/tkg/kind"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

type (
	// Provider CAPI provider interface
	Provider clusterctlconfig.Provider
	// Components CAPI repostory components interface
	Components repository.Components
)

// ClusterConfigOptions contains options required to generate a cluster configuration
type ClusterConfigOptions clusterctl.GetClusterTemplateOptions

// CreateClusterOptions contains options required to create a cluster
type CreateClusterOptions struct {
	ClusterConfigOptions
	TKRVersion                   string
	NodeSizeOptions              NodeSizeOptions
	CniType                      string
	ClusterOptionsEnableList     []string
	VsphereControlPlaneEndpoint  string
	SkipValidation               bool
	IsWindowsWorkloadCluster     bool
	ClusterType                  TKGClusterType
	Edition                      string
	IsInputFileClusterClassBased bool
	ClusterConfigFile            string
}

// InitRegionOptions contains options supported by InitRegion
type InitRegionOptions struct {
	NodeSizeOptions              NodeSizeOptions
	ClusterConfigFile            string
	Kubeconfig                   string
	Plan                         string
	ClusterName                  string
	CoreProvider                 string
	BootstrapProvider            string
	InfrastructureProvider       string
	ControlPlaneProvider         string
	Namespace                    string
	CniType                      string
	VsphereControlPlaneEndpoint  string
	Edition                      string
	Annotations                  map[string]string
	Labels                       map[string]string
	FeatureFlags                 map[string]string
	LaunchUI                     bool
	CeipOptIn                    bool
	UseExistingCluster           bool
	IsInputFileClusterClassBased bool
}

// DeleteRegionOptions contains options supported by DeleteRegion
type DeleteRegionOptions struct {
	Kubeconfig         string
	ClusterName        string
	Force              bool
	UseExistingCluster bool
}

//go:generate counterfeiter -o ../fakes/featureflagclient.go --fake-name FeatureFlagClient . FeatureFlagClient

// FeatureFlagClient is used to check if a feature is active
type FeatureFlagClient interface {
	IsConfigFeatureActivated(featurePath string) (bool, error)
}

//go:generate counterfeiter -o ../fakes/client.go --fake-name Client . Client

// Client is used to interact with the tkg client library
type Client interface {
	// InitRegion creates and initializes a management cluster via a
	// self-provisioned bootstrap cluster if necessary
	InitRegion(options *InitRegionOptions) error
	// InitRegionDryRun generates the management cluster manifest that would be
	// used by InitRegion to provision a new cluster
	InitRegionDryRun(options *InitRegionOptions) ([]byte, error)
	// GetClusterConfiguration returns a cluster configuration generated with
	// parameters provided in the set of provided template options
	GetClusterConfiguration(options *CreateClusterOptions) ([]byte, error)
	// ConfigureAndValidateTkrVersion takes tkrVersion, if empty fetches default tkr & k8s version from config
	// and validates k8s version format is valid semantic version
	ConfigureAndValidateTkrVersion(tkrVersion string) (string, string, error)
	// CreateCluster creates a workload cluster based on a cluster template
	// generated from the provided options. Returns whether cluster creation was attempted
	// information along with error information
	CreateCluster(options *CreateClusterOptions, waitForCluster bool) (attempedClusterCreation bool, err error)
	// CreateAWSCloudFormationStack create aws cloud formation stack
	CreateAWSCloudFormationStack() error
	// DeleteRegion deletes management cluster via a self-provisioned kind cluster
	DeleteRegion(options DeleteRegionOptions) error
	// VerifyRegion checks if the kube context points to a management clusters,
	VerifyRegion(kubeConfigPath string) (region.RegionContext, error)
	// AddRegionContext adds a management cluster context to tkg config file
	AddRegionContext(region region.RegionContext, overwrite bool, useDirectReference bool) error
	// GetRegionContexts gets all tkg managed management cluster context context
	GetRegionContexts(clusterName string) ([]region.RegionContext, error)
	// SetRegionContext sets a management cluster context to be current context
	SetRegionContext(clusterName string, contextName string) error
	// GenerateAWSCloudFormationTemplate generates a CloudFormation YAML template
	GenerateAWSCloudFormationTemplate() (string, error)
	// GetCurrentRegionContext() gets the current management cluster context
	GetCurrentRegionContext() (region.RegionContext, error)
	// GetWorkloadClusterCredentials merges workload cluster credentials into kubeconfig path
	GetWorkloadClusterCredentials(options GetWorkloadClusterCredentialsOptions) (string, string, error)
	// ListTKGClusters lists workload clusters managed by the management cluster
	ListTKGClusters(options ListTKGClustersOptions) ([]ClusterInfo, error)
	// DeleteWorkloadCluster deletes a workload cluster managed by the management cluster
	DeleteWorkloadCluster(options DeleteWorkloadClusterOptions) error
	// ScaleCluster scales the cluster
	ScaleCluster(options ScaleClusterOptions) error
	// UpgradeCluster upgrades tkg cluster to specific kubernetes version
	UpgradeCluster(options *UpgradeClusterOptions) error
	// ConfigureAndValidateManagementClusterConfiguration validates the management cluster configuration
	// User is expected to validate the configuration before creating management cluster using init operation
	ConfigureAndValidateManagementClusterConfiguration(options *InitRegionOptions, skipValidation bool) *ValidationError
	// UpgradeManagementCluster upgrades tkg cluster to specific kubernetes version
	UpgradeManagementCluster(options *UpgradeClusterOptions) error
	// Opt-in/out to CEIP on Management Cluster
	SetCEIPParticipation(ceipOptIn bool, isProd string, labels string) error
	// Get opt-in/out status for CEIP on all Management Clusters
	GetCEIPParticipation() (ClusterCeipInfo, error)
	// DeleteMachineHealthCheck deletes MachineHealthCheck for the given cluster
	DeleteMachineHealthCheck(options MachineHealthCheckOptions) error
	// ListMachineHealthChecks lists all the machine health check
	GetMachineHealthChecks(options MachineHealthCheckOptions) ([]MachineHealthCheck, error)
	// IsPacificManagementCluster checks if the cluster pointed to by kubeconfig is Pacific management cluster(supervisor)
	IsPacificManagementCluster() (bool, error)
	// SetMachineHealthCheck create or update a machine health check object
	SetMachineHealthCheck(options *SetMachineHealthCheckOptions) error
	// GetMachineDeployments gets a list of MachineDeployments for a cluster
	GetMachineDeployments(options GetMachineDeploymentOptions) ([]capi.MachineDeployment, error)
	// GetPacificMachineDeployments gets machine deployments from a Pacific cluster
	// Note: This would be soon deprecated after TKGS and TKGm adopt the clusterclass
	GetPacificMachineDeployments(options GetMachineDeploymentOptions) ([]capiv1alpha3.MachineDeployment, error)
	// SetMachineDeployment create machine deployment in a cluster
	SetMachineDeployment(options *SetMachineDeploymentOptions) error
	// DeleteMachineDeployment deletes a machine deployment in a cluster
	DeleteMachineDeployment(options DeleteMachineDeploymentOptions) error
	// GetKubernetesVersions returns the supported k8s versions for workload cluster
	GetKubernetesVersions() (*KubernetesVersionsInfo, error)
	// ParseHiddenArgsAsFeatureFlags adds the hidden flags from InitRegionOptions as enabled feature flags
	ParseHiddenArgsAsFeatureFlags(options *InitRegionOptions)
	// SaveFeatureFlags saves the feature flags to the config file via featuresClient
	SaveFeatureFlags(featureFlags map[string]string) error
	// ValidatePrerequisites validates prerequisites for init command
	ValidatePrerequisites(validateDocker, validateKubectl bool) error
	// ValidateDockerResourcePrerequisites validates resource prerequisites for docker
	ValidateDockerResourcePrerequisites() error
	// GetVSphereEndpoint creates the vSphere client using the credentials from the management cluster if cluster client is provided,
	// otherwise, the vSphere client will be created from the credentials set in the user's environment.
	GetVSphereEndpoint(client clusterclient.Client) (vc.Client, error)
	// ConfigureTimeout updates/configures timeout already set in the tkgClient
	ConfigureTimeout(timeout time.Duration)
	// TKGConfigReaderWriter returns tkgConfigReaderWriter client
	TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter
	// UpdateManagementCluster updates a management cluster
	UpdateCredentialsRegion(options *UpdateCredentialsOptions) error
	// UpdateCredentialsCluster updates a workload cluster
	UpdateCredentialsCluster(options *UpdateCredentialsOptions) error
	// GetClusterPinnipedInfo returns the cluster and pinniped info
	GetClusterPinnipedInfo(options GetClusterPinnipedInfoOptions) (*ClusterPinnipedInfo, error)
	// DescribeCluster describes all the objects in the Cluster
	DescribeCluster(options DescribeTKGClustersOptions) (*clusterctltree.ObjectTree, *capi.Cluster, *clusterctlv1.ProviderList, error)
	// DescribeProvider describes all the installed providers
	DescribeProvider() (*clusterctlv1.ProviderList, error)
	// DownloadBomFile downloads BomFile from management cluster's config map
	DownloadBomFile(tkrName string) error
	// IsManagementClusterAKindCluster determines if the creation of management cluster is successful
	IsManagementClusterAKindCluster(clusterName string) (bool, error)
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
	// IsPacificRegionalCluster checks if the cluster pointed to by kubeconfig  is Pacific management cluster(supervisor)
	IsPacificRegionalCluster() (bool, error)
	// GetPacificClusterObject gets Pacific cluster object
	GetPacificClusterObject(clusterName, namespace string) (*tkgsv1alpha2.TanzuKubernetesCluster, error)
	// IsFeatureActivated checks if a given feature flag is active
	IsFeatureActivated(feature string) bool
}

// TkgClient implements Client.
type TkgClient struct {
	clusterctlClient         clusterctl.Client
	kindClient               kind.Client
	readerwriterConfigClient tkgconfigreaderwriter.Client
	regionManager            region.Manager
	tkgConfigDir             string
	timeout                  time.Duration
	featuresClient           features.Client
	tkgConfigProvidersClient tkgconfigproviders.Client
	tkgBomClient             tkgconfigbom.Client
	tkgConfigUpdaterClient   tkgconfigupdater.Client
	tkgConfigPathsClient     tkgconfigpaths.Client
	clusterKubeConfig        *types.ClusterKubeConfig
	clusterClientFactory     clusterclient.ClusterClientFactory
	vcClientFactory          vc.VcClientFactory
	featureFlagClient        FeatureFlagClient
}

// Options new client options
type Options struct {
	ClusterCtlClient         clusterctl.Client
	ReaderWriterConfigClient tkgconfigreaderwriter.Client
	RegionManager            region.Manager
	TKGConfigDir             string
	Timeout                  time.Duration
	FeaturesClient           features.Client
	TKGConfigProvidersClient tkgconfigproviders.Client
	TKGBomClient             tkgconfigbom.Client
	TKGConfigUpdater         tkgconfigupdater.Client
	TKGPathsClient           tkgconfigpaths.Client
	ClusterKubeConfig        *types.ClusterKubeConfig
	ClusterClientFactory     clusterclient.ClusterClientFactory
	VcClientFactory          vc.VcClientFactory
	FeatureFlagClient        FeatureFlagClient
}

// ensure tkgClient implements Client.
var _ Client = &TkgClient{}

// New returns a tkgClient.
func New(options Options) (*TkgClient, error) { // nolint:gocritic
	err := options.TKGConfigUpdater.DecodeCredentialsInViper()
	if err != nil {
		return nil, errors.Wrap(err, "unable to update encoded credentials")
	}

	// Set default configuration
	options.TKGConfigUpdater.SetDefaultConfiguration()

	return &TkgClient{
		clusterctlClient:         options.ClusterCtlClient,
		kindClient:               nil,
		readerwriterConfigClient: options.ReaderWriterConfigClient,
		regionManager:            options.RegionManager,
		tkgConfigDir:             options.TKGConfigDir,
		timeout:                  options.Timeout,
		featuresClient:           options.FeaturesClient,
		tkgConfigProvidersClient: options.TKGConfigProvidersClient,
		tkgBomClient:             options.TKGBomClient,
		tkgConfigUpdaterClient:   options.TKGConfigUpdater,
		tkgConfigPathsClient:     options.TKGPathsClient,
		clusterKubeConfig:        options.ClusterKubeConfig,
		clusterClientFactory:     options.ClusterClientFactory,
		vcClientFactory:          options.VcClientFactory,
		featureFlagClient:        options.FeatureFlagClient,
	}, nil
}

// TKGConfigReaderWriter returns tkgConfigReaderWriter client
func (c *TkgClient) TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter {
	return c.readerwriterConfigClient.TKGConfigReaderWriter()
}

// IsFeatureActivated checkes if a feature flag is set to "true"
func (c *TkgClient) IsFeatureActivated(feature string) bool {
	result, err := c.featureFlagClient.IsConfigFeatureActivated(feature)
	if err != nil {
		return false
	}
	return result
}
