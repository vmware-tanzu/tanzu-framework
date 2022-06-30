// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AntreaProxyNodePortAddress []string

// AntreaConfigSpec defines the desired state of AntreaConfig
type AntreaConfigSpec struct {
	Antrea Antrea `json:"antrea,omitempty"`
}

type Antrea struct {
	AntreaConfigDataValue AntreaConfigDataValue `json:"config,omitempty"`
}

type AntreaEgress struct {
	//+ kubebuilder:validation:Optional
	EgressExceptCIDRs []string `json:"exceptCIDRs,omitempty"`
}

type AntreaNodePortLocal struct {
	//+ kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`

	//+ kubebuilder:validation:Optional
	PortRange string `json:"portRange,omitempty"`
}

type AntreaProxy struct {
	//+ kubebuilder:validation:Optional
	ProxyAll bool `json:"enabled,omitempty"`

	//+ kubebuilder:validation:Optional
	NodePortAddresses []string `json:"nodePortAddresses,omitempty"`

	//+ kubebuilder:validation:Optional
	SkipServices []string `json:"skipServices,omitempty"`

	//+ kubebuilder:validation:Optional
	ProxyLoadBalancerIPs bool `json:"proxyLoadBalancerIPs,omitempty"`
}

type AntreaFlowExporter struct {
	//+ kubebuilder:validation:Optional
	CollectorAddress string `json:"collectorAddress,omitempty"`

	//+ kubebuilder:validation:Optional
	PollInterval string `json:"pollInterval,omitempty"`

	//+ kubebuilder:validation:Optional
	ActiveFlowTimeout string `json:"activeFlowTimeout,omitempty"`

	//+ kubebuilder:validation:Optional
	IdleFlowTimeout string `json:"idleFlowTimeout,omitempty"`
}

type AntreaWireGuard struct {
	//+ kubebuilder:validation:Optional
	Port int `json:"port,omitempty"`
}

type AntreaConfigDataValue struct {
	// Specifies Egress related configuration.
	// +kubebuilder:validation:Optional
	Egress AntreaEgress `json:"egress,omitempty"`

	// Specifies NodePortLocal related configuration.
	// +kubebuilder:validation:Optional
	NodePortLocal AntreaNodePortLocal `json:"nodePortLocal,omitempty"`

	// Specifies AntreaProxy related configuration.
	// +kubebuilder:validation:Optional
	AntreaProxy AntreaProxy `json:"antreaProxy,omitempty"`

	// Specifies FlowExporter related configuration.
	// +kubebuilder:validation:Optional
	AntreaFlowExporter AntreaFlowExporter `json:"flowExporter,omitempty"`

	// Specifies WireGuard related configuration.
	// +kubebuilder:validation:Optional
	WireGuard AntreaWireGuard `json:"wireGuard,omitempty"`

	// The name of the interface on Node which is used for tunneling or routing.
	// +kubebuilder:validation:Optional
	TransportInterface string `json:"transportInterface,omitempty"`

	// The network CIDRs of the interface on Node which is used for tunneling or routing.
	// +kubebuilder:validation:Optional
	TransportInterfaceCIDRs []string `json:"transportInterfaceCIDRs,omitempty"`

	// The names of the interfaces on Nodes that are used to forward multicast traffic.
	// +kubebuilder:validation:Optional
	MulticastInterface string `json:"multicastInterface,omitempty"`

	// The traffic encapsulation mode. One of the following options => encap, noEncap, hybrid, networkPolicyOnly
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="encap";"noEncap";"hybrid";"networkPolicyOnly"
	// +kubebuilder:default:=encap
	KubeAPIServerOverride string `json:"kubeAPIServerOverride,omitempty"`

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
	DisableUDPTunnelOffload bool `json:"disableUdpTunnelOffload,omitempty"`

	// Default MTU to use for the host gateway interface and the network interface of each Pod. If omitted, antrea-agent will discover the MTU of the Node's primary interface
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=""
	DefaultMTU string `json:"defaultMTU,omitempty"`

	// List of allowed cipher suites. If omitted, the default Go Cipher Suites will be used
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384"
	TLSCipherSuites string `json:"tlsCipherSuites,omitempty"`

	// FeatureGates is a map of feature names to flags that enable or disable experimental features
	// +kubebuilder:validation:Optional
	FeatureGates AntreaFeatureGates `json:"featureGates,omitempty"`
}

type AntreaFeatureGates struct {
	// Flag to enable/disable antrea proxy
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
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
	// +kubebuilder:default:=true
	Egress bool `json:"Egress,omitempty"`

	// Flag to enable/disable NodePortLocal feature to make the pods reachable externally through NodePort
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	NodePortLocal bool `json:"NodePortLocal,omitempty"`

	// Flag to enable/disable antrea traceflow
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	AntreaTraceflow bool `json:"AntreaTraceflow,omitempty"`

	// Flag to enable/disable network policy stats
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	NetworkPolicyStats bool `json:"NetworkPolicyStats,omitempty"`

	// Flag to enable/disable antrea IPAM
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	AntreaIPAM bool `json:"AntreaIPAM,omitempty"`

	// Flag to enable/disable service external IP
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	ServiceExternalIP bool `json:"ServiceExternalIP,omitempty"`

	// Flag to enable/disable multicast
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Multicast bool `json:"Multicast,omitempty"`
}

// AntreaConfigStatus defines the observed state of AntreaConfig
type AntreaConfigStatus struct {
	// Reference to the data value secret created by controller
	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=antreaconfigs,shortName=antreaconf,scope=Namespaced
// +kubebuilder:printcolumn:name="TrafficEncapMode",type="string",JSONPath=".spec.antrea.config.trafficEncapMode",description="The traffic encapsulation mode. One of the following options => encap, noEncap, hybrid, networkPolicyOnly"
// +kubebuilder:printcolumn:name="DefaultMTU",type="string",JSONPath=".spec.antrea.config.defaultMTU",description="Default MTU to use for the host gateway interface and the network interface of each Pod. If omitted, antrea-agent will discover the MTU of the Node's primary interface"
// +kubebuilder:printcolumn:name="AntreaProxy",type="string",JSONPath=".spec.antrea.config.featureGates.AntreaProxy",description="Flag to enable/disable antrea proxy"
// +kubebuilder:printcolumn:name="AntreaPolicy",type="string",JSONPath=".spec.antrea.config.featureGates.AntreaPolicy",description="Flag to enable/disable antrea policy"
// +kubebuilder:printcolumn:name="SecretRef",type="string",JSONPath=".status.secretRef",description="Name of the antrea data values secret"

// AntreaConfig is the Schema for the antreaconfigs API
type AntreaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AntreaConfigSpec   `json:"spec"`
	Status AntreaConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AntreaConfigList contains a list of AntreaConfig
type AntreaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AntreaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AntreaConfig{}, &AntreaConfigList{})
}
