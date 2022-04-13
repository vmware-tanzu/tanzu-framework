// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

// CalicoConfigSpec defines the desired state of CalicoConfig
type CalicoConfigSpec struct {
	Namespace     string `yaml:"namespace,omitempty"`
	InfraProvider string `yaml:"infraProvider"`
	IPFamily      string `yaml:"ipFamily,omitempty"`
	Calico        calico `yaml:"calico,omitempty"`
}

type calico struct {
	Config config `yaml:"config,omitempty"`
}

type config struct {
	VethMTU     string `yaml:"vethMTU,omitempty"`
	ClusterCIDR string `yaml:"clusterCIDR"`
}

func MapCalicoConfigSpec(cluster *clusterapiv1beta1.Cluster, config *cniv1alpha1.CalicoConfig) (*CalicoConfigSpec, error) {
	var err error

	configSpec := &CalicoConfigSpec{}
	configSpec.Namespace = config.Spec.Namespace
	configSpec.Calico.Config.VethMTU = strconv.FormatInt(config.Spec.Calico.Config.VethMTU, 10)

	// Derive InfraProvider from the cluster
	configSpec.InfraProvider, err = GetInfraProvider(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get 'InfraProvider' setting for CalicoConfig")
	}

	// Derive IPFamily, ClusterCIDR from the cluster
	configSpec.IPFamily, configSpec.Calico.Config.ClusterCIDR, err = GetCalicoNetworkSettings(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get 'clusterCIDR' and 'ipFamily' settings for CalicoConfig")
	}

	return configSpec, nil
}

func GetCalicoNetworkSettings(cluster *clusterapiv1beta1.Cluster) (string, string, error) {
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

func GetCalicoDataValuesFromAddonSecret(addonSecret *corev1.Secret) ([]byte, error) {
	calicoDataValues := addonSecret.Data[constants.TKGDataValueFileName]
	configSpec := &CalicoConfigSpec{}

	// This unmarshal and marshal process will filter out the redundant information inside of the addon secret.
	// The reason that we must do this is because calico package is implementing the schema to strictly control the input format.
	// Any redundant information will not be allowed.
	// These redundant information are mostly image configurations, which were used in kapp App when packages were not implemented,
	// but useless for now.
	err := yaml.Unmarshal(calicoDataValues, configSpec)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Could not unmarshal the calico configurations from the calico addon secret")
	}
	if configSpec.Calico.Config.VethMTU == "" {
		configSpec.Calico.Config.VethMTU = "0"
	}

	dataValueYamlBytes, err := yaml.Marshal(configSpec)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Could not marshal the the calico configurations to bytes")
	}
	return dataValueYamlBytes, nil
}
