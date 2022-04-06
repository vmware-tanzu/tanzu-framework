// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

// DataValues is the data values type generated by the VSphereCSIConfig CR
type DataValues struct {
	VSphereCSI   *DataValuesVSphereCSI   `yaml:"vsphereCSI,omitempty"`
	VSpherePVCSI *DataValuesVSpherePVCSI `yaml:"vspherePVCSI,omitempty"`
}

type DataValuesVSpherePVCSI struct {
	ClusterName                      string `yaml:"cluster_name"`
	ClusterUID                       string `yaml:"cluster_uid"`
	Namespace                        string `yaml:"namespace"`
	SupervisorMasterEndpointHostname string `yaml:"supervisor_master_endpoint_hostname"`
	SupervisorMasterPort             int32  `yaml:"supervisor_master_port"`
}

type DataValuesVSphereCSI struct {
	TLSThumbprint         string `yaml:"tlsThumbprint"`
	Namespace             string `yaml:"namespace"`
	ClusterName           string `yaml:"clusterName"`
	Server                string `yaml:"server"`
	Datacenter            string `yaml:"datacenter"`
	PublicNetwork         string `yaml:"publicNetwork"`
	Username              string `yaml:"username"`
	Password              string `yaml:"password"`
	Region                string `yaml:"region"`
	Zone                  string `yaml:"zone"`
	InsecureFlag          bool   `yaml:"insecureFlag"`
	UseTopologyCategories bool   `yaml:"useTopologyCategories"`
	ProvisionTimeout      string `yaml:"provisionTimeout"`
	AttachTimeout         string `yaml:"attachTimeout"`
	ResizerTimeout        string `yaml:"resizerTimeout"`
	VSphereVersion        string `yaml:"vSphereVersion"`
	HttpProxy             string `yaml:"http_proxy"`
	HttpsProxy            string `yaml:"https_proxy"`
	NoProxy               string `yaml:"no_proxy"`
	DeploymentReplicas    int32  `yaml:"deployment_replicas"`
	WindowsSupport        bool   `yaml:"windows_support"`
}
