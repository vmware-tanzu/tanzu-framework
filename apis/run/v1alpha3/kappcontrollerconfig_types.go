// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KappControllerConfigSpec defines the desired state of KappControllerConfig
type KappControllerConfigSpec struct {
	// The namespace in which calico is deployed
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=kube-system
	Namespace string `json:"namespace,omitempty"`

	KappController KappController `json:"kappController,omitempty"`
}

type KappController struct {
	// Whether to create namespace specified for kapp-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	CreateNamespace bool `json:"createNamespace,omitempty"`

	// The namespace value used for global packaging resources. Any Package and PackageMetadata CRs within that namespace will be included in all other namespaces on the cluster, without duplicating them
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=tanzu-package-repo-global
	GlobalNamespace string `json:"globalNamespace,omitempty"`

	Deployment KappDeployment `json:"deployment,omitempty"`

	Config KappConfig `json:"config,omitempty"`
}

type KappDeployment struct {
	// Whether to enable host networking for kapp-controller deployment
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// The priority value that various system components use to find the priority of the kapp-controller pod
	// +kubebuilder:validation:Optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// Concurrency of kapp-controller deployment
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=4
	Concurrency int `json:"concurrency,omitempty"`

	// kapp-controller deployment tolerations
	// +kubebuilder:validation:Optional
	Tolerations []map[string]string `json:"tolerations,omitempty"`

	// Bind port for kapp-controller API
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=10350
	APIPort int `json:"apiPort,omitempty"`

	// Address for metrics server
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=:8080
	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`
}

type KappConfig struct {
	// A cert chain of trusted CA certs. These will be added to the system-wide cert pool of trusted CA's
	// +kubebuilder:validation:Optional
	CaCerts string `json:"caCerts,omitempty"`

	// The url/ip of a proxy for kapp controller to use when making network requests
	// +kubebuilder:validation:Optional
	HTTPProxy string `json:"httpProxy,omitempty"`

	// The url/ip of a TLS capable proxy for kapp-controller to use when making network requests
	// +kubebuilder:validation:Optional
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// A comma delimited list of domain names which kapp-controller should bypass the proxy for when making requests
	// +kubebuilder:validation:Optional
	NoProxy string `json:"noProxy,omitempty"`

	// A comma delimited list of hostnames for which kapp-controller should skip TLS verification
	// +kubebuilder:validation:Optional
	DangerousSkipTLSVerify string `json:"dangerousSkipTLSVerify,omitempty"`
}

// KappControllerConfigStatus defines the observed state of KappControllerConfig
type KappControllerConfigStatus struct {
	// Name of the data value secret created by controller
	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KappControllerConfig is the Schema for the kappcontrollerconfigs API
type KappControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KappControllerConfigSpec   `json:"spec,omitempty"`
	Status KappControllerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KappControllerConfigList contains a list of KappControllerConfig
type KappControllerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KappControllerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KappControllerConfig{}, &KappControllerConfigList{})
}
