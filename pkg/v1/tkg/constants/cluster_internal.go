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
	DefaultPacificClusterAPIVersion = "run.tanzu.vmware.com/v1alpha2"

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

	TkgNamespace = "tkg-system"

	TkrNamespace       = "tkr-system"
	TkrConfigMapName   = "tkr-controller-config"
	TkgPublicNamespace = "tkg-system-public"
	TmcNamespace       = "vmware-system-tmc"

	KappControllerNamespace     = "tkg-system"
	KappControllerConfigMapName = "kapp-controller-config"

	AddonsManagerDeploymentName  = "tanzu-addons-controller-manager"
	KappControllerDeploymentName = "kapp-controller"
	TkrControllerDeploymentName  = "tkr-controller-manager"
	KappControllerPackageName    = "kapp-controller"

	AkoStatefulSetName  = "ako"
	AkoAddonName        = "load-balancer-and-ingress-service"
	AkoNamespace        = "avi-system"
	AkoCleanupCondition = "ako.vmware.com/ObjectDeletionInProgress"

	ServiceDNSSuffix             = ".svc"
	ServiceDNSClusterLocalSuffix = ".svc.cluster.local"

	// TKGDataValueFormatString is required annotations for YTT data value file
	TKGDataValueFormatString = "#@data/values\n#@overlay/match-child-defaults missing_ok=True\n---\n"
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

const (
	// InfrastructureRefVSphere is the vSphere infrastructure
	InfrastructureRefVSphere = "VSphereCluster"
	// InfrastructureRefAWS is the AWS infrastructure
	InfrastructureRefAWS = "AWSCluster"
	// InfrastructureRefAzure is the Azure infrastructure
	InfrastructureRefAzure = "AzureCluster"
)

// networking constants
const (
	IPv4Family                 = "ipv4"
	IPv6Family                 = "ipv6"
	DualStackPrimaryIPv4Family = "ipv4,ipv6"
	DualStackPrimaryIPv6Family = "ipv6,ipv4"

	LocalHost     = "localhost"
	LocalHostIP   = "127.0.0.1"
	LocalHostIPv6 = "::1"

	LinkLocalAddress = "169.254.0.0/16"
	AzurePublicVIP   = "168.63.129.16"
)

// addons related constants
const (
	// AddonSecretType is the add on Secret type
	AddonSecretType = "tkg.tanzu.vmware.com/addon" // nolint:gosec
	// AddonNameLabel is the label on the Secret to indicate the name of addon to be installed
	AddonNameLabel = "tkg.tanzu.vmware.com/addon-name"
	// ClusterNameLabel is the label on the Secret to indicate the cluster on which addon is to be installed
	ClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"
)
