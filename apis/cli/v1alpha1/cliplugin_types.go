// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Target is the namespace of the CLI to which plugin is applicable
type Target string

const (
	// TargetK8s is a kubernetes target of the CLI
	TargetK8s Target = "k8s"

	// TargetTMC is a Tanzu Mission Control target of the CLI
	TargetTMC Target = "tmc"

	// TargetNone is used for plugins that are not associated with any target
	TargetNone Target = ""
)

// ArtifactList contains an Artifact object for every supported platform of a version.
type ArtifactList []Artifact

// Artifact points to an individual plugin binary specific to a version and platform.
type Artifact struct {
	// Image is a fully qualified OCI image for the plugin binary.
	Image string `json:"image,omitempty"`
	// AssetURI is a URI of the plugin binary. This can be a fully qualified HTTP path or a local path.
	URI string `json:"uri,omitempty"`
	// SHA256 hash of the plugin binary.
	Digest string `json:"digest,omitempty"`
	// Type of the binary artifact. Valid values are S3, GCP, OCIImage.
	Type string `json:"type"`
	// OS of the plugin binary in `GOOS` format.
	OS string `json:"os"`
	// Arch is CPU architecture of the plugin binary in `GOARCH` format.
	Arch string `json:"arch"`
}

// CLIPluginSpec defines the desired state of CLIPlugin.
type CLIPluginSpec struct {
	// Description is the plugin's description.
	Description string `json:"description"`
	// Recommended version that Tanzu CLI should use if available.
	// The value should be a valid semantic version as defined in
	// https://semver.org/. E.g., 2.0.1
	RecommendedVersion string `json:"recommendedVersion"`
	// Artifacts contains an artifact list for every supported version.
	Artifacts map[string]ArtifactList `json:"artifacts"`
	// Optional specifies whether the plugin is mandatory or optional
	// If optional, the plugin will not get auto-downloaded as part of
	// `tanzu login` or `tanzu plugin sync` command
	// To view the list of plugin, user can use `tanzu plugin list` and
	// to download a specific plugin run, `tanzu plugin install <plugin-name>`
	Optional bool `json:"optional"`
	// Target specifies the target of the plugin. Only needed for standalone plugins
	Target Target `json:"target"`
}

//+kubebuilder:object:root=true

// CLIPlugin denotes a Tanzu cli plugin.
type CLIPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CLIPluginSpec `json:"spec"`
}

//+kubebuilder:object:root=true

// CLIPluginList contains a list of CLIPlugin
type CLIPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CLIPlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CLIPlugin{}, &CLIPluginList{})
}
