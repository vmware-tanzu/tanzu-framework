/*
Copyright 2020 The TKG Contributors.

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

package helper

// TestAllClusterComponentOptions describes options for
// Cluster creation with all it's dependent components
type TestAllClusterComponentOptions struct {
	ClusterName                 string
	Namespace                   string
	Labels                      map[string]string
	ClusterOptions              TestClusterOptions
	CPOptions                   TestCPOptions
	ListMDOptions               []TestMDOptions
	MachineOptions              []TestMachineOptions
	ClusterConfigurationOptions TestClusterConfiguration
	ClusterTopology             TestClusterTopology
	InfraComponentsOptions      TestInfraComponentsOptions
}

type TestInfraComponentsOptions struct {
	AWSCluster *TestAWSClusterOptions
}

type TestAWSClusterOptions struct {
	Name      string
	Namespace string
	Region    string
}

type TestClusterConfiguration struct {
	ImageRepository     string
	DNSImageRepository  string
	DNSImageTag         string
	EtcdLocalDataDir    string
	EtcdImageRepository string
	EtcdImageTag        string
}

type TestClusterTopology struct {
	Class   string
	Version string
}

// TestClusterOptions describes options for CAPI/TKC cluster
type TestClusterOptions struct {
	Phase                   string
	InfrastructureReady     bool
	ControlPlaneInitialized bool
	ControlPlaneReady       bool

	OperationType         string
	OperationtTimeout     int // seconds
	StartTimestamp        string
	LastObservedTimestamp string
}

// TestCPOptions describes options for ControlPlane
// This applies to KCP for CAPI and TKC.Spec.Topology.ControlPlane
// for TKC cluster
type TestCPOptions struct {
	SpecReplicas           int32
	ReadyReplicas          int32
	UpdatedReplicas        int32
	Replicas               int32
	K8sVersion             string
	InfrastructureTemplate TestObject
}

// TestObject describes options for Infrastructure Template
type TestObject struct {
	Name      string
	Namespace string
	Kind      string
}

// TestMDOptions describes options for MachineDeployment
type TestMDOptions struct {
	SpecReplicas           int32
	ReadyReplicas          int32
	UpdatedReplicas        int32
	Replicas               int32
	InfrastructureTemplate TestObject
}

// TestMachineOptions describes options for Machine
type TestMachineOptions struct {
	Phase      string
	K8sVersion string
	IsCP       bool
}

// TestDaemonSetOption describes options for DaemonSet
type TestDaemonSetOption struct {
	Name             string
	Namespace        string
	Image            string
	IncludeContainer bool
}

// TestDeploymentOption describes options for Deployment
type TestDeploymentOption struct {
	Name      string
	Namespace string
}

// TestClusterRoleBindingOption describes options for ClusterRoleBinding
type TestClusterRoleBindingOption struct {
	Name string
}

// TestClusterRoleOption describes options for ClusterRole
type TestClusterRoleOption struct {
	Name string
}

// TestServiceAccountOption describe options for ServiceAccount
type TestServiceAccountOption struct {
	Name      string
	Namespace string
}

// TestConfigMapOption describes options for ConfigMap
type TestConfigMapOption struct {
	Name      string
	Namespace string
}

// TestMachineHealthCheckOption describes options for MachineHealthCheck
type TestMachineHealthCheckOption struct {
	Name        string
	Namespace   string
	ClusterName string
}

// TestCLIPluginOption describes options for CLIPlugin
type TestCLIPluginOption struct {
	Name               string
	Description        string
	RecommendedVersion string
}
