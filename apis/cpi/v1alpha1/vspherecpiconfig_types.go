// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VSphereCPIConfigSpec defines the desired state of VSphereCPIConfig
type VSphereCPIConfigSpec struct {
	VSphereCPI VSphereCPI `json:"vsphereCPI"`
}

type NSXTRouteConfig struct {
	// NSX-T T0/T1 logical router path
	// +kubebuilder:validation:Optional
	RouterPath *string `json:"routerPath,omitempty"`
}

type VMNetwork struct {
	// Internal VM network name
	// +kubebuilder:validation:Optional
	Internal *string `json:"internal,omitempty"`

	// External VM network name
	// +kubebuilder:validation:Optional
	External *string `json:"external,omitempty"`

	// Internal VM network CIDR to be excluded
	// +kubebuilder:validation:Optional
	ExcludeInternalSubnetCidr *string `json:"excludeInternalSubnetCidr,omitempty"`

	// External VM network CIDR to be excluded
	// +kubebuilder:validation:Optional
	ExcludeExternalSubnetCidr *string `json:"excludeExternalSubnetCidr,omitempty"`
}

type NSXT struct {
	// A flag that enables pod routing
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	PodRoutingEnabled *bool `json:"podRoutingEnabled,omitempty"`

	// Route configuration for NSXT
	// +kubebuilder:validation:Optional
	Route *NSXTRouteConfig `json:"route,omitempty"`

	// A secret reference that contains Nsx-T login credential to access NSX-T
	// consists of the field username and password
	// +kubebuilder:validation:Optional
	CredentialLocalObjRef *v1.TypedLocalObjectReference `json:"credentialLocalObjRef,omitempty"`

	// The NSX-T server
	// +kubebuilder:validation:Optional
	APIHost *string `json:"apiHost,omitempty"`

	// Insecure is to be set to true if NSX-T uses self-signed cert
	// +kubebuilder:validation:Optional
	Insecure *bool `json:"insecure,omitempty"`

	// RemoteAuth is to be set to true if NSX-T uses remote authentication (authentication done through the vIDM)
	// +kubebuilder:validation:Optional
	RemoteAuth *bool `json:"remoteAuth,omitempty"`

	// VMCAccessToken is VMC access token for token based authentication
	// +kubebuilder:validation:Optional
	VMCAccessToken *string `json:"vmcAccessToken,omitempty"`

	// VMCAuthHost is VMC verification host for token based authentication
	// +kubebuilder:validation:Optional
	VMCAuthHost *string `json:"vmcAuthHost,omitempty"`

	// Client certificate key for NSX-T
	// +kubebuilder:validation:Optional
	ClientCertKeyData *string `json:"clientCertKeyData,omitempty"`

	// Client certificate data for NSX-T
	// +kubebuilder:validation:Optional
	ClientCertData *string `json:"clientCertData,omitempty"`

	// The certificate authority for the server certificate for locally signed certificates
	// +kubebuilder:validation:Optional
	RootCAData *string `json:"rootCAData,omitempty"`
}

type Proxy struct {
	// HTTP proxy setting
	// +kubebuilder:validation:Optional
	HTTPProxy *string `json:"http_proxy,omitempty"`

	// HTTPS proxy setting
	// +kubebuilder:validation:Optional
	HTTPSProxy *string `json:"https_proxy,omitempty"`

	// No-proxy setting
	// +kubebuilder:validation:Optional
	NoProxy *string `json:"no_proxy,omitempty"`
}

type NonParavirtualConfig struct {

	// The IP address or FQDN of the vSphere endpoint
	// +kubebuilder:validation:Optional
	VCenterAPIEndpoint *string `json:"vCenterAPIEndpoint,omitempty"`

	// The datacenter in which VMs are created/located
	// +kubebuilder:validation:Optional
	Datacenter *string `json:"datacenter,omitempty"`

	// A secret reference that contains vSphere login credentials to access a vSphere endpoint
	// consists of the fields username and password
	// +kubebuilder:validation:Optional
	VSphereCredentialLocalObjRef *v1.TypedLocalObjectReference `json:"vSphereCredentialLocalObjRef,omitempty"`

	// The cryptographic thumbprint of the vSphere endpoint's certificate. Default value is "".
	// +kubebuilder:validation:Optional
	TLSThumbprint *string `json:"tlsThumbprint,omitempty"`

	// The region used by vSphere multi-AZ feature
	// +kubebuilder:validation:Optional
	Region *string `json:"region,omitempty"`

	// The zone used by vSphere multi-AZ feature
	// +kubebuilder:validation:Optional
	Zone *string `json:"zone,omitempty"`

	// The flag that disables TLS peer verification
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Insecure *bool `json:"insecure,omitempty"`

	// The IP family configuration
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="ipv4";"ipv6";"ipv4,ipv6";"ipv6,ipv4"
	IPFamily *string `json:"ipFamily,omitempty"`

	// +kubebuilder:validation:Optional
	VMNetwork *VMNetwork `json:"vmNetwork,omitempty"`

	// External arguments for cloud provider
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	TLSCipherSuites *string `json:"tlsCipherSuites,omitempty"`

	// +kubebuilder:validation:Optional
	NSXT *NSXT `json:"nsxt,omitempty"`

	// +kubebuilder:validation:Optional
	Proxy *Proxy `json:"proxy,omitempty"`
}

type ParavirtualConfig struct {
	// A flag that enables pod routing by Antrea NSX for paravirtual mode
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	AntreaNSXPodRoutingEnabled *bool `json:"antreaNSXPodRoutingEnabled,omitempty"`
}

type VSphereCPI struct {
	// The vSphere mode. Either `vsphereCPI` or `vsphereParavirtualCPI`.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=vsphereCPI;vsphereParavirtualCPI
	Mode *string `json:"mode,omitempty"`

	*NonParavirtualConfig `json:",omitempty"`

	*ParavirtualConfig `json:",omitempty"`
}

// VSphereCPIConfigStatus defines the observed state of VSphereCPIConfig
type VSphereCPIConfigStatus struct {
	// Name of the data value secret created by vSphere CPI controller
	//+ kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=vspherecpiconfigs,shortName=vcpic,scope=Namespaced
//+kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.namespace",description="The name of the VSphereCPIConfig"
//+kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.vsphereCPI.mode",description="Name of the kapp-controller data values secret"

// VSphereCPIConfig is the Schema for the VSphereCPIConfig API
type VSphereCPIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereCPIConfigSpec   `json:"spec,omitempty"`
	Status VSphereCPIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VSphereCPIConfigList contains a list of VSphereCPIConfig
type VSphereCPIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereCPIConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSphereCPIConfig{}, &VSphereCPIConfigList{})
}
