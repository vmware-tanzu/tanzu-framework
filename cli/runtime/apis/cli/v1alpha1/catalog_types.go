// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

// Distro is the Schema for the catalogs API
type Distro []string

// CmdGroup is a group of CLI commands.
type CmdGroup string

// PluginCompletionType is the mechanism used for determining command line completion options.
type PluginCompletionType int

// PluginAssociation is a set of plugin names and their associated installation paths.
type PluginAssociation map[string]string

// Add adds plugin entry to the map
func (pa PluginAssociation) Add(pluginName, installationPath string) {
	if pa == nil {
		pa = map[string]string{}
	}
	pa[pluginName] = installationPath
}

// Remove deletes plugin entry from the map
func (pa PluginAssociation) Remove(pluginName string) {
	delete(pa, pluginName)
}

// Get returns installation path for the plugin
// If plugin doesn't exists in map it will return empty string
func (pa PluginAssociation) Get(pluginName string) string {
	return pa[pluginName]
}

// Map returns associated list of plugins as a map
func (pa PluginAssociation) Map() map[string]string {
	return pa
}

// +kubebuilder:object:generate=false

// Hook is the mechanism used to define function for plugin hooks
type Hook func() error

// DeepCopyInto is an deepcopy function implementation of Hook
// currently there is nothing that we need to copy hence keeping this empty
func (in *Hook) DeepCopyInto(out *Hook) {
}

const (
	// NativePluginCompletion indicates command line completion is determined using the built in
	// cobra.Command __complete mechanism.
	NativePluginCompletion PluginCompletionType = iota
	// StaticPluginCompletion indicates command line completion will be done by using a statically
	// defined list of options.
	StaticPluginCompletion
	// DynamicPluginCompletion indicates command line completion will be retrieved from the plugin
	// at runtime.
	DynamicPluginCompletion

	// RunCmdGroup are commands associated with Tanzu Run.
	RunCmdGroup CmdGroup = "Run"

	// ManageCmdGroup are commands associated with Tanzu Manage.
	ManageCmdGroup CmdGroup = "Manage"

	// BuildCmdGroup are commands associated with Tanzu Build.
	BuildCmdGroup CmdGroup = "Build"

	// ObserveCmdGroup are commands associated with Tanzu Observe.
	ObserveCmdGroup CmdGroup = "Observe"

	// SystemCmdGroup are system commands.
	SystemCmdGroup CmdGroup = "System"

	// TargetCmdGroup are various target commands.
	TargetCmdGroup CmdGroup = "Target"

	// VersionCmdGroup are version commands.
	VersionCmdGroup CmdGroup = "Version"

	// AdminCmdGroup are admin commands.
	AdminCmdGroup CmdGroup = "Admin"

	// TestCmdGroup is the test command group.
	TestCmdGroup CmdGroup = "Test"

	// ExtraCmdGroup is the extra command group.
	ExtraCmdGroup CmdGroup = "Extra"
)

// PluginDescriptor describes a plugin binary.
type PluginDescriptor struct {
	// Name is the name of the plugin.
	Name string `json:"name" yaml:"name"`

	// Description is the plugin's description.
	Description string `json:"description" yaml:"description"`

	// Version of the plugin. Must be a valid semantic version https://semver.org/
	Version string `json:"version" yaml:"version"`

	// BuildSHA is the git commit hash the plugin was built with.
	BuildSHA string `json:"buildSHA" yaml:"buildSHA"`

	// Digest is the SHA256 hash of the plugin binary.
	Digest string `json:"digest" yaml:"digest"`

	// Command group for the plugin.
	Group CmdGroup `json:"group" yaml:"group"`

	// DocURL for the plugin.
	DocURL string `json:"docURL" yaml:"docURL"`

	// Hidden tells whether the plugin should be hidden from the help command.
	Hidden bool `json:"hidden,omitempty" yaml:"hidden,omitempty"`

	// CompletionType determines how command line completion will be determined.
	CompletionType PluginCompletionType `json:"completionType" yaml:"completionType"`

	// CompletionArgs contains the valid command line completion values if `CompletionType`
	// is set to `StaticPluginCompletion`.
	CompletionArgs []string `json:"completionArgs,omitempty" yaml:"completionArgs,omitempty"`

	// CompletionCommand is the command to call from the plugin to retrieve a list of
	// valid completion nouns when `CompletionType` is set to `DynamicPluginCompletion`.
	CompletionCommand string `json:"completionCmd,omitempty" yaml:"completionCmd,omitempty"`

	// Aliases are other text strings used to call this command
	Aliases []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`

	// InstallationPath is a relative installation path for a plugin binary.
	// E.g., cluster/v0.3.2@sha256:...
	InstallationPath string `json:"installationPath"`

	// Discovery is the name of the discovery from where
	// this plugin is discovered.
	Discovery string `json:"discovery"`

	// Scope is the scope of the plugin. Stand-Alone or Context
	Scope string `json:"scope"`

	// Status is the current plugin installation status
	Status string `json:"status"`

	// DiscoveredRecommendedVersion specifies the recommended version of the plugin that was discovered
	DiscoveredRecommendedVersion string `json:"discoveredRecommendedVersion"`

	// Target specifies the target of the plugin
	Target cliv1alpha1.Target `json:"target"`

	// PostInstallHook is function to be run post install of a plugin.
	PostInstallHook Hook `json:"-" yaml:"-"`

	// DefaultFeatureFlags is default featureflags to be configured if missing when invoking plugin
	DefaultFeatureFlags map[string]bool `json:"defaultFeatureFlags"`
}

// +kubebuilder:object:root=true

// Catalog is the Schema for the catalogs API
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// PluginDescriptors is a list of PluginDescriptor
	PluginDescriptors []*PluginDescriptor `json:"pluginDescriptors,omitempty" yaml:"pluginDescriptors"`

	// IndexByPath of PluginDescriptors for all installed plugins by installation path.
	IndexByPath map[string]PluginDescriptor `json:"indexByPath,omitempty"`
	// IndeByName of all plugin installation paths by name.
	IndexByName map[string][]string `json:"indexByName,omitempty"`
	// StandAlonePlugins is a set of stand-alone plugin installations aggregated across all context types.
	// Note: Shall be reduced to only those stand-alone plugins that are common to all context types.
	StandAlonePlugins PluginAssociation `json:"standAlonePlugins,omitempty"`
	// ServerPlugins links a server and a set of associated plugin installations.
	ServerPlugins map[string]PluginAssociation `json:"serverPlugins,omitempty"`
}

// +kubebuilder:object:root=true

// CatalogList contains a list of Catalog
type CatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catalog `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Catalog{}, &CatalogList{})
}
