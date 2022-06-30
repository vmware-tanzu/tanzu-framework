// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"gopkg.in/yaml.v3"
)

// VSphereCPIDataValues serializes the CPIConfig CR
type VSphereCPIDataValues interface {
	// Serialize the struct into a data format expected by the consumer. This could be yaml or json
	Serialize() ([]byte, error)
}
type VSphereCPINonParaVirtDataValues struct {
	Mode                               string                 `yaml:"mode"`
	TLSThumbprint                      string                 `yaml:"tlsThumbprint"`
	Server                             string                 `yaml:"server"`
	Datacenter                         string                 `yaml:"datacenter"`
	Username                           string                 `yaml:"username"`
	Password                           string                 `yaml:"password"`
	Region                             string                 `yaml:"region"`
	Zone                               string                 `yaml:"zone"`
	InsecureFlag                       bool                   `yaml:"insecureFlag"`
	IPFamily                           string                 `yaml:"ipFamily"`
	VMInternalNetwork                  string                 `yaml:"vmInternalNetwork"`
	VMExternalNetwork                  string                 `yaml:"vmExternalNetwork"`
	VMExcludeInternalNetworkSubnetCidr string                 `yaml:"vmExcludeInternalNetworkSubnetCidr"`
	VMExcludeExternalNetworkSubnetCidr string                 `yaml:"vmExcludeExternalNetworkSubnetCidr"`
	CloudProviderExtraArgs             CloudProviderExtraArgs `yaml:"cloudProviderExtraArgs"`
	Nsxt                               struct {
		PodRoutingEnabled bool `yaml:"podRoutingEnabled"`
		Routes            struct {
			RouterPath  string `yaml:"routerPath"`
			ClusterCidr string `yaml:"clusterCidr"`
		} `yaml:"routes"`
		Username          string `yaml:"username"`
		Password          string `yaml:"password"`
		Host              string `yaml:"host"`
		Insecure          bool   `yaml:"insecure"`
		RemoteAuthEnabled bool   `yaml:"remoteAuthEnabled"`
		VmcAccessToken    string `yaml:"vmcAccessToken"`
		VmcAuthHost       string `yaml:"vmcAuthHost"`
		ClientCertKeyData string `yaml:"clientCertKeyData"`
		ClientCertData    string `yaml:"clientCertData"`
		RootCAData        string `yaml:"rootCAData"`
		SecretName        string `yaml:"secretName"`
		SecretNamespace   string `yaml:"secretNamespace"`
	} `yaml:"nsxt"`
	HTTPProxy  string `yaml:"http_proxy"`
	HTTPSProxy string `yaml:"https_proxy"`
	NoProxy    string `yaml:"no_proxy"`
}

func (v *VSphereCPINonParaVirtDataValues) Serialize() ([]byte, error) {
	dataValues := struct {
		DataValues VSphereCPINonParaVirtDataValues `yaml:"vsphereCPI"`
	}{DataValues: *v}
	return yaml.Marshal(dataValues)
}

type CloudProviderExtraArgs struct {
	TLSCipherSuites string `yaml:"tls-cipher-suites"`
}

type VSphereCPIParaVirtDataValues struct {
	Mode                       string `yaml:"mode"`
	ClusterAPIVersion          string `yaml:"clusterAPIVersion"`
	ClusterKind                string `yaml:"clusterKind"`
	ClusterName                string `yaml:"clusterName"`
	ClusterUID                 string `yaml:"clusterUID"`
	SupervisorMasterEndpointIP string `yaml:"supervisorMasterEndpointIP"`
	SupervisorMasterPort       string `yaml:"supervisorMasterPort"`
	AntreaNSXPodRoutingEnabled bool   `yaml:"antreaNSXPodRoutingEnabled"`
}

func (v *VSphereCPIParaVirtDataValues) Serialize() ([]byte, error) {
	dataValues := struct {
		VSphereCPI VSphereCPIParaVirtDataValues `yaml:"vsphereCPI"`
	}{VSphereCPI: *v}
	return yaml.Marshal(dataValues)
}
