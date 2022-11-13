// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

// TKGPackageConfig defines TKG package configuration
type TKGPackageConfig struct {
	Metadata                             Metadata                             `yaml:"metadata"`
	ConfigValues                         map[string]interface{}               `yaml:"configvalues"`
	FrameworkPackage                     FrameworkPackage                     `yaml:"frameworkPackage"`
	ClusterClassPackage                  ClusterClassPackage                  `yaml:"clusterclassPackage"`
	TKRSourceControllerPackage           TKRSourceControllerPackage           `yaml:"tkrSourceControllerPackage"`
	CoreManagementPluginsPackage         CoreManagementPluginsPackage         `yaml:"coreManagementPluginsPackage"`
	AkoOperatorPackage                   AkoOperatorPackage                   `yaml:"akoOperatorPackage"`
	LoadBalancerAndIngressServicePackage LoadBalancerAndIngressServicePackage `yaml:"loadBalancerAndIngressServicePackage"`
}

// Metadata specifies metadata as part of TKG package config
type Metadata struct {
	InfraProvider string `yaml:"infraProvider"`
}

type TKRSourceControllerPackage struct {
	NamespaceForPackageInstallation  string                           `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints               string                           `yaml:"versionConstraints,omitempty"`
	TKRSourceControllerPackageValues TKRSourceControllerPackageValues `yaml:"tkrSourceControllerPackageValues,omitempty"`
}

type TKRSourceControllerPackageValues struct {
	Namespace            string                                     `yaml:"namespace,omitempty"`
	CreateNamespace      string                                     `yaml:"createNamespace,omitempty"`
	VersionConstraints   string                                     `yaml:"versionConstraints,omitempty"`
	BomImagePath         string                                     `yaml:"bomImagePath,omitempty"`
	BomMetadataImagePath string                                     `yaml:"bomMetadataImagePath,omitempty"`
	TKRRepoImagePath     string                                     `yaml:"tkrRepoImagePath,omitempty"`
	DefaultCompatibleTKR string                                     `yaml:"defaultCompatibleTKR,omitempty"`
	Deployment           TKRSourceControllerPackageValuesDeployment `yaml:"deployment,omitempty"`
}

type TKRSourceControllerPackageValuesDeployment struct {
	HttpProxy  string `yaml:"httpProxy,omitempty"`
	HttpsProxy string `yaml:"httpsProxy,omitempty"`
	NoProxy    string `yaml:"noProxy,omitempty"`
}

type FrameworkPackage struct {
	NamespaceForPackageInstallation string                     `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints              string                     `yaml:"versionConstraints,omitempty"`
	FeaturegatePackageValues        FeaturegatePackageValues   `yaml:"featureGatesPackageValues,omitempty"`
	TKRServicePackageValues         TKRServicePackageValues    `yaml:"tkrServicePackageValues,omitempty"`
	CLIPluginsPackageValues         CLIPluginsPackageValues    `yaml:"clipluginsPackageValues,omitempty"`
	AddonsManagerPackageValues      AddonsManagerPackageValues `yaml:"addonsManagerPackageValues,omitempty"`
	TanzuAuthPackageValues          TanzuAuthPackageValues     `yaml:"tanzuAuthPackageValues,omitempty"`
}

type ClusterClassPackage struct {
	NamespaceForPackageInstallation string                         `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints              string                         `yaml:"versionConstraints,omitempty"`
	ClusterClassInfraPackageValues  ClusterClassInfraPackageValues `yaml:"clusterclassInfraPackageValues,omitempty"`
}

type AkoOperatorPackage struct {
	AkoOperatorPackageValues AkoOperatorPackageValues `yaml:"akoOperator,omitempty"`
}

type LoadBalancerAndIngressServicePackage struct {
	LoadBalancerAndIngressServicePackageValues LoadBalancerAndIngressServicePackageValues `yaml:"loadBalancerAndIngressService,omitempty"`
}

type AddonsFeatureGates struct {
	ClusterBootstrapController bool `yaml:"clusterBootstrapController,omitempty"`
	PackageInstallStatus       bool `yaml:"packageInstallStatus,omitempty"`
}

type TanzuAddonsManager struct {
	FeatureGates AddonsFeatureGates `yaml:"featureGates,omitempty"`
}
type CoreManagementPluginsPackage struct {
	NamespaceForPackageInstallation   string                            `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints                string                            `yaml:"versionConstraints,omitempty"`
	CoreManagementPluginsPackageValue CoreManagementPluginsPackageValue `yaml:"clusterclassInfraPackageValues,omitempty"`
}

type CoreManagementPluginsPackageValue struct {
	DeployCLIPluginCRD bool `yaml:"deployCLIPluginCRD,omitempty"`
}

type AddonsManagerPackageValues struct {
	VersionConstraints string             `yaml:"versionConstraints,omitempty"`
	TanzuAddonsManager TanzuAddonsManager `yaml:"tanzuAddonsManager,omitempty"`
}

type FeaturegatePackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

type TKRServicePackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

type CLIPluginsPackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

type ClusterClassInfraPackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

type TanzuAuthPackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

// AkoOperatorPackageValues
type AkoOperatorPackageValues struct {
	AviEnable         bool              `yaml:"avi_enable,omitempty"`
	ClusterName       string            `yaml:"cluster_name,omitempty"`
	AkoOperatorConfig AkoOperatorConfig `yaml:"config,omitempty"`
}

// LoadBalancerAndIngressServicePackageValues
type LoadBalancerAndIngressServicePackageValues struct {
	Name                                string                              `yaml:"name,omitempty"`
	Namespace                           string                              `yaml:"namespace,omitempty"`
	LoadBalancerAndIngressServiceConfig LoadBalancerAndIngressServiceConfig `yaml:"config,omitempty"`
}

// AkoOperatorConfig
type AkoOperatorConfig struct {
	AviControllerAddress                           string `yaml:"avi_controller,omitempty"`
	AviControllerVersion                           string `yaml:"avi_controller_version,omitempty"`
	AviControllerUsername                          string `yaml:"avi_username,omitempty"`
	AviControllerPassword                          string `yaml:"avi_password,omitempty"`
	AviControllerCA                                string `yaml:"avi_ca_data_b64,omitempty"`
	AviCloudName                                   string `yaml:"avi_cloud_name,omitempty"`
	AviServiceEngineGroup                          string `yaml:"avi_service_engine_group,omitempty"`
	AviManagementClusterServiceEngineGroup         string `yaml:"avi_management_cluster_service_engine_group,omitempty"`
	AviDataPlaneNetworkName                        string `yaml:"avi_data_network,omitempty"`
	AviDataPlaneNetworkCIDR                        string `yaml:"avi_data_network_cidr,omitempty"`
	AviControlPlaneNetworkName                     string `yaml:"avi_control_plane_network,omitempty"`
	AviControlPlaneNetworkCIDR                     string `yaml:"avi_control_plane_network_cidr,omitempty"`
	AviManagementClusterDataPlaneNetworkName       string `yaml:"avi_management_cluster_vip_network_name,omitempty"`
	AviManagementClusterDataPlaneNetworkCIDR       string `yaml:"avi_management_cluster_vip_network_cidr,omitempty"`
	AviManagementClusterControlPlaneVipNetworkName string `yaml:"avi_management_cluster_control_plane_vip_network_name,omitempty"`
	AviManagementClusterControlPlaneVipNetworkCIDR string `yaml:"avi_management_cluster_control_plane_vip_network_cidr,omitempty"`
	AviControlPlaneHaProvider                      bool   `yaml:"avi_control_plane_ha_provider,omitempty"`
}

// NetworkSettings
type NetworkSettings struct {
	VipNetworkList string `yaml:"vip_network_list,omitempty"`
}

// ControllerSettings
type ControllerSettings struct {
	ServiceEngineGroupName string `yaml:"service_engine_group_name,omitempty"`
	CloudName              string `yaml:"cloud_name,omitempty"`
	ControllerIp           string `yaml:"controller_ip,omitempty"`
}

// AviCredentials
type AviCredentials struct {
	Username                 string `yaml:"username,omitempty"`
	Password                 string `yaml:"password,omitempty"`
	CertificateAuthorityData string `yaml:"certificate_authority_data,omitempty"`
}

// AkoSettings
type AkoSettings struct {
	ClusterName            string `yaml:"cluster_name,omitempty"`
	DisableStaticRouteSync string `yaml:"disable_static_route_sync,omitempty"`
	CniPlugin              string `yaml:"cni_plugin,omitempty"`
}

// LoadBalancerAndIngressServiceConfig
type LoadBalancerAndIngressServiceConfig struct {
	NetworkSettings    NetworkSettings    `yaml:"network_settings,omitempty"`
	ControllerSettings ControllerSettings `yaml:"controller_settings,omitempty"`
	AviCredentials     AviCredentials     `yaml:"avi_credentials,omitempty"`
	AkoSettings        AkoSettings        `yaml:"ako_settings,omitempty"`
}
