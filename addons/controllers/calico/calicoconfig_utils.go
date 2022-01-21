// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// TODO: remove after Shivaani's PR got merged
const (
	// InfrastructureRefDocker is the Docker infrastructure
	InfrastructureRefDocker = "DockerCluster"
)

// calicoConfigSpec defines the desired state of CalicoConfig
type calicoConfigSpec struct {
	Namespace     string `yaml:"namespace,omitempty"`
	InfraProvider string `yaml:"infraProvider"`
	IPFamily      string `yaml:"ipFamily,omitempty"`
	Calico        calico `yaml:"calico,omitempty"`
}

type calico struct {
	Config config `yaml:"config,omitempty"`
}

type config struct {
	VethMTU     int64  `yaml:"vethMTU,omitempty"`
	ClusterCIDR string `yaml:"clusterCIDR"`
}

func mapCalicoConfigSpec(cluster *clusterapiv1beta1.Cluster, config *cniv1alpha1.CalicoConfig) (*calicoConfigSpec, error) {
	var err error

	configSpec := &calicoConfigSpec{}
	configSpec.Namespace = config.Spec.Namespace
	configSpec.Calico.Config.VethMTU = config.Spec.Calico.Config.VethMTU

	// Derive InfraProvider from the cluster
	configSpec.InfraProvider, err = getInfraProvider(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get 'InfraProvider' setting for CalicoConfig")
	}

	// Derive IPFamily, ClusterCIDR from the cluster
	configSpec.IPFamily, configSpec.Calico.Config.ClusterCIDR, err = getCalicoNetworkSettings(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get 'clusterCIDR' and 'ipFamily' settings for CalicoConfig")
	}

	return configSpec, nil
}

func getCalicoNetworkSettings(cluster *clusterapiv1beta1.Cluster) (string, string, error) {
	clusterNetwork := cluster.Spec.ClusterNetwork
	if clusterNetwork == nil {
		return "", "", fmt.Errorf("cluster.Spec.ClusterNetwork is not set for cluster '%s'", cluster.Name)
	}

	if clusterNetwork.Pods == nil || len(clusterNetwork.Pods.CIDRBlocks) == 0 {
		return "", "", fmt.Errorf("cluster.Spec.ClusterNetwork.Pods is not set for cluster '%s'", cluster.Name)
	}

	var result string
	for _, cidr := range clusterNetwork.Pods.CIDRBlocks {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return "", "", fmt.Errorf("could not parse CIDR '%s': %s", cidr, err)
		}
		if ip.To4() != nil {
			result += "ipv4,"
		} else {
			if ip.To16() != nil {
				result += "ipv6,"
			} else {
				return "", "", fmt.Errorf("invalid IP address '%s' in cluster.Spec.ClusterNetwork.Pods.CIDRBlocks for cluster '%s'", ip.String(), cluster.Name)
			}
		}
	}

	cidrBlocks := strings.Join(clusterNetwork.Pods.CIDRBlocks, ",")
	return strings.TrimSuffix(result, ","), cidrBlocks, nil
}

// TODO: remove after Shivaani's PR got merged
func getInfraProvider(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var infraProvider string

	infrastructureRef := cluster.Spec.InfrastructureRef
	if infrastructureRef == nil {
		return "", fmt.Errorf("cluster.Spec.InfrastructureRef is not set for cluster '%s", cluster.Name)
	}

	switch infrastructureRef.Kind {
	case tkgconstants.InfrastructureRefVSphere:
		infraProvider = tkgconstants.InfrastructureProviderVSphere
	case tkgconstants.InfrastructureRefAWS:
		infraProvider = tkgconstants.InfrastructureProviderAWS
	case tkgconstants.InfrastructureRefAzure:
		infraProvider = tkgconstants.InfrastructureProviderAzure
	case InfrastructureRefDocker:
		infraProvider = tkgconstants.InfrastructureProviderDocker
	default:
		return "", fmt.Errorf("unknown cluster.Spec.InfrastructureRef.Kind is set for cluster '%s", cluster.Name)
	}

	return infraProvider, nil
}
