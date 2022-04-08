// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TkgServiceConfigurationSpec defines the desired state of TkgServiceConfiguration
type TkgServiceConfigurationSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Default CNI for TanzuKubernetesCluster
	DefaultCNI string `json:"defaultCNI,omitempty"`

	// Proxy specifies default global HTTP(s) Proxy Configuration for all new TanzuKubernetesClusters in this Supervisor cluster
	//
	// If omitted, no proxy will be configured for new TanzuKubernetesClusters
	// +optional
	Proxy *ProxyConfiguration `json:"proxy,omitempty"`

	// Trust specifies default global Trust settings for all new TanzuKubernetesClusters
	// in the Supervisor Cluster.
	//
	// If omitted, no additional Trust settings will be configured for the new TanzuKubernetesCluster.
	//
	// +optional
	Trust *TrustConfiguration `json:"trust,omitempty"`

	// DefaultNodeDrainTimeout specifies the total amount of time that the
	// controller will spend on draining a node by default. Undefined, the value
	// is 0, meaning that the node can be drained without any time limitations.
	// NOTE: NodeDrainTimeout is different from `kubectl drain --timeout`
	// +optional
	DefaultNodeDrainTimeout *metav1.Duration `json:"defaultNodeDrainTimeout,omitempty"`
}

// TkgServiceConfigurationStatus defines the observed state of TkgServiceConfiguration
type TkgServiceConfigurationStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tkgserviceconfigurations,scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Default CNI",type=string,JSONPath=.spec.defaultCNI

// TkgServiceConfiguration is the Schema for the tkgserviceconfigurations API
type TkgServiceConfiguration struct { // nolint:maligned
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TkgServiceConfigurationSpec   `json:"spec,omitempty"`
	Status TkgServiceConfigurationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TkgServiceConfigurationList contains a list of TkgServiceConfiguration
type TkgServiceConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TkgServiceConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TkgServiceConfiguration{}, &TkgServiceConfigurationList{})
}
