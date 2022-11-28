// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants provides TKG constants
package constants

// cluster related constants used internally
const (
	ClusterClassFeature         = "vmware-system-tkg-clusterclass"
	TKCAPIFeature               = "vmware-system-tkg-tkc-api"
	TKGSClusterClassNamespace   = "vmware-system-tkg"
	TKGSTKCAPINamespace         = "vmware-system-tkg"
	CorePackagesNamespaceInTKGS = "vmware-system-tkg"
	CorePackagesNamespaceInTKGM = "tkg-system"

	ErrorMsgFeatureGateNotActivated = "vSphere with Tanzu environment detected, however, the feature '%v' is not activated in '%v' namespace"
	ErrorMsgFeatureGateStatus       = "error while checking feature '%v' status in namespace '%v'"
	ErrorMsgClusterExistsAlready    = "cluster with name %s already exists, please specify another name"
	ErrorMsgClusterListError        = "unable to get list of workload clusters managed by current management cluster"

	ErrorMsgCClassInputFeatureFlagDisabled = "input file is cluster class based but CLI feature flag '%v' is disabled, make sure its enabled to create cluster class based cluster"

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

	AddonsManagerDeploymentName      = "tanzu-addons-controller-manager"
	KappControllerDeploymentName     = "kapp-controller"
	TkrControllerDeploymentName      = "tkr-controller-manager"
	KappControllerPackageName        = "kapp-controller"
	CoreManagementPluginsPackageName = "tanzu-core-management-plugins"
	AKODeploymentName                = "load-balancer-and-ingress-service(ako)"

	AkoStatefulSetName       = "ako"
	AkoAddonName             = "load-balancer-and-ingress-service"
	AkoNamespace             = "avi-system"
	AkoCleanUpAnnotationKey  = "AviObjectDeletionStatus"
	AkoCleanUpFinishedStatus = "Done"

	ServiceDNSSuffix             = ".svc"
	ServiceDNSClusterLocalSuffix = ".svc.cluster.local"

	// TKGDataValueFormatString is required annotations for YTT data value file
	TKGDataValueFormatString = "#@data/values\n#@overlay/match-child-defaults missing_ok=True\n---\n"

	CAPVClusterSelectorKey = "capv.vmware.com/cluster.name"
)

// deployment plan constants
const (
	PlanDev    = "dev"
	PlanProd   = "prod"
	PlanDevCC  = "devcc"
	PlanProdCC = "prodcc"
)

// infrastructure provider name constants
const (
	InfrastructureProviderVSphere = "vsphere"
	InfrastructureProviderTkgs    = "tkgs"
	InfrastructureProviderAWS     = "aws"
	InfrastructureProviderAzure   = "azure"
	InfrastructureProviderDocker  = "docker"
	InfrastructureProviderOCI     = "oci"
)

var InfrastructureProviders = map[string]bool{
	InfrastructureProviderVSphere: true,
	InfrastructureProviderTkgs:    true,
	InfrastructureProviderAWS:     true,
	InfrastructureProviderAzure:   true,
	InfrastructureProviderDocker:  true,
	InfrastructureProviderOCI:     true,
}

const (
	// InfrastructureRefVSphere is the vSphere infrastructure
	InfrastructureRefVSphere = "VSphereCluster"
	// InfrastructureRefAWS is the AWS infrastructure
	InfrastructureRefAWS = "AWSCluster"
	// InfrastructureRefAzure is the Azure infrastructure
	InfrastructureRefAzure = "AzureCluster"
	// InfrastructureRefDocker is the docker infrastructure
	InfrastructureRefDocker = "DockerCluster"
	// InfrastructureRefOCI is the Oracle Cloud infrastructure
	InfrastructureRefOCI = "OCICluster"
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
	// ClusterPauseLabel is the label on the Cluster Object to indicate the cluster is paused by TKG
	ClusterPauseLabel = "tkg.tanzu.vmware.com/paused"
	// PackageTypeLabel is the label on the PackageInstall which mentions type of the package
	PackageTypeLabel = "tkg.tanzu.vmware.com/package-type"
	// CLIPluginImageRepositoryOverrideLabel is the label on the configmap which specifies CLIPlugin image repository override
	CLIPluginImageRepositoryOverrideLabel = "cli.tanzu.vmware.com/cliplugin-image-repository-override"
	// KappController labels
	KappControllerAppLabel         = "kapp.k14s.io/app"
	KappControllerAssociationLabel = "kapp.k14s.io/association"
)

// TKG management package related constants
const (
	TKGManagementPackageName           = "tkg.tanzu.vmware.com"
	TKGManagementPackageInstallName    = "tkg-pkg"
	TKGManagementPackageRepositoryName = "tanzu-management"
	PackageTypeManagement              = "management"
)

// AviControllerVersionRegex Avi aka. NSX Advanced LoadBalancer version regex expression
const AviControllerVersionRegex = `^\d+(\.\d+)*$`

const (
	TanzuCLISystemNamespace = "tanzu-cli-system"
)

// Kind constants
const (
	KindCluster                     = "Cluster"
	KindTanzuKubernetesCluster      = "TanzuKubernetesCluster"
	KindClusterClass                = "ClusterClass"
	KindKubeadmControlPlaneTemplate = "KubeadmControlPlaneTemplate"
	KindKubeadmConfigTemplate       = "KubeadmConfigTemplate"
	KindAWSClusterTemplate          = "AWSClusterTemplate"
	KindAWSMachineTemplate          = "AWSMachineTemplate"
	KindAzureClusterTemplate        = "AzureClusterTemplate"
	KindAzureMachineTemplate        = "AzureMachineTemplate"
	KindVSphereClusterTemplate      = "VSphereClusterTemplate"
	KindVSphereMachineTemplate      = "VSphereMachineTemplate"
	KindDockerMachineTemplate       = "DockerMachineTemplate"
)

// Resources constants
const (
	ResourceClusterClass                = "clusterclasses"
	ResourceKubeadmControlPlaneTemplate = "kubeadmcontrolplanetemplates"
	ResourceKubeadmConfigTemplate       = "kubeadmconfigtemplates"
	ResourceAWSClusterTemplate          = "awsclustertemplates"
	ResourceAWSMachineTemplate          = "awsmachinetemplates"
	ResourceAzureClusterTemplate        = "azureclustertemplates"
	ResourceAzureMachineTemplate        = "azuremachinetemplates"
	ResourceVSphereClusterTemplate      = "vsphereclustertemplates"
	ResourceVSphereMachineTemplate      = "vspheremachinetemplates"
)

type OperationType int

const (
	OperationTypeInstall OperationType = iota
	OperationTypeUpgrade
)
