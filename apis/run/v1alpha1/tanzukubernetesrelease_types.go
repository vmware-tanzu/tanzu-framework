// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ConditionCompatible indicates the condition is compatible.
	ConditionCompatible = "Compatible"

	// ConditionUpdatesAvailable indicates the condition is an updates is available.
	ConditionUpdatesAvailable = "UpdatesAvailable"

	// ConditionUpgradeAvailable (deprecated) indicates the condition is an updates is available.
	ConditionUpgradeAvailable = "UpgradeAvailable"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TanzuKubernetesReleaseSpec defines the desired state of TanzuKubernetesRelease
type TanzuKubernetesReleaseSpec struct {
	// Version is the fully qualified Semantic Versioning conformant version of the TanzuKubernetesRelease.
	// Version MUST be unique across all TanzuKubernetesRelease objects.
	Version string `json:"version"`

	// KubernetesVersion is the fully qualified Semantic Versioning conformant version of Kubernetes shipped by this TKR.
	// The same KubernetesVersion MAY be shipped by different TKRs.
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// Repository is the container image repository for Kubernetes core images, such as kube-apiserver, kube-proxy, etc.
	// It MUST be a DNS-compatible name.
	// +optional
	Repository string `json:"repository,omitempty"`

	// Images is the list of (other than Kubernetes core) essential container images shipped by this TKR (e.g. coredns, etcd).
	// +optional
	Images []ContainerImage `json:"images,omitempty"`

	// NodeImageRef refers to an object representing the image used to create TKC nodes (e.g. VirtualMachineImage).
	// +optional
	NodeImageRef *corev1.ObjectReference `json:"nodeImageRef,omitempty"`
}

// TanzuKubernetesReleaseStatus defines the observed state of TanzuKubernetesRelease
type TanzuKubernetesReleaseStatus struct {
	Conditions []clusterv1.Condition `json:"conditions,omitempty"`
}

// ContainerImage is a struct representing a single fully qualified container image name, constructed as
// `{Repository}/{Name}:{Tag}`.
type ContainerImage struct {
	// Repository is the container image repository used by this image. It MUST be a DNS-compatible name.
	// +optional
	Repository string `json:"repository,omitempty"`

	// Name is the container image name without the repository prefix.
	// It MUST be a valid URI path, MAY contain zero or more '/', and SHOULD NOT start or end with '/'.
	Name string `json:"name"`

	// Tag is the container image version tag. It is the suffix coming after ':' in a fully qualified image name.
	// +optional
	Tag string `json:"tag,omitempty"`
}

func (ci ContainerImage) String() string {
	var prefix, suffix string
	if ci.Repository != "" {
		prefix = ci.Repository + "/"
	}
	if ci.Tag != "" {
		suffix = ":" + ci.Tag
	}
	return prefix + ci.Name + suffix
}

// +kubebuilder:object:root=true

// TanzuKubernetesRelease is the schema for the tanzukubernetesreleases API.
// TanzuKubernetesRelease objects represent Kubernetes releases available via TKG, which can be used to create
// TanzuKubernetesCluster instances. TKRs are immutable to end-users. They are created and managed by TKG to
// provide discovery of Kubernetes releases to TKG users.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tanzukubernetesreleases,scope=Cluster,shortName=tkr
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=.spec.version
// +kubebuilder:printcolumn:name="Compatible",type=string,JSONPath=.status.conditions[?(@.type=='Compatible')].status
// +kubebuilder:printcolumn:name="Created",type="date",JSONPath=.metadata.creationTimestamp
type TanzuKubernetesRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TanzuKubernetesReleaseSpec   `json:"spec,omitempty"`
	Status TanzuKubernetesReleaseStatus `json:"status,omitempty"`
}

// GetConditions implements capi conditions Getter interface
func (r *TanzuKubernetesRelease) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

// SetConditions implements capi conditions Setter interface
func (r *TanzuKubernetesRelease) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

// +kubebuilder:object:root=true

// TanzuKubernetesReleaseList contains a list of TanzuKubernetesRelease
type TanzuKubernetesReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuKubernetesRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TanzuKubernetesRelease{}, &TanzuKubernetesReleaseList{})
}
