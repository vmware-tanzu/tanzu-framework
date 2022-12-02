// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerType is the type of server.
// Deprecation targeted for a a future version. Superseded by ContextType.
type ServerType string

// ContextType is the type of the context (control plane).
// TODO(anuj): Rename to Target
type ContextType string

const (
	// ManagementClusterServerType is a management cluster server.
	// Deprecation targeted for a a future version. Superseded by CtxTypeK8s.
	ManagementClusterServerType ServerType = "managementcluster"

	// GlobalServerType is a global control plane server.
	// Deprecation targeted for a a future version. Superseded by CtxTypeTMC.
	GlobalServerType ServerType = "global"

	// CtxTypeK8s is a kubernetes cluster API server.
	CtxTypeK8s ContextType = "k8s"

	// CtxTypeTMC is a Tanzu Mission Control server.
	CtxTypeTMC ContextType = "tmc"
)

var (
	// SupportedCtxTypes is a list of all supported Context Types
	SupportedCtxTypes = []ContextType{CtxTypeK8s, CtxTypeTMC}
)

// Server connection.
// Deprecation targeted for a future version. Superseded by Context.
type Server struct {
	// Name of the server.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Type of the endpoint.
	Type ServerType `json:"type,omitempty" yaml:"type,omitempty"`

	// GlobalOpts if the server is global.
	GlobalOpts *GlobalServer `json:"globalOpts,omitempty" yaml:"globalOpts,omitempty"`

	// ManagementClusterOpts if the server is a management cluster.
	ManagementClusterOpts *ManagementClusterServer `json:"managementClusterOpts,omitempty" yaml:"managementClusterOpts,omitempty"`

	// DiscoverySources determines from where to discover plugins
	// associated with this server
	DiscoverySources []PluginDiscovery `json:"discoverySources,omitempty" yaml:"discoverySources,omitempty"`
}

// Context configuration for a control plane. This can one of the following,
// 1. Kubernetes Cluster
// 2. Tanzu Mission Control endpoint
type Context struct {
	// Name of the context.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Type of the context.
	Type ContextType `json:"type,omitempty" yaml:"type,omitempty"`

	// GlobalOpts if the context is a global control plane (e.g., TMC).
	GlobalOpts *GlobalServer `json:"globalOpts,omitempty" yaml:"globalOpts,omitempty"`

	// ClusterOpts if the context is a kubernetes cluster.
	ClusterOpts *ClusterServer `json:"clusterOpts,omitempty" yaml:"clusterOpts,omitempty"`

	// DiscoverySources determines from where to discover plugins
	// associated with this context.
	DiscoverySources []PluginDiscovery `json:"discoverySources,omitempty" yaml:"discoverySources,omitempty"`
}

// ManagementClusterServer is the configuration for a management cluster kubeconfig.
// Deprecation targeted for a a future version. Superseded by ClusterServer.
type ManagementClusterServer struct {
	// Endpoint for the login.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Path to the kubeconfig.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// The context to use (if required), defaults to current.
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
}

// ClusterServer contains the configuration for a kubernetes cluster (kubeconfig).
type ClusterServer struct {
	// Endpoint for the login.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Path to the kubeconfig.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// The kubernetes context to use (if required), defaults to current.
	Context string `json:"context,omitempty" yaml:"context,omitempty"`

	// Denotes whether this server is a management cluster or not (workload cluster).
	IsManagementCluster bool `json:"isManagementCluster,omitempty" yaml:"isManagementCluster,omitempty"`
}

// GlobalServer is the configuration for a global server.
type GlobalServer struct {
	// Endpoint for the server.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Auth for the global server.
	Auth GlobalServerAuth `json:"auth,omitempty" yaml:"auth,omitempty"`
}

// GlobalServerAuth is authentication for a global server.
type GlobalServerAuth struct {
	// Issuer url for IDP, compliant with OIDC Metadata Discovery.
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty"`

	// UserName is the authorized user the token is assigned to.
	UserName string `json:"userName,omitempty" yaml:"userName,omitempty"`

	// Permissions are roles assigned to the user.
	Permissions []string `json:"permissions,omitempty" yaml:"permissions,omitempty"`

	// AccessToken is the current access token based on the context.
	AccessToken string `json:"accessToken,omitempty" yaml:"accessToken,omitempty"`

	// IDToken is the current id token based on the context scoped to the CLI.
	IDToken string `json:"IDToken,omitempty" yaml:"IDToken,omitempty"`

	// RefreshToken will be stored only in case of api-token login flow.
	RefreshToken string `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`

	// Expiration times of the token.
	Expiration metav1.Time `json:"expiration,omitempty" yaml:"expiration,omitempty"`

	// Type of the token (user or client).
	Type string `json:"type" yaml:"type,omitempty"`
}

// ClientOptions are the client specific options.
type ClientOptions struct {
	// CLI options specific to the CLI.
	CLI      *CLIOptions           `json:"cli,omitempty" yaml:"cli,omitempty"`
	Features map[string]FeatureMap `json:"features,omitempty" yaml:"features,omitempty"`
	Env      map[string]string     `json:"env,omitempty" yaml:"env,omitempty"`
}

// FeatureMap is simply a hash table, but needs an explicit type to be an object in another hash map (cf ClientOptions.Features)
type FeatureMap map[string]string

// EnvMap is simply a hash table, but needs an explicit type to be an object in another hash map (cf ClientOptions.Env)
type EnvMap map[string]string

// CLIOptions are options for the CLI.
type CLIOptions struct {
	// Repositories are the plugin repositories.
	Repositories []PluginRepository `json:"repositories,omitempty" yaml:"repositories,omitempty"`
	// DiscoverySources determines from where to discover stand-alone plugins
	DiscoverySources []PluginDiscovery `json:"discoverySources,omitempty" yaml:"discoverySources,omitempty"`
	// UnstableVersionSelector determined which version tags are allowed
	UnstableVersionSelector VersionSelectorLevel `json:"unstableVersionSelector,omitempty" yaml:"unstableVersionSelector,omitempty"`
	// Deprecated: Edition has been deprecated and will be removed from future version
	// Edition
	Edition EditionSelector `json:"edition,omitempty" yaml:"edition,omitempty"`
	// Deprecated: BOMRepo has been deprecated and will be removed from future version
	// BOMRepo is the root repository URL used to resolve the compatibiilty file
	// and bill of materials. An example URL is projects.registry.vmware.com/tkg.
	BOMRepo string `json:"bomRepo,omitempty" yaml:"bomRepo,omitempty"`
	// Deprecated: CompatibilityFilePath has been deprecated and will be removed from future version
	// CompatibilityFilePath is the path, from the BOM repo, to download and access the compatibility file.
	// the compatibility file is used for resolving the bill of materials for creating clusters.
	CompatibilityFilePath string `json:"compatibilityFilePath,omitempty" yaml:"compatibilityFilePath,omitempty"`
}

// PluginDiscovery contains a specific distribution mechanism. Only one of the
// configs must be set.
type PluginDiscovery struct {
	// GCPStorage is set if the plugins are to be discovered via Google Cloud Storage.
	GCP *GCPDiscovery `json:"gcp,omitempty" yaml:"gcp,omitempty"`
	// OCIDiscovery is set if the plugins are to be discovered via an OCI Image Registry.
	OCI *OCIDiscovery `json:"oci,omitempty" yaml:"oci,omitempty"`
	// GenericRESTDiscovery is set if the plugins are to be discovered via a REST API endpoint.
	REST *GenericRESTDiscovery `json:"rest,omitempty" yaml:"rest,omitempty"`
	// KubernetesDiscovery is set if the plugins are to be discovered via the Kubernetes API server.
	Kubernetes *KubernetesDiscovery `json:"k8s,omitempty" yaml:"k8s,omitempty"`
	// LocalDiscovery is set if the plugins are to be discovered via Local Manifest fast.
	Local *LocalDiscovery `json:"local,omitempty" yaml:"local,omitempty"`
}

// GCPDiscovery provides a plugin discovery mechanism via a Google Cloud Storage
// bucket with a manifest.yaml file.
type GCPDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Bucket is a Google Cloud Storage bucket.
	// E.g., tanzu-cli
	Bucket string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	// BasePath is a URI path that is prefixed to the object name/path.
	// E.g., plugins/cluster
	ManifestPath string `json:"manifestPath,omitempty" yaml:"manifestPath,omitempty"`
}

// OCIDiscovery provides a plugin discovery mechanism via a OCI Image Registry
type OCIDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Image is an OCI compliant image. Which include DNS-compatible registry name,
	// a valid URI path(MAY contain zero or more ‘/’) and a valid tag.
	// E.g., harbor.my-domain.local/tanzu-cli/plugins-manifest:latest
	// Contains a directory containing YAML files, each of which contains single
	// CLIPlugin API resource.
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}

// GenericRESTDiscovery provides a plugin discovery mechanism via any REST API
// endpoint. The fully qualified list URL is constructed as
// `https://{Endpoint}/{BasePath}` and the get plugin URL is constructed as .
// `https://{Endpoint}/{BasePath}/{Plugin}`.
type GenericRESTDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Endpoint is the REST API server endpoint.
	// E.g., api.my-domain.local
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	// BasePath is the base URL path of the plugin discovery API.
	// E.g., /v1alpha1/cli/plugins
	BasePath string `json:"basePath,omitempty" yaml:"basePath,omitempty"`
}

// KubernetesDiscovery provides a plugin discovery mechanism via the Kubernetes API server.
type KubernetesDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Path to the kubeconfig.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// The context to use (if required), defaults to current.
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
	// Version of the CLIPlugins API to query.
	// E.g., v1alpha1
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// LocalDiscovery is a artifact discovery endpoint utilizing a local host OS.
type LocalDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Path is a local path pointing to directory
	// containing YAML files, each of which contains single
	// CLIPlugin API resource.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// PluginRepository is a CLI plugin repository
type PluginRepository struct {
	// GCPPluginRepository is a plugin repository that utilizes GCP cloud storage.
	GCPPluginRepository *GCPPluginRepository `json:"gcpPluginRepository,omitempty" yaml:"gcpPluginRepository,omitempty"`
}

// GCPPluginRepository is a plugin repository that utilizes GCP cloud storage.
type GCPPluginRepository struct {
	// Name of the repository.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// BucketName is the name of the bucket.
	BucketName string `json:"bucketName,omitempty" yaml:"bucketName,omitempty"`

	// RootPath within the bucket.
	RootPath string `json:"rootPath,omitempty" yaml:"rootPath,omitempty"`
}

// +kubebuilder:object:root=true

// ClientConfig is the Schema for the configs API
type ClientConfig struct {
	metav1.TypeMeta   `json:",omitempty" yaml:",omitempty"`
	metav1.ObjectMeta `json:",omitempty" yaml:",omitempty"`
	// KnownServers available.
	// Deprecation targeted for a future version. Superseded by KnownContexts.
	KnownServers []*Server `json:"servers,omitempty" yaml:"servers,omitempty"`

	// CurrentServer in use.
	// Deprecation targeted for a future version. Superseded by CurrentContext.
	CurrentServer string `json:"current,omitempty" yaml:"current,omitempty"`

	// KnownContexts available.
	KnownContexts []*Context `json:"contexts,omitempty" yaml:"contexts,omitempty"`

	// CurrentContext for every type.
	CurrentContext map[ContextType]string `json:"currentContext,omitempty" yaml:"currentContext,omitempty"`

	// ClientOptions are client specific options like feature flags, env variables, repositories , discoverySources etc.
	ClientOptions *ClientOptions `json:"clientOptions,omitempty" yaml:"clientOptions,omitempty"`
}

// +kubebuilder:object:root=true

// ClientConfigList contains a list of ClientConfig
type ClientConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClientConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClientConfig{}, &ClientConfigList{})
}
