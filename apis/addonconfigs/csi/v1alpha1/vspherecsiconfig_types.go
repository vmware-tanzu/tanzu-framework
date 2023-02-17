// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VSphereCSIConfigSpec defines the desired state of VSphereCSIConfig
type VSphereCSIConfigSpec struct {
	VSphereCSI VSphereCSI `json:"vsphereCSI"`
}

// VSphereCSIConfigStatus defines the observed state of VSphereCSIConfig
type VSphereCSIConfigStatus struct {
	// Name of the secret created by csi controller
	//+ kubebuilder:validation:Optional
	SecretRef *string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VSphereCSIConfig is the Schema for the vspherecsiconfigs API
type VSphereCSIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereCSIConfigSpec   `json:"spec,omitempty"`
	Status VSphereCSIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VSphereCSIConfigList contains a list of VSphereCSIConfig
type VSphereCSIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereCSIConfig `json:"items"`
}

type VSphereCSI struct {
	// The vSphere mode. Either `vsphereCSI` or `vsphereParavirtualCSI`.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=vsphereCSI;vsphereParavirtualCSI
	Mode string `json:"mode"`

	*NonParavirtualConfig `json:"config,omitempty"`
}

type NonParavirtualConfig struct {
	// +kubebuilder:validation:Optional
	TLSThumbprint string `json:"tlsThumbprint,omitempty"`

	// The namespace csi components are to be deployed in
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Optional
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:Optional
	Server string `json:"server"`

	// +kubebuilder:validation:Optional
	Datacenter string `json:"datacenter"`

	// +kubebuilder:validation:Optional
	PublicNetwork string `json:"publicNetwork"`

	// A secret reference that contains vSphere login credentials to access a vSphere endpoint
	// consists of the fields username and password
	// +kubebuilder:validation:Optional
	VSphereCredentialLocalObjRef *v1.TypedLocalObjectReference `json:"vSphereCredentialLocalObjRef"`

	// +kubebuilder:validation:Optional
	Region string `json:"region,omitempty"`

	// +kubebuilder:validation:Optional
	Zone string `json:"zone,omitempty"`

	// +kubebuilder:validation:Optional
	InsecureFlag *bool `json:"insecureFlag,omitempty"`

	// +kubebuilder:validation:Optional
	UseTopologyCategories *bool `json:"useTopologyCategories,omitempty"`

	// +kubebuilder:validation:Optional
	ProvisionTimeout string `json:"provisionTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	AttachTimeout string `json:"attachTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	ResizerTimeout string `json:"resizerTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	VSphereVersion string `json:"vSphereVersion,omitempty"`

	// +kubebuilder:validation:Optional
	HTTPProxy string `json:"httpProxy,omitempty"`

	// +kubebuilder:validation:Optional
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// +kubebuilder:validation:Optional
	NoProxy string `json:"noProxy,omitempty"`

	// +kubebuilder:validation:Optional
	DeploymentReplicas *int32 `json:"deploymentReplicas,omitempty"`

	// +kubebuilder:validation:Optional
	WindowsSupport *bool `json:"windowsSupport,omitempty"`

	// +kubebuilder:validation:Optional
	NetPermissions map[string]*NetPermissionConfig `json:"netPermissions,omitempty"`
}

// NetPermissionConfig consists of information used to restrict the
// network permissions set on file share volumes
type NetPermissionConfig struct {
	// Client IP address, IP range or IP subnet. Example: "10.20.30.0/24"; defaults to "*" if not specified
	Ips string `json:"ips,omitempty"`
	// Is it READ_ONLY, READ_WRITE or NO_ACCESS. Defaults to "READ_WRITE" if not specified
	Permissions string `json:"permissions,omitempty"`
	// Disallow root access for this IP range. Defaults to "false" if not specified
	RootSquash bool `json:"rootsquash,omitempty"`
}

func init() {
	SchemeBuilder.Register(&VSphereCSIConfig{}, &VSphereCSIConfigList{})
}
