// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AntreaConfigSpec defines the desired state of AntreaConfig
type AntreaConfigSpec struct {
	// The cloud provider in use. One of the following options => aws, azure, vsphere, docker
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum="aws";"azure";"vsphere";"docker"
	// +kubebuilder:default:=vsphere
	InfraProvider string `json:"infraProvider"`

	Antrea Antrea `json:"antrea,omitempty"`
}

type Antrea struct {
	Config Config `json:"config,omitempty"`
}

type Config struct {
	// The traffic encapsulation mode. One of the following options => encap, noEncap, hybrid, networkPolicyOnly
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="encap";"noEncap";"hybrid";"networkPolicyOnly"
	// +kubebuilder:default:=encap
	TrafficEncapMode string `json:"trafficEncapMode,omitempty"`

	// Flag to enable/disable SNAT for the egress traffic from a Pod to the external network
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	NoSNAT bool `json:"noSNAT,omitempty"`

	// Disable UDP tunnel offload feature on default NIC
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	DisableUdpTunnelOffload bool `json:"disableUdpTunnelOffload,omitempty"`

	// Default MTU to use for the host gateway interface and the network interface of each Pod. If omitted, antrea-agent will discover the MTU of the Node's primary interface
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=""
	DefaultMTU string `json:"defaultMTU,omitempty"`

	// List of allowed cipher suites. If omitted, the default Go Cipher Suites will be used
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384"
	TlsCipherSuites string `json:"tlsCipherSuites,omitempty"`

	// FeatureGates is a map of feature names to flags that enable or disable experimental features
	FeatureGates AntreaFeatureGates `json:"featureGates,omitempty"`
}

type AntreaFeatureGates struct {
	// Flag to enable/disable antrea proxy
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	AntreaProxy bool `json:"AntreaProxy,omitempty"`

	// Flag to enable/disable EndpointSlice support in AntreaProxy. If AntreaProxy is not enabled, this flag will not take effect
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	EndpointSlice bool `json:"EndpointSlice,omitempty"`

	// Flag to enable/disable antrea policy
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	AntreaPolicy bool `json:"AntreaPolicy,omitempty"`

	// Flag to enable/disable flow exporter
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	FlowExporter bool `json:"FlowExporter,omitempty"`

	// Flag to enable/disable SNAT IPs of Pod egress traffic
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Egress bool `json:"Egress,omitempty"`

	// Flag to enable/disable NodePortLocal feature to make the pods reachable externally through NodePort
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	NodePortLocal bool `json:"NodePortLocal,omitempty"`

	// Flag to enable/disable antrea traceflow
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	AntreaTraceflow bool `json:"AntreaTraceflow,omitempty"`

	// Flag to enable/disable network policy stats
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	NetworkPolicyStats bool `json:"NetworkPolicyStats,omitempty"`
}

// AntreaConfigStatus defines the observed state of AntreaConfig
type AntreaConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AntreaConfig is the Schema for the antreaconfigs API
type AntreaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AntreaConfigSpec   `json:"spec,omitempty"`
	Status AntreaConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AntreaConfigList contains a list of AntreaConfig
type AntreaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AntreaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AntreaConfig{}, &AntreaConfigList{})
}
