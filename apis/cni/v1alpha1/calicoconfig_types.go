// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CalicoConfigSpec defines the desired state of CalicoConfig
type CalicoConfigSpec struct {

	// The namespace in which calico is deployed
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=kube-system
	Namespace string `json:"namespace,omitempty"`

	// Infrastructure provider in use
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum="aws";"azure";"vsphere";"docker"
	InfraProvider string `json:"infraProvider"`

	// The IP family calico should be configured with
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="ipv4";"ipv6";"ipv4,ipv6";"ipv6,ipv4"
	// +kubebuilder:default:=ipv4
	IPFamily string `json:"ipFamily,omitempty"`

	Calico Calico `json:"calico,omitempty"`
}

type Calico struct {
	Config Config `json:"config,omitempty"`
}

type Config struct {
	// Maximum transmission unit setting
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=0
	// "0" as default means MTU will be auto detected
	VethMTU int64 `json:"vethMTU,omitempty"`
}

// CalicoConfigStatus defines the observed state of CalicoConfig
type CalicoConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CalicoConfig is the Schema for the calicoconfigs API
type CalicoConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CalicoConfigSpec   `json:"spec,omitempty"`
	Status CalicoConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CalicoConfigList contains a list of CalicoConfig
type CalicoConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CalicoConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CalicoConfig{}, &CalicoConfigList{})
}
