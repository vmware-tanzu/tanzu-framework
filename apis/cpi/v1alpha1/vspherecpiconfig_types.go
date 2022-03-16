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

type NSXTRoute struct {
	// NSX-T T0/T1 logical router path
	// +kubebuilder:validation:Optional
	RouterPath string `json:"routerPath"`

	// Cluster CIDR
	// +kubebuilder:validation:Optional
	ClusterCidr string `json:"clusterCidr"`
}

type NSXT struct {
	// A flag that enables pod routing
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	PodRoutingEnabled bool `json:"podRoutingEnabled"`

	// Route configuration for NSXT
	// +kubebuilder:validation:Optional
	Routes *NSXTRoute `json:"routes"`

	// A secret reference that contains Nsx-T login credentials to access NSX-T
	// consists of the field username and password
	// +kubebuilder:validation:Optional
	NSXTCredentialsRef *v1.SecretReference `json:"nsxtCredentialsRef"`

	// The NSX-T server
	// +kubebuilder:validation:Optional
	Host string `json:"host"`

	// InsecureFlag is to be set to true if NSX-T uses self-signed cert
	// +kubebuilder:validation:Optional
	InsecureFlag bool `json:"insecureFlag"`

	// RemoteAuth is to be set to true if NSX-T uses remote authentication (authentication done through the vIDM)
	// +kubebuilder:validation:Optional
	RemoteAuth bool `json:"remoteAuth"`

	// VMCAccessToken is VMC access token for token based authentication
	// +kubebuilder:validation:Optional
	VMCAccessToken string `json:"vmcAccessToken"`

	// VMCAuthHost is VMC verification host for token based authentication
	// +kubebuilder:validation:Optional
	VMCAuthHost string `json:"vmcAuthHost"`

	// Client certificate key for NSX-T
	// +kubebuilder:validation:Optional
	ClientCertKeyData string `json:"clientCertKeyData"`

	// Client certificate data for NSX-T
	// +kubebuilder:validation:Optional
	ClientCertData string `json:"clientCertData"`

	// The certificate authority for the server certificate for locally signed certificates
	// +kubebuilder:validation:Optional
	RootCAData string `json:"rootCAData"`

	// The name of secret that stores CPI configuration
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName"`

	// The namespace of secret that stores CPI configuration
	// +kubebuilder:validation:Optional
	SecretNamespace string `json:"secretNamespace"`
}

type NonParavirtualConfig struct {

	// The IP address or FQDN of the vSphere endpoint
	// +kubebuilder:validation:Optional
	Server string `json:"server"`

	// The datacenter in which VMs are created/located
	// +kubebuilder:validation:Optional
	Datacenter string `json:"datacenter"`

	// A secret reference that contains vSphere login credentials to access a vSphere endpoint
	// consists of the fields username and password
	// +kubebuilder:validation:Optional
	VSphereCredentialRef *v1.SecretReference `json:"vSphereCredentialRef"`

	// The cryptographic thumbprint of the vSphere endpoint's certificate. Default value is "".
	// +kubebuilder:validation:Optional
	TLSThumbprint string `json:"tlsThumbprint"`

	// The region used by vSphere multi-AZ feature
	// +kubebuilder:validation:Optional
	Region string `json:"region"`

	// The zone used by vSphere multi-AZ feature
	// +kubebuilder:validation:Optional
	Zone string `json:"zone"`

	// The flag that disables TLS peer verification
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	InsecureFlag bool `json:"insecureFlag"`

	// The IP family configuration
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="ipv4";"ipv6";"ipv4,ipv6";"ipv6,ipv4"
	IPFamily string `json:"ipFamily"`

	// Internal VM network name
	// +kubebuilder:validation:Optional
	VMInternalNetwork string `json:"vmInternalNetwork"`

	// External VM network name
	// +kubebuilder:validation:Optional
	VMExternalNetwork string `json:"vmExternalNetwork"`

	// Internal VM network CIDR to be excluded
	// +kubebuilder:validation:Optional
	VMExcludeInternalNetworkSubnetCidr string `json:"vmExcludeInternalNetworkSubnetCidr"`

	// External VM network CIDR to be excluded
	// +kubebuilder:validation:Optional
	VMExcludeExternalNetworkSubnetCidr string `json:"vmExcludeExternalNetworkSubnetCidr"`

	// External arguments for cloud provider
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	TLSCipherSuites string `json:"tlsCipherSuites"`

	// +kubebuilder:validation:Optional
	NSXT *NSXT `json:"nsxt"`

	// HTTP proxy setting
	// +kubebuilder:validation:Optional
	HTTPProxy string `json:"http_proxy"`

	// HTTPS proxy setting
	// +kubebuilder:validation:Optional
	HTTPSProxy string `json:"https_proxy"`

	// No-proxy setting
	// +kubebuilder:validation:Optional
	NoProxy string `json:"no_proxy"`
}

type ParavirtualConfig struct {
	// Used in vsphereParavirtual mode, defines the Cluster API versions.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=cluster.x-k8s.io/v1beta1
	ClusterAPIVersion string `json:"clusterAPIVersion"`

	// Used in vsphereParavirtual mode, defines the Cluster kind.
	// +kubebuilder:validation:Optional
	ClusterKind string `json:"clusterKind"`

	// Used in vsphereParavirtual mode, defines the Cluster name.
	// +kubebuilder:validation:Optional
	ClusterName string `json:"clusterName"`

	// Used in vsphereParavirtual mode, defines the Cluster UID.
	// +kubebuilder:validation:Optional
	ClusterUID string `json:"clusterUID"`

	// Used in vsphereParavirtual mode, the endpoint IP of supervisor cluster's API server.
	// +kubebuilder:validation:Optional
	SupervisorMasterEndpointIP string `json:"supervisorMasterEndpointIP"`

	// Used in vsphereParavirtual mode, the endpoint port of supervisor cluster's API server port.
	// +kubebuilder:validation:Optional
	SupervisorMasterPort string `json:"supervisorMasterPort"`
}

type VSphereCPI struct {
	// The vSphere mode. Either `vsphereCPI` or `vsphereParavirtualCPI`.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=vsphereCPI;vsphereParavirtualCPI
	Mode string `json:"mode"`

	*NonParavirtualConfig `json:""`

	*ParavirtualConfig `json:""`
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
