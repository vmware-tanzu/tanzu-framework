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

// Package constants provides TKG constants
package constants

// cluster related constants used internally
const (
	KindCluster                = "Cluster"
	KindTanzuKubernetesCluster = "TanzuKubernetesCluster"

	PacificGCMControllerDeployment = "vmware-system-tkg-controller-manager"
	PacificGCMControllerNamespace  = "vmware-system-tkg"
	// PacificClusterKind vsphere-pacific provider work load cluster kind
	PacificClusterKind              = "TanzuKubernetesCluster"
	PacificClusterListKind          = "TanzuKubernetesClusterList"
	DefaultPacificClusterAPIVersion = "run.tanzu.vmware.com/v1alpha1"

	CronJobKind    = "CronJob"
	CeipNamespace  = "tkg-system-telemetry"
	CeipAPIVersion = "batch/v1beta1"
	CeipJobName    = "tkg-telemetry"

	AntreaDeploymentName      = "antrea-controller"
	AntreaDeploymentNamespace = "kube-system"
	CalicoDeploymentName      = "calico-kube-controllers"
	CalicoDeploymentNamespace = "kube-system"

	TanzuRunAPIGroupPath = "/apis/run.tanzu.vmware.com"

	PinnipedSupervisorNameSpace              = "pinniped-supervisor"
	PinnipedFederationDomainObjectName       = "pinniped-federation-domain"
	PinnipedFederationDomainObjectKind       = "FederationDomain"
	PinnipedFederationDomainObjectAPIVersion = "config.supervisor.pinniped.dev/v1alpha1"
	PinnipedSupervisorDefaultTLSSecretName   = "pinniped-supervisor-default-tls-certificate" // #nosec

	TkrNamespace       = "tkr-system"
	TkrConfigMapName   = "tkr-controller-config"
	TkgPublicNamespace = "tkg-system-public"

	KappControllerNamespace     = "tkg-system"
	KappControllerConfigMapName = "kapp-controller-config"

	AddonsManagerDeploymentName  = "tanzu-addons-controller-manager"
	KappControllerDeploymentName = "kapp-controller"
	TkrControllerDeploymentName  = "tkr-controller-manager"

	ServiceDNSSuffix             = ".svc"
	ServiceDNSClusterLocalSuffix = ".svc.cluster.local"

	LocalHost   = "localhost"
	LocalHostIP = "127.0.0.1"

	LinkLocalAddress = "169.254.0.0/16"
	AzurePublicVIP   = "168.63.129.16"
)

// deployment plan constants
const (
	PlanDev  = "dev"
	PlanProd = "prod"
)

// infrastructure provider name constants
const (
	InfrastructureProviderVSphere = "vsphere"
	InfrastructureProviderAWS     = "aws"
	InfrastructureProviderAzure   = "azure"
	InfrastructureProviderDocker  = "docker"
)

// machine template name constants
const (
	VSphereMachineTemplate = "VSphereMachineTemplate"
	AWSMachineTemplate     = "AWSMachineTemplate"
	AzureMachineTemplate   = "AzureMachineTemplate"
	DockerMachineTemplate  = "DockerMachineTemplate"
)
