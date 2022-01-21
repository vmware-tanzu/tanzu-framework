// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretGenControllerConfigSpec defines the desired state of SecretGenControllerConfig
type SecretGenControllerConfigSpec struct {
	// The namespace in which to deploy secretgen-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=secretgen-controller
	Namespace string `json:"namespace,omitempty"`

	// Whether to create namespace specified for secretgen-controller
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=true
	CreateNamespace bool `json:"createNamespace,omitempty"`
}

// SecretGenControllerConfigStatus defines the observed state of SecretGenControllerConfig
type SecretGenControllerConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SecretGenControllerConfig is the Schema for the secretgencontrollerconfigs API
type SecretGenControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretGenControllerConfigSpec   `json:"spec,omitempty"`
	Status SecretGenControllerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretGenControllerConfigList contains a list of SecretGenControllerConfig
type SecretGenControllerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretGenControllerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretGenControllerConfig{}, &SecretGenControllerConfigList{})
}
