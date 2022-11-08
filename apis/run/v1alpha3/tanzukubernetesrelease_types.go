// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	ConditionCompatible = "Compatible"
	ConditionValid      = "Valid"
	ConditionReady      = "Ready"

	ConditionUpdatesAvailable = "UpdatesAvailable"

	ReasonCannotParseTKR  = "CannotParseTKR"
	ReasonAlreadyUpToDate = "AlreadyUpToDate"

	LabelIncompatible = "incompatible"
	LabelDeactivated  = "deactivated"
	LabelInvalid      = "invalid"

	AnnotationResolveTKR     = "run.tanzu.vmware.com/resolve-tkr"
	AnnotationResolveOSImage = "run.tanzu.vmware.com/resolve-os-image"

	LabelTKR     = "run.tanzu.vmware.com/tkr"
	LabelOSImage = "run.tanzu.vmware.com/os-image"

	LabelLegacyTKR = "run.tanzu.vmware.com/legacy-tkr"
)

// TanzuKubernetesReleaseSpec defines the desired state of TanzuKubernetesRelease
type TanzuKubernetesReleaseSpec struct {
	// Version is the fully qualified Semantic Versioning conformant version of the TanzuKubernetesRelease.
	// Version MUST be unique across all TanzuKubernetesRelease objects.
	Version string `json:"version"`

	// Kubernetes is Kubernetes
	Kubernetes KubernetesSpec `json:"kubernetes"`

	// OSImages lists references to all OSImage objects shipped with this TKR.
	OSImages []corev1.LocalObjectReference `json:"osImages,omitempty"`

	// BootstrapPackages lists references to all bootstrap packages shipped with this TKR.
	BootstrapPackages []corev1.LocalObjectReference `json:"bootstrapPackages,omitempty"`
}

// KubernetesSpec specifies the details about the Kubernetes distribution shipped by this TKR.
type KubernetesSpec struct {
	// Version is Semantic Versioning conformant version of the Kubernetes build shipped by this TKR.
	// The same Kubernetes build MAY be shipped by multiple TKRs.
	Version string `json:"version"`

	// ImageRepository specifies container image registry to pull images from.
	ImageRepository string `json:"imageRepository,omitempty"`

	// Etcd specifies the container image repository and tag for etcd.
	// +optional
	Etcd *ContainerImageInfo `json:"etcd"`

	// Pause specifies the container image repository and tag for pause.
	// +optional
	Pause *ContainerImageInfo `json:"pause"`

	// CoreDNS specifies the container image repository and tag for coredns.
	// +optional
	CoreDNS *ContainerImageInfo `json:"coredns"`

	// KubeVIP specifies the container image repository and tag for kube-vip.
	// +optional
	KubeVIP *ContainerImageInfo `json:"kube-vip"`
}

// ContainerImageInfo allows to customize the image used for components that are not
// originated from the Kubernetes/Kubernetes release process (such as etcd and coredns).
type ContainerImageInfo struct {
	// ImageRepository sets the container registry to pull images from.
	// if not set, defaults to the ImageRepository defined in KubernetesSpec.
	// +optional
	ImageRepository string `json:"imageRepository,omitempty"`

	// ImageTag specifies a tag for the image.
	ImageTag string `json:"imageTag,omitempty"`
}

// TanzuKubernetesReleaseStatus defines the observed state of TanzuKubernetesRelease
type TanzuKubernetesReleaseStatus struct {
	Conditions []clusterv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// TanzuKubernetesRelease is the schema for the tanzukubernetesreleases API.
// TanzuKubernetesRelease objects represent Kubernetes releases available via TKG, which can be used to create
// TanzuKubernetesCluster instances. TKRs are immutable to end-users. They are created and managed by TKG to
// provide discovery of Kubernetes releases to TKG users.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tanzukubernetesreleases,scope=Cluster,shortName=tkr
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=.spec.version
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=.status.conditions[?(@.type=='Ready')].status
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
