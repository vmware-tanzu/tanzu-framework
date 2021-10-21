// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerType is the type of server.
type ServerType string

const (
	// ManagementClusterServerType is a management cluster server.
	ManagementClusterServerType ServerType = "managementcluster"

	// GlobalServerType is a global control plane server.
	GlobalServerType ServerType = "global"
)

// Server connection.
type Server struct {
	// Name of the server.
	Name string `json:"name,omitempty" yaml:"name"`

	// Type of the endpoint.
	Type ServerType `json:"type,omitempty" yaml:"type"`

	// GlobalOpts if the server is global.
	GlobalOpts *GlobalServer `json:"globalOpts,omitempty" yaml:"globalOpts"`

	// ManagementClusterOpts if the server is a management cluster.
	ManagementClusterOpts *ManagementClusterServer `json:"managementClusterOpts,omitempty" yaml:"managementClusterOpts"`

	// DiscoverySources determines from where to discover plugins
	// associated with this server
	DiscoverySources []PluginDiscovery `json:"discoverySources,omitempty" yaml:"discoverySources"`
}

// ManagementClusterServer is the configuration for a management cluster kubeconfig.
type ManagementClusterServer struct {
	// Endpoint for the login.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint"`

	// Path to the kubeconfig.
	Path string `json:"path,omitempty" yaml:"path"`

	// The context to use (if required), defaults to current.
	Context string `json:"context,omitempty" yaml:"context"`
}

// GlobalServer is the configuration for a global server.
type GlobalServer struct {
	// Endpoint for the server.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint"`

	// Auth for the global server.
	Auth GlobalServerAuth `json:"auth,omitempty" yaml:"auth"`
}

// GlobalServerAuth is authentication for a global server.
type GlobalServerAuth struct {
	// Issuer url for IDP, compliant with OIDC Metadata Discovery.
	Issuer string `json:"issuer" yaml:"issuer"`

	// UserName is the authorized user the token is assigned to.
	UserName string `json:"userName" yaml:"userName"`

	// Permissions are roles assigned to the user.
	Permissions []string `json:"permissions" yaml:"permissions"`

	// AccessToken is the current access token based on the context.
	AccessToken string `json:"accessToken" yaml:"accessToken"`

	// IDToken is the current id token based on the context scoped to the CLI.
	IDToken string `json:"IDToken" yaml:"IDToken"`

	// RefreshToken will be stored only in case of api-token login flow.
	RefreshToken string `json:"refresh_token" yaml:"refresh_token"`

	// Expiration times of the token.
	Expiration metav1.Time `json:"expiration" yaml:"expiration"`

	// Type of the token (user or client).
	Type string `json:"type" yaml:"type"`
}

// ClientOptions are the client specific options.
type ClientOptions struct {
	// CLI options specific to the CLI.
	CLI      *CLIOptions           `json:"cli,omitempty" yaml:"cli"`
	Features map[string]FeatureMap `json:"features,omitempty" yaml:"features"`
}

// FeatureMap is simply a hash table, but needs an explicit type to be an object in another hash map (cf ClientOptions.Features)
type FeatureMap map[string]string

// CLIOptions are options for the CLI.
type CLIOptions struct {
	// Repositories are the plugin repositories.
	Repositories []PluginRepository `json:"repositories,omitempty" yaml:"repositories"`
	// DiscoverySources determines from where to discover stand-alone plugins
	DiscoverySources []PluginDiscovery `json:"discoverySources,omitempty" yaml:"discoverySources"`
	// UnstableVersionSelector determined which version tags are allowed
	UnstableVersionSelector VersionSelectorLevel `json:"unstableVersionSelector,omitempty" yaml:"unstableVersionSelector"`
}

// PluginDiscovery contains a specific distribution mechanism. Only one of the
// configs must be set.
type PluginDiscovery struct {
	// GCPStorage is set if the plugins are to be discovered via Google Cloud Storage.
	GCP *GCPDiscovery `json:"gcp,omitempty"`
	// OCIDiscovery is set if the plugins are to be discovered via an OCI Image Registry.
	OCI *OCIDiscovery `json:"oci,omitempty"`
	// GenericRESTDiscovery is set if the plugins are to be discovered via a REST API endpoint.
	REST *GenericRESTDiscovery `json:"rest,omitempty"`
	// KubernetesDiscovery is set if the plugins are to be discovered via the Kubernetes API server.
	Kubernetes *KubernetesDiscovery `json:"k8s,omitempty"`
	// LocalDiscovery is set if the plugins are to be discovered via Local Manifest fast.
	Local *LocalDiscovery `json:"local,omitempty"`
}

// GCPDiscovery provides a plugin discovery mechanism via a Google Cloud Storage
// bucket with a manifest.yaml file.
type GCPDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name"`
	// Bucket is a Google Cloud Storage bucket.
	// E.g., tanzu-cli
	Bucket string `json:"bucket"`
	// BasePath is a URI path that is prefixed to the object name/path.
	// E.g., plugins/cluster
	ManifestPath string `json:"manifestPath"`
}

// OCIDiscovery provides a plugin discovery mechanism via a OCI Image Registry
type OCIDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name"`
	// Image is an OCI compliant image. Which include DNS-compatible registry name,
	// a valid URI path(MAY contain zero or more ‘/’) and a valid tag.
	// E.g., harbor.my-domain.local/tanzu-cli/plugins-manifest:latest
	// Contains a directory containing YAML files, each of which contains single
	// CLIPlugin API resource.
	Image string `json:"image"`
}

// GenericRESTDiscovery provides a plugin discovery mechanism via any REST API
// endpoint. The fully qualified list URL is constructed as
// `https://{Endpoint}/{BasePath}` and the get plugin URL is constructed as .
// `https://{Endpoint}/{BasePath}/{Plugin}`.
type GenericRESTDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name"`
	// Endpoint is the REST API server endpoint.
	// E.g., api.my-domain.local
	Endpoint string `json:"endpoint"`
	// BasePath is the base URL path of the plugin discovery API.
	// E.g., /v1alpha1/cli/plugins
	BasePath string `json:"basePath"`
}

// KubernetesDiscovery provides a plugin discovery mechanism via the Kubernetes API server.
type KubernetesDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name"`
	// Path to the kubeconfig.
	Path string `json:"path"`
	// The context to use (if required), defaults to current.
	Context string `json:"context"`
	// Version of the CLIPlugins API to query.
	// E.g., v1alpha1
	Version string `json:"version"`
}

// LocalDiscovery is a artifact discovery endpoint utilizing a local host OS.
type LocalDiscovery struct {
	// Name is a name of the discovery
	Name string `json:"name"`
	// Path is a local path pointing to directory
	// containing YAML files, each of which contains single
	// CLIPlugin API resource.
	Path string `json:"path"`
}

// PluginRepository is a CLI plugin repository
type PluginRepository struct {
	// GCPPluginRepository is a plugin repository that utilizes GCP cloud storage.
	GCPPluginRepository *GCPPluginRepository `json:"gcpPluginRepository,omitempty" yaml:"gcpPluginRepository"`
}

// GCPPluginRepository is a plugin repository that utilizes GCP cloud storage.
type GCPPluginRepository struct {
	// Name of the repository.
	Name string `json:"name,omitempty" yaml:"name"`

	// BucketName is the name of the bucket.
	BucketName string `json:"bucketName,omitempty" yaml:"bucketName"`

	// RootPath within the bucket.
	RootPath string `json:"rootPath,omitempty" yaml:"rootPath"`
}

// +kubebuilder:object:root=true

// ClientConfig is the Schema for the configs API
type ClientConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// KnownServers available.
	KnownServers []*Server `json:"servers,omitempty" yaml:"servers"`

	// CurrentServer in use.
	CurrentServer string `json:"current,omitempty" yaml:"current"`

	// ClientOptions are client specific options.
	ClientOptions *ClientOptions `json:"clientOptions,omitempty" yaml:"clientOptions"`
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
