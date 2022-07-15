// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OracleCPIConfigSpec defines the desired state of OracleCPIConfig
type OracleCPIConfigSpec struct {
	// Compartment configures the Oracle Cloud compartment within which the cluster resides.
	Compartment string `json:"compartment"`

	// VCN configures the Virtual Cloud Network (VCN) within which the cluster resides.
	VCN string `json:"vcn"`

	// LoadBalancer configures the load balancer provisioning for the Oracle CPI
	// +kubebuilder:validation:Optional
	LoadBalancer OracleLoadBalancer `json:"loadBalancer,omitempty"`

	// Proxy configures the proxy settings for the Oracle CPI
	// +kubebuilder:validation:Optional
	Proxy Proxy `json:"proxy,omitempty"`
}

type OracleLoadBalancer struct {
	// +kubebuilder:validation:Pattern:=^ocid1\.subnet\.oc[0-9]+\.[a-z0-9]*\.[a-z0-9]+$
	Subnet1 string `json:"subnet1,omitempty"`
	// +kubebuilder:validation:Pattern:=^ocid1\.subnet\.oc[0-9]+\.[a-z0-9]*\.[a-z0-9]+$
	Subnet2 string `json:"subnet2,omitempty"`

	// SecurityListManagementMode configures how security lists are managed by the CCM.
	// If you choose to have security lists managed by the CCM, ensure you have setup the following additional OCI policy:
	// Allow dynamic-group [your dynamic group name] to manage security-lists in compartment [your compartment name]
	// "All" (default): Manage all required security list rules for load balancer services.
	// "Frontend":  Manage only security list rules for ingress to the load balancer.
	// Requires that the user has setup a rule that allows inbound traffic to the appropriate ports for kube proxy health
	// port, node port ranges, and health check port ranges. E.g. 10.82.0.0/16 30000-32000.
	// "None": Disables all security list management. Requires that the user has setup a rule that allows inbound traffic to
	// the appropriate ports for kube proxy health port, node port ranges, and health check port ranges.
	// E.g. 10.82.0.0/16 30000-32000.
	// Additionally requires the user to mange rules to allow inbound traffic to load balancers.
	//
	// +kubebuilder:validation:Enum=All;Frontend;None
	// +kubebuilder:default:=All
	SecurityListManagementMode string `json:"securityListManagementMode"`

	// SecurityListSubnetMapping controls an optional specification of security lists to modify per subnet.
	// This does not apply if security list management is off.
	//
	// +kubebuilder:validation:Optional
	SecurityListSubnetMapping []SecurityListSubnetMapping `json:"securityListSubnetMapping,omitempty"`
}

type SecurityListSubnetMapping struct {
	// Subnet specifies the subnet to which to modify a security list for.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern:=^ocid1\.subnet\.oc[0-9]+\.[a-z0-9]*\.[a-z0-9]+$
	Subnet string `json:"subnet"`

	// SecurityList specifies the security list to modify for the subnet.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern:=^ocid1\.securitylist\.oc[0-9]+\.[a-z0-9]*\.[a-z0-9]+$
	SecurityList string `json:"securityList"`
}

// OracleCPIConfigStatus defines the observed state of OracleCPIConfig
type OracleCPIConfigStatus struct {
	// Name of the data value secret created by Oracle CPI controller
	//+ kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=oraclecpiconfigs,shortName=ocicpicfgs,scope=Namespaced
//+kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.namespace",description="The name of the oraclecpiconfigs"
//+kubebuilder:printcolumn:name="Secret",type="string",JSONPath=".status.secretRef",description="Name of the kapp-controller data values secret"

// OracleCPIConfig is the Schema for the VSphereCPIConfig API
type OracleCPIConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OracleCPIConfigSpec   `json:"spec,omitempty"`
	Status OracleCPIConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OracleCPIConfigList contains a list of OracleCPIConfig
type OracleCPIConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OracleCPIConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OracleCPIConfig{}, &OracleCPIConfigList{})
}
