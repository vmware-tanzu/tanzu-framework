// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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

// networking constants
const (
	IPv6Family = "ipv6"

	LocalHost     = "localhost"
	LocalHostIP   = "127.0.0.1"
	LocalHostIPv6 = "::1"

	LinkLocalAddress = "169.254.0.0/16"
	AzurePublicVIP   = "168.63.129.16"
)
