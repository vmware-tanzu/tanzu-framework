// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tanzukubernetesaddons,scope=Cluster,shortName=tka
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Addon",type=string,JSONPath=.spec.addonName
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=.spec.type
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=.spec.version
// +kubebuilder:printcolumn:name="Created",type="date",JSONPath=.metadata.creationTimestamp

// TanzuKubernetesAddon is the schema for the tanzukubernetesaddons API.
// TanzuKubernetesAddon objects represent Kubernetes addons available via TKG Service, which can be used to create
// TanzuKubernetesCluster instances. TKAs are immutable to end-users. They are created and managed by TKG Service to
// provide discovery of Kubernetes addons to TKG Service users.
type TanzuKubernetesAddon struct { // nolint:maligned
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TanzuKubernetesAddonSpec   `json:"spec,omitempty"`
	Status TanzuKubernetesAddonStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TanzuKubernetesAddonList contains a list of TanzuKubernetesAddon objects.
type TanzuKubernetesAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuKubernetesAddon `json:"items"`
}

// TanzuKubernetesAddonSpec defines the desired state of TanzuKubernetesAddon
type TanzuKubernetesAddonSpec struct {
	// AddonName is the generic name of this addon, e.g. "antrea", "calico", "pvcsi", etc.
	AddonName string `json:"addonName"`

	// Version is the fully qualified Semantic Versioning conformant version of the TanzuKubernetesAddon.
	// If set, Version MUST be unique across all TanzuKubernetesAddon objects with the same `addonName`.
	// +optional
	Version string `json:"version,omitempty"`

	// Repository is the default container image repository used by Images. It MUST be a DNS-compatible name.
	// +optional
	Repository string `json:"repository,omitempty"`

	// Images is the list of container images shipped by this addon (e.g. coredns, etcd).
	Images []ContainerImage `json:"images,omitempty"`

	// Resource contains the YAML manifest for installing the addon.
	Resource *ManifestResource `json:"resource,omitempty"`
}

// ManifestResource represents a YAML manifest for installing an addon.
type ManifestResource struct {
	// Version is the addon version.
	// +optional
	Version string `json:"version,omitempty"`

	// Type is the type of the manifest resource. In VirtualMachineImage based addons its value is 'inline'.
	Type string `json:"type"`

	// Value is the text of the YAML manifest.
	Value string `json:"value"`
}

// TanzuKubernetesAddonStatus defines the observed state of TanzuKubernetesAddon
type TanzuKubernetesAddonStatus struct {
}

func init() {
	SchemeBuilder.Register(&TanzuKubernetesAddon{}, &TanzuKubernetesAddonList{})
}
