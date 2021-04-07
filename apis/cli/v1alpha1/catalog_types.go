// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Distro is the Schema for the catalogs API
type Distro []string

// CmdGroup is a group of CLI commands.
type CmdGroup string

// PluginCompletionType is the mechanism used for determining command line completion options.
type PluginCompletionType int

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
}

// +kubebuilder:object:root=true

// Catalog is the Schema for the catalogs API
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// PluginDescriptors is a list of PluginDescriptor
	PluginDescriptors []*PluginDescriptor `json:"pluginDescriptors,omitempty" yaml:"pluginDescriptors"`
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
