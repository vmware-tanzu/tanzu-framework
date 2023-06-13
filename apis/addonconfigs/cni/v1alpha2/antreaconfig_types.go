// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AntreaProxyNodePortAddress []string

// AntreaConfigSpec defines the desired state of AntreaConfig
type AntreaConfigSpec struct {
	Antrea Antrea `json:"antrea,omitempty"`
	// AntreaNsx defines nsxt adapter related configurations
	AntreaNsx AntreaNsx `json:"antreaNsx,omitempty"`
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
	ProxyAll bool `json:"proxyAll,omitempty"`

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

type AntreaMultiCluster struct {
	//+ kubebuilder:validation:Optional
	Enable bool `json:"enable,omitempty"`
	//+ kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

type AntreaMulticast struct {
	//+ kubebuilder:validation:Optional
	IGMPQueryInterval string `json:"igmpQueryInterval,omitempty"`
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

	// Provide the address of Kubernetes apiserver, to override any value provided in kubeconfig or InClusterConfig.
	// +kubebuilder:validation:Optional
	KubeAPIServerOverride string `json:"kubeAPIServerOverride,omitempty"`

	// Multicast related configuration.
	// +kubebuilder:validation:Optional
	Multicast AntreaMulticast `json:"multicast,omitempty"`

	// MultiCluster realted configuration.
	// +kubebuilder:validation:Optional
	MultiCluster AntreaMultiCluster `json:"multicluster,omitempty"`

	// The name of the interface on Node which is used for tunneling or routing.
	// +kubebuilder:validation:Optional
	TransportInterface string `json:"transportInterface,omitempty"`

	// The network CIDRs of the interface on Node which is used for tunneling or routing.
	// +kubebuilder:validation:Optional
	TransportInterfaceCIDRs []string `json:"transportInterfaceCIDRs,omitempty"`

	// The names of the interfaces on Nodes that are used to forward multicast traffic.
	// +kubebuilder:validation:Optional
	MulticastInterfaces []string `json:"multicastInterfaces,omitempty"`

	// Tunnel protocols used for encapsulating traffic across Nodes. One of the following options =:> geneve, vxlan, gre, stt
	// +kubebuilder:validation:Optional
	TunnelType string `json:"tunnelType,omitempty"`

	// Determines how tunnel traffic is encrypted. One of the following options =:> none, ipsec, wireguard
	// +kubebuilder:validation:Optional
	TrafficEncryptionMode string `json:"trafficEncryptionMode,omitempty"`

	// Enable usage reporting (telemetry) to VMware.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	EnableUsageReporting bool `json:"enableUsageReporting,omitempty"`

	// Specifies WireGuard related configuration.
	// +kubebuilder:validation:Optional
	WireGuard AntreaWireGuard `json:"wireGuard,omitempty"`

	// ClusterIP CIDR range for Services.
	// +kubebuilder:validation:Optional
	ServiceCIDR string `json:"serviceCIDR,omitempty"`

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

	// Enable bridging mode of Pod network on Nodes
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	EnableBridgingMode bool `json:"enableBridgingMode,omitempty"`

	// Disable TX checksum offloading for container network interfaces
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	DisableTXChecksumOffload bool `json:"disableTXChecksumOffload,omitempty"`

	// Provide the address of DNS server, to override the kube-dns service
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=""
	DNSServerOverride string `json:"dnsServerOverride,omitempty"`

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

	// Enable Antrea Multi-cluster Gateway to support cross-cluster traffic.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	MultiCluster bool `json:"Multicluster,omitempty"`

	// Enable support for provisioning secondary network interfaces for Pods (using Pod annotations).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	SecondaryNetwork bool `json:"SecondaryNetwork,omitempty"`

	// Enable mirroring or redirecting the traffic Pods send or receive.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	TrafficControl bool `json:"TrafficControl,omitempty"`
}

// AntreaConfigStatus defines the observed state of AntreaConfig
type AntreaConfigStatus struct {
	// Message to indicate failure reason
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
	// Reference to the data value secret created by controller
	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

type AntreaNsx struct {
	// Enable indicates whether nsxt adapter shall be enabled in the cluster
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Enable bool `json:"enable,omitempty"`
	// BootstrapFrom either providerRef or inline configs
	// +kubebuilder:validation:Optional
	BootstrapFrom AntreaNsxBootstrapFrom `json:"bootstrapFrom,omitempty"`
	// Config is  configuration for nsxt adapter
	// +kubebuilder:validation:Optional
	AntreaNsxConfig AntreaNsxConfig `json:"config,omitempty"`
}

type AntreaNsxBootstrapFrom struct {
	// ProviderRef is used with uTKG, which will be filled by uTKG Addon Controller
	// +kubebuilder:validation:Optional
	ProviderRef *AntreaNsxProvider `json:"providerRef,omitempty"`
	// Inline is used with TKGm, user need to fill in manually
	// +kubebuilder:validation:Optional
	Inline *AntreaNsxInline `json:"inline,omitempty"`
}

type AntreaNsxProvider struct {
	// Api version for nsxServiceAccount, its value is "nsx.vmware.com/v1alpha1" now
	// +kubebuilder:validation:Optional
	ApiGroup string `json:"apigroup,omitempty"`
	// Kind is the kind for crd, here its value is NsxServiceAccount
	// +kubebuilder:validation:Optional
	Kind string `json:"kind,omitempty"`
	// Name is the name for NsxServiceAccount
	// +kubebuilder:validation:Optional
	Name string `json:"name,omitempty"`
}

type AntreaNsxInline struct {
	// NsxManagers is the list for nsx managers, it can be either IP address or domain name
	// +kubebuilder:validation:Optional
	NsxManagers []string `json:"nsxManagers,omitempty"`
	// ClusterName is the name for the created cluster
	// +kubebuilder:validation:Optional
	ClusterName string `json:"clusterName,omitempty"`
	// NsxCertName is cert files to access nsx manager
	// +kubebuilder:validation:Optional
	NsxCertName string `json:"nsxCertName,omitempty"`
}

type AntreaNsxConfig struct {
	// InfraType is the type for infrastructure, so far it is vSphere, VMC, AWS, Azure
	// +kubebuilder:validation:Optional
	InfraType string `json:"infraType,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=antreaconfigs,shortName=antreaconf,scope=Namespaced
// +kubebuilder:printcolumn:name="TrafficEncapMode",type="string",JSONPath=".spec.antrea.config.trafficEncapMode",description="The traffic encapsulation mode. One of the following options => encap, noEncap, hybrid, networkPolicyOnly"
// +kubebuilder:printcolumn:name="DefaultMTU",type="string",JSONPath=".spec.antrea.config.defaultMTU",description="Default MTU to use for the host gateway interface and the network interface of each Pod. If omitted, antrea-agent will discover the MTU of the Node's primary interface"
// +kubebuilder:printcolumn:name="AntreaProxy",type="string",JSONPath=".spec.antrea.config.featureGates.AntreaProxy",description="Flag to enable/disable antrea proxy"
// +kubebuilder:printcolumn:name="AntreaPolicy",type="string",JSONPath=".spec.antrea.config.featureGates.AntreaPolicy",description="Flag to enable/disable antrea policy"
// +kubebuilder:printcolumn:name="SecretRef",type="string",JSONPath=".status.secretRef",description="Name of the antrea data values secret"
// +kubebuilder:storageversion
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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
