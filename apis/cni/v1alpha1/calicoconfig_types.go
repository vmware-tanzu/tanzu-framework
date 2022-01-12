// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"net"
	"strings"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CalicoConfigSpec defines the desired state of CalicoConfig
type CalicoConfigSpec struct {

	// The namespace in which calico is deployed
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=kube-system
	Namespace string `json:"namespace,omitempty"`

	// Infrastructure provider in use
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum="aws";"azure";"vsphere";"docker"
	InfraProvider string `json:"infraProvider"`

	// The IP family calico should be configured with
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="ipv4";"ipv6";"ipv4,ipv6";"ipv6,ipv4"
	// +kubebuilder:default:=ipv4
	IPFamily string `json:"ipFamily,omitempty"`

	// The CIDR pool used to assign IP addresses to the pods in the cluster
	// +kubebuilder:validation:Optional
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// ClusterCIDR TypeCIDR `json:"clusterCIDR,omitempty"`

	// Maximum transmission unit setting
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=0
	VethMTU int64 `json:"vethMTU,omitempty"`
}

// CalicoConfigStatus defines the observed state of CalicoConfig
type CalicoConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CalicoConfig is the Schema for the calicoconfigs API
type CalicoConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CalicoConfigSpec   `json:"spec,omitempty"`
	Status CalicoConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CalicoConfigList contains a list of CalicoConfig
type CalicoConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CalicoConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CalicoConfig{}, &CalicoConfigList{})
}

type TypeCIDR struct {
	CIDR string `json:"clusterCIDR,omitempty"`
}

// if none-empty, validates all the comma-concatenated IP pools to be valid ipv4 or ipv6 subnets
func (s TypeCIDR) validateClusterCIDRPool() error { //nolint
	if s.CIDR == "" {
		return nil
	}
	ipAddresses := strings.Split(s.CIDR, ",")
	for _, addr := range ipAddresses {
		if err := isValidCIDRPool(addr); err != nil {
			return err
		}
	}

	return nil
}

func isValidCIDRPool(addr string) error { //nolint
	var (
		ipAddr  net.IP
		numBits int
	)

	// parses the input IP address into IP address and the network implied by the IP
	ip, subnet, err := net.ParseCIDR(addr)
	if err != nil {
		return err
	}

	// tries to convert the IP address to a 4-byte representation. It returns nil if not possible
	if ipAddr = ip.To4(); ipAddr != nil {
		numBits = 32
	} else {
		// tries to convert the IP address to a 16-byte representation. It returns nil if not possible
		if ipAddr = ip.To16(); ipAddr != nil {
			numBits = 128
		} else {
			return errors.New("invalid IP address")
		}
	}

	// gets the number of leading ones in the network mask
	leadingOnes, _ := subnet.Mask.Size()
	if leadingOnes == 0 {
		return errors.New("invalid network mask")
	}

	// returns an IPMask consisting of 'leadingOnes' 1 bits followed by 0s up to a total length of 'numBits' bits
	ipMask := net.CIDRMask(leadingOnes, numBits)

	// checks if the IP address and ipMask match to determine the validity of the provided IP pool
	for i := 0; i < len(ipAddr); i++ {
		if ipAddr[i]&(^ipMask[i]) != 0 {
			return errors.New("invalid IP pool")
		}
	}

	return nil
}
