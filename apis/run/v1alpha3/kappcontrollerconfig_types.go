// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KappControllerConfigSpec defines the desired state of KappControllerConfig
type KappControllerConfigSpec struct {
	// The namespace in which kapp-controller is deployed
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=tkg-system
	Namespace string `json:"namespace,omitempty"`

	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={deployment:{hostNetwork:true}}
	KappController KappController `json:"kappController,omitempty"` // object default needs at least one param so that CRD generation is not null, issue https://github.com/kubernetes-sigs/controller-tools/issues/631
}

type KappController struct {
	// Whether to create namespace specified for kapp-controller
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	CreateNamespace bool `json:"createNamespace,omitempty"`

	// The namespace value used for global packaging resources. Any Package and PackageMetadata CRs within that namespace will be included in all other namespaces on the cluster, without duplicating them
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=tkg-system
	GlobalNamespace string `json:"globalNamespace,omitempty"`

	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={hostNetwork:true}
	Deployment KappDeployment `json:"deployment,omitempty"` // object default needs at least one param so that CRD generation is not null, issue https://github.com/kubernetes-sigs/controller-tools/issues/631

	Config KappConfig `json:"config,omitempty"`
}

type KappDeployment struct {
	// Whether to enable host networking for kapp-controller deployment
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// The priority value that various system components use to find the priority of the kapp-controller pod
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=system-cluster-critical
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// Concurrency of kapp-controller deployment
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=4
	Concurrency int `json:"concurrency,omitempty"`

	// kapp-controller deployment tolerations
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={{key: CriticalAddonsOnly, operator: Exists}, {effect: NoSchedule, key: node-role.kubernetes.io/master}, {effect: NoSchedule, key: node.kubernetes.io/not-ready}, {effect: NoSchedule, key: node.cloudprovider.kubernetes.io/uninitialized, value: "true"}}
	Tolerations []map[string]string `json:"tolerations,omitempty"`

	// Bind port for kapp-controller API
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=10100
	APIPort int `json:"apiPort,omitempty"`

	// Address for metrics server
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="0"
	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`
}

type KappConfig struct {
	// A cert chain of trusted CA certs. These will be added to the system-wide cert pool of trusted CA's. Cluster-wide CA Certificate setting will be used if this is not provided.
	//+kubebuilder:validation:Optional
	CaCerts string `json:"caCerts,omitempty"`

	// The url/ip of a proxy for kapp controller to use when making network requests. Cluster-wide HTTP proxy setting will be used if this is not provided.
	//+kubebuilder:validation:Optional
	HTTPProxy string `json:"httpProxy,omitempty"`

	// The url/ip of a TLS capable proxy for kapp-controller to use when making network requests. Cluster-wide HTTPS proxy setting will be used if this is not provided.
	//+kubebuilder:validation:Optional
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// A comma delimited list of domain names which kapp-controller should bypass the proxy for when making requests. Cluster-wide no-proxy setting will be used if this is not provided.
	//+kubebuilder:validation:Optional
	NoProxy string `json:"noProxy,omitempty"`

	// A comma delimited list of hostnames for which kapp-controller should skip TLS verification
	//+kubebuilder:validation:Optional
	DangerousSkipTLSVerify string `json:"dangerousSkipTLSVerify,omitempty"`
}

// KappControllerConfigStatus defines the observed state of KappControllerConfig
type KappControllerConfigStatus struct {
	// Name of the data value secret created by controller
	//+kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=kappcontrollerconfigs,shortName=kappconf,scope=Namespaced
//+kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.namespace",description="The namespace in which kapp-controller is deployed"
//+kubebuilder:printcolumn:name="GlobalNamespace",type="string",JSONPath=".spec.kappController.globalNamespace",description="The namespace value used for global packaging resources. Any Package and PackageMetadata CRs within that namespace will be included in all other namespaces on the cluster, without duplicating them"
//+kubebuilder:printcolumn:name="SecretName",type="string",JSONPath=".status.secretName",description="Name of the kapp-controller data values secret"

// KappControllerConfig is the Schema for the kappcontrollerconfigs API
type KappControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KappControllerConfigSpec   `json:"spec"`
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
