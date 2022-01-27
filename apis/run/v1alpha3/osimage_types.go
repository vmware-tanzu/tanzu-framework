// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// OSInfo describes the "OS" part of the OSImage, defined by the Operating System's name, version and CPU architecture.
type OSInfo struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
}

// MachineImageInfo describes the "Image" part of the OSImage, defined by the image type.
type MachineImageInfo struct {
	// Type of the OSImage, roughly corresponding to the infrastructure provider (vSphere can serve both ova and vmop).
	// Some of currently known types are: "ami", "azure", "docker", "ova", "vmop".
	Type string `json:"type"`

	// Ref is a key-value map identifying the image within the infrastructure provider. This is the data
	// to be injected into the infra-Machine objects (like AWSMachine) on creation.
	Ref map[string]string `json:"ref"`
}

// OSImageSpec defines the desired state of OSImage
type OSImageSpec struct {
	// KubernetesVersion specifies the build version of the Kubernetes shipped with this OSImage.
	KubernetesVersion string `json:"kubernetesVersion"`

	// OS specifies the "OS" part of the OSImage.
	OS OSInfo `json:"os"`

	// Image specifies the "Image" part of the OSImage.
	Image MachineImageInfo `json:"image"`
}

// OSImageStatus defines the observed state of OSImage
type OSImageStatus struct {
	Conditions []clusterv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// OSImage is the schema for the OSImages API.
// OSImage objects represent OSImages shipped as parts of TKRs. OSImages are immutable to end-users.
// They are created and managed by TKG to provide discovery of Kubernetes releases to TKG users and OS image details
// for infrastructure Machines.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=osimages,scope=Cluster,shortName=osimg
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="K8S Version",type=string,JSONPath=.spec.kubernetesVersion
// +kubebuilder:printcolumn:name="OS Name",type=string,JSONPath=.spec.os.name
// +kubebuilder:printcolumn:name="OS Version",type=string,JSONPath=.spec.os.version
// +kubebuilder:printcolumn:name="Arch",type=string,JSONPath=.spec.os.arch
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=.spec.image.type
// +kubebuilder:printcolumn:name="Compatible",type=string,JSONPath=.status.conditions[?(@.type=='Compatible')].status
// +kubebuilder:printcolumn:name="Created",type="date",JSONPath=.metadata.creationTimestamp
type OSImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OSImageSpec   `json:"spec,omitempty"`
	Status OSImageStatus `json:"status,omitempty"`
}

// GetConditions implements capi conditions Getter interface
func (r *OSImage) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

// SetConditions implements capi conditions Setter interface
func (r *OSImage) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

// +kubebuilder:object:root=true

// OSImageList contains a list of OSImage
type OSImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OSImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OSImage{}, &OSImageList{})
}
