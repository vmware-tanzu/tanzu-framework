// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

// TKGPackageConfig defines TKG package configuration
type TKGPackageConfig struct {
	Metadata                     Metadata                     `yaml:"metadata"`
	ConfigValues                 map[string]interface{}       `yaml:"configvalues"`
	FrameworkPackage             FrameworkPackage             `yaml:"frameworkPackage"`
	ClusterClassPackage          ClusterClassPackage          `yaml:"clusterclassPackage"`
	TKRSourceControllerPackage   TKRSourceControllerPackage   `yaml:"tkrSourceControllerPackage"`
	CoreManagementPluginsPackage CoreManagementPluginsPackage `yaml:"coreManagementPluginsPackage"`
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
	CaCerts              string                                     `yaml:"caCerts,omitempty"`
	ImageRepo            string                                     `yaml:"imageRepository,omitempty"`
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
