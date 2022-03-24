// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

// DataValues is the data values type generated by the VSphereCSIConfig CR
type DataValues struct {
	VSphereCSI   *DataValuesVSphereCSI   `yaml:"vsphereCSI,omitempty"`
	VSpherePVCSI *DataValuesVSpherePVCSI `yaml:"vspherePVCSI,omitempty"`
}

type DataValuesVSpherePVCSI struct {
	ClusterName                      string `yaml:"clusterName"`
	ClusterUID                       string `yaml:"clusterUID"`
	Namespace                        string `yaml:"namespace"`
	SupervisorMasterEndpointHostname string `yaml:"supervisorMasterEndpointHostname"`
	SupervisorMasterPort             int32  `yaml:"supervisorMasterPort"`
}

type DataValuesVSphereCSI struct {
	Namespace             string `yaml:"namespace"`
	ClusterName           string `yaml:"clusterName"`
	Server                string `yaml:"server"`
	Datacenter            string `yaml:"datacenter"`
	PublicNetwork         string `yaml:"publicNetwork"`
	Username              string `yaml:"username"`
	Password              string `yaml:"password"`
	Region                string `yaml:"region,omitempty"`
	Zone                  string `yaml:"zone,omitempty"`
	UseTopologyCategories *bool  `yaml:"useTopologyCategories,omitempty"`
	ProvisionTimeout      string `yaml:"provisionTimeout,omitempty"`
	AttachTimeout         string `yaml:"attachTimeout,omitempty"`
	ResizerTimeout        string `yaml:"resizerTimeout,omitempty"`
	VSphereVersion        string `yaml:"vSphereVersion,omitempty"`
	HttpProxy             string `yaml:"httpProxy,omitempty"`
	HttpsProxy            string `yaml:"httpsProxy,omitempty"`
	NoProxy               string `yaml:"noProxy,omitempty"`
	DeploymentReplicas    *int32 `yaml:"deploymentReplicas,omitempty"`
	WindowsSupport        *bool  `yaml:"windowsSupport,omitempty"`
}
