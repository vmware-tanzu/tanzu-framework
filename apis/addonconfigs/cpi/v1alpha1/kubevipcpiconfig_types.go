// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubevipCPIConfigSpec defines the desired state of KubevipCPIConfig
type KubevipCPIConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// loadbalancerCIDRs is a list of comma separated cidrs will
	// be used to allocate IP for external load balancer.
	// For example 192.168.0.200/29,192.168.1.200/29
	//+ kubebuilder:validation:Optional
	LoadbalancerCIDRs *string `json:"loadbalancerCIDRs,omitempty"`

	// loadbalancerIPRanges is a list of comma separated IP ranges will
	// be used to allocate IP for external load balancer.
	// For example 192.168.0.10-192.168.0.11,192.168.0.10-192.168.0.13
	//+ kubebuilder:validation:Optional
	LoadbalancerIPRanges *string `json:"loadbalancerIPRanges,omitempty"`
}

// KubevipCPIConfigStatus defines the observed state of KubevipCPIConfig
type KubevipCPIConfigStatus struct {
	// Name of the secret created by kubevip cloudprovider config controller
	//+ kubebuilder:validation:Optional
	SecretRef *string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubevipCPIConfig is the Schema for the kubevipcpiconfigs API
type KubevipCPIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubevipCPIConfigSpec   `json:"spec,omitempty"`
	Status KubevipCPIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubevipCPIConfigList contains a list of KubevipCPIConfig
type KubevipCPIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubevipCPIConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubevipCPIConfig{}, &KubevipCPIConfigList{})
}
