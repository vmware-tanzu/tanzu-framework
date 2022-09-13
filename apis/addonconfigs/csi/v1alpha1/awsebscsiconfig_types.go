// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AwsEbsCSIConfigSpec defines the desired state of AwsEbsCSIConfig
type AwsEbsCSIConfigSpec struct {
	AwsEbsCSI AwsEbsCSI `json:"awsEBSCSIDriver"`
}

// AwsEbsCSIConfigStatus defines the observed state of AwsEbsCSIConfig
type AwsEbsCSIConfigStatus struct {
	// Name of the secret created by csi controller
	//+ kubebuilder:validation:Optional
	SecretRef *string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AwsEbsCSIConfig is the Schema for the awsebscsiconfigs API
type AwsEbsCSIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsEbsCSIConfigSpec   `json:"spec,omitempty"`
	Status AwsEbsCSIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AwsEbsCSIConfigList contains a list of AwsEbsCSIConfig
type AwsEbsCSIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsEbsCSIConfig `json:"items"`
}

// AwsEbsCSI is the Schema for the awsebscsiconfigs API
type AwsEbsCSI struct {
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
	SchemeBuilder.Register(&AwsEbsCSIConfig{}, &AwsEbsCSIConfigList{})
}
