/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CSIConfigSpec defines the desired state of CSIConfig
// CSIConfigSpec defines the desired state of CSIConfig
type CSIConfigSpec struct {
	VSphereCSI VSphereCSI `json:"vsphereCSI"`
}

// CSIConfigStatus defines the observed state of CSIConfig
type CSIConfigStatus struct {
	// Name of the secret created by csi controller
	//+ kubebuilder:validation:Optional
	SecretRef *string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CSIConfig is the Schema for the csiconfigs API
type CSIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIConfigSpec   `json:"spec,omitempty"`
	Status CSIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CSIConfigList contains a list of CSIConfig
type CSIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIConfig `json:"items"`
}

type VSphereCSI struct {
	// The vSphere mode. Either `vsphereCSI` or `vsphereParavirtualCSI`.
	// +kubebuilder:validation:Enum=vsphereCSI;vsphereParavirtualCSI
	Mode string `json:"mode"`

	*ParavirtualConfig `json:"pvconfig,omitempty"`

	*NonParavirtualConfig `json:"config,omitempty"`
}

type ParavirtualConfig struct {

	// The name of the guest cluster using the csi components
	// +kubebuilder:validation:Required
	ClusterName string `json:"clusterName"`

	// The unique id of the guest cluster using the csi components
	// +kubebuilder:validation:Required
	ClusterUID string `json:"clusterUID"`

	// The namespace csi components are to be deployed in
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// The DNS name of the supervisor cluster endpoint
	// +kubebuilder:validation:Required
	SupervisorMasterEndpointHostname string `json:"supervisorMasterEndpointHostname"`

	// The IP port via which to communicate with the supervisor cluster
	// +kubebuilder:validation:Required
	SupervisorMasterPort int32 `json:"supervisorMasterPort"`
}

type NonParavirtualConfig struct {
	// The namespace csi components are to be deployed in
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:Required
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:Required
	Server string `json:"server"`

	// +kubebuilder:validation:Required
	Datacenter string `json:"datacenter"`

	// +kubebuilder:validation:Required
	PublicNetwork string `json:"publicNetwork"`

	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// +kubebuilder:validation:Required
	Password string `json:"password"`

	// +kubebuilder:validation:Optional
	Region *string `json:"region,omitempty"`

	// +kubebuilder:validation:Optional
	Zone *string `json:"zone,omitempty"`

	// +kubebuilder:validation:Optional
	UseTopologyCategories *bool `json:"useTopologyCategories,omitempty"`

	// +kubebuilder:validation:Optional
	ProvisionTimeout *string `json:"provisionTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	AttachTimeout *string `json:"attachTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	ResizerTimeout *string `json:"resizerTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	VSphereVersion *string `json:"vSphereVersion,omitempty"`

	// +kubebuilder:validation:Optional
	HttpProxy *string `json:"httpProxy,omitempty"`

	// +kubebuilder:validation:Optional
	HttpsProxy *string `json:"httpsProxy,omitempty"`

	// +kubebuilder:validation:Optional
	NoProxy *string `json:"noProxy,omitempty"`

	// +kubebuilder:validation:Optional
	DeploymentReplicas *int32 `json:"deploymentReplicas,omitempty"`

	// +kubebuilder:validation:Optional
	WindowsSupport *bool `json:"windowsSupport,omitempty"`
}

func init() {
	SchemeBuilder.Register(&CSIConfig{}, &CSIConfigList{})
}
