// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureFileCSIConfigSpec defines the desired state of AzureFileCSIConfig
type AzureFileCSIConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AzureFileCSIConfig. Edit azurefilecsiconfig_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AzureFileCSIConfigStatus defines the observed state of AzureFileCSIConfig
type AzureFileCSIConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AzureFileCSIConfig is the Schema for the azurefilecsiconfigs API
type AzureFileCSIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureFileCSIConfigSpec   `json:"spec,omitempty"`
	Status AzureFileCSIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureFileCSIConfigList contains a list of AzureFileCSIConfig
type AzureFileCSIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureFileCSIConfig `json:"items"`
}

// AzureFileCSI is the Schema for the AzureFileCSIConfig API
type AzureFileCSI struct {
	// The namespace csi components are to be deployed in
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Optional
	HTTPProxy string `json:"httpProxy,omitempty"`

	// +kubebuilder:validation:Optional
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// +kubebuilder:validation:Optional
	NoProxy string `json:"noProxy,omitempty"`

	// +kubebuilder:validation:Optional
	DeploymentReplicas *int32 `json:"deploymentReplicas,omitempty"`
}

func init() {
	SchemeBuilder.Register(&AzureFileCSIConfig{}, &AzureFileCSIConfigList{})
}