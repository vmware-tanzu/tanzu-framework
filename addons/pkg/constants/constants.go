// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants defines various constants used in the code.
package constants

import (
	"reflect"
	"time"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	/* Addon constants section */

	// AddonControllerManagerLeaderElectionResourceName is the name of the resource that leader election of
	// addons controller manager will use for holding the leader lock.
	AddonControllerManagerLeaderElectionResourceName = "tanzu-addons-manager-leader-lock"

	// CalicoAddonName is name of the Calico addon
	CalicoAddonName = "calico"

	// CPIAddonName is name of the cloud-provider-vsphere addon
	CPIAddonName = "vsphere-cpi"

	// PVCSIAddonName is name of the vsphere-pv-csi addon
	PVCSIAddonName = "vsphere-pv-csi"

	// CSIAddonName is name of the vsphere-csi addon
	CSIAddonName = "vsphere-csi"

	// AwsEbsCSIAddonName is name of the vsphere-csi addon
	AwsEbsCSIAddonName = "aws-ebs-csi"

	// TKGBomNamespace is the TKG add on BOM namespace.
	TKGBomNamespace = "tkr-system"

	// TKRLabel is the TKR label.
	TKRLabel = "tanzuKubernetesRelease"

	// TKRLabelClassyClusters is the TKR label for the clusters created using cluster-class
	TKRLabelClassyClusters = "run.tanzu.vmware.com/tkr"

	// TKRLabelLegacyClusters is the TKR label for legacy clusters
	TKRLabelLegacyClusters = "run.tanzu.vmware.com/legacy-tkr"

	// TKGAnnotationTemplateConfig is the TKG annotation for addon config CRs used by ClusterBootstrapTemplate
	TKGAnnotationTemplateConfig = "tkg.tanzu.vmware.com/template-config"

	// TKGBomContent is the TKG BOM content.
	TKGBomContent = "bomContent"

	// TKRConfigmapName is the name of TKR config map
	TKRConfigmapName = "tkr-controller-config"

	// TKRRepoKey is the key for image repository in TKR config map data.
	TKRRepoKey = "imageRepository"

	// TKGPackageReconcilerKey is the log key for "name".
	TKGPackageReconcilerKey = "Package"

	// TKGAppReconcilerKey is the log key for "name".
	TKGAppReconcilerKey = "App"

	// TKGDataValueFormatString is required annotations for YTT data value file
	TKGDataValueFormatString = "#@data/values\n#@overlay/match-child-defaults missing_ok=True\n---\n"

	// TKGDataValueFileName is the default name of YTT data value file
	TKGDataValueFileName = "values.yaml"

	// TKGSDataValueFileName is the default name of YTT data value file for TKR info
	TKGSDataValueFileName = "tkgs-values.yaml"

	// TKGCorePackageRepositoryComponentName is the name of component that includes the package and repository images
	TKGCorePackageRepositoryComponentName = "tkg-core-packages"

	// TKGCorePackageRepositoryImageName is the name of core package repository image
	TKGCorePackageRepositoryImageName = "tanzuCorePackageRepositoryImage"

	// TKGSDeploymentUpdateStrategy is the update strategy used by TKGS deployments
	TKGSDeploymentUpdateStrategy = "RollingUpdate"

	// TKGSDeploymentUpdateMaxSurge is the MaxSurge used by TKGS deployments rollingUpdate
	TKGSDeploymentUpdateMaxSurge = 1

	// TKGSDeploymentUpdateMaxUnavailable is the MaxUnavailableused by TKGS deployments rollingUpdate
	TKGSDeploymentUpdateMaxUnavailable = 0

	// TKGSDaemonsetUpdateStrategy is the update strategy used by TKGS daemonsets
	TKGSDaemonsetUpdateStrategy = "OnDelete"

	/* log key section */

	// NameLogKey is the log key for "name".
	NameLogKey = "name"

	// NamespaceLogKey is the log key for "namespace".
	NamespaceLogKey = "namespace"

	// AddonSecretNameLogKey is the log key for "addon-secret-name".
	AddonSecretNameLogKey = "addon-secret-name"

	// AddonSecretNamespaceLogKey is the log key for "addon-secret-ns"
	AddonSecretNamespaceLogKey = "addon-secret-ns" // nolint:gosec

	// AddonNameLogKey is the log key for "addon-name"
	AddonNameLogKey = "addon-name"

	// ImageNameLogKey is the log key for "image-name".
	ImageNameLogKey = "image-name"

	// ImageURLLogKey is the log key for "image-url".
	ImageURLLogKey = "image-url"

	// ComponentNameLogKey is the log key for "component-name".
	ComponentNameLogKey = "component-name"

	// KCPNameLogKey is the log key for "kcp-name"
	KCPNameLogKey = "kcp-name"

	// KCPNamespaceLogKey is the log key for "kcp-ns"
	KCPNamespaceLogKey = "kcp-ns"

	// TKRNameLogKey is the log key for "tkr-name"
	TKRNameLogKey = "tkr-name"

	// ClusterNameLogKey is the log key for "cluster-name"
	ClusterNameLogKey = "cluster-name"

	// ClusterNamespaceLogKey is the log key for "cluster-ns"
	ClusterNamespaceLogKey = "cluster-ns"

	// BOMNameLogKey is the log key for "bom-name"
	BOMNameLogKey = "bom-name"

	// BOMNamespaceLogKey is the log key for "bom-ns"
	BOMNamespaceLogKey = "bom-ns"

	// PackageRepositoryLogKey is the log key for "core-package-repository"
	PackageRepositoryLogKey = "core-package-repository"

	// AddonControllerName is name of addon-controller
	AddonControllerName = "addon-controller"

	// CRDWaitPollInterval is poll interval for checking server resources
	CRDWaitPollInterval = time.Second * 5

	// CRDWaitPollTimeout is poll timeout for checking server resources
	CRDWaitPollTimeout = time.Minute * 10

	// ClusterBootstrapNameLogKey is the log key for "ClusterBootstrapNameLogKey"
	ClusterBootstrapNameLogKey = "clusterbootstrap-name"

	// TKGSystemNS is the TKG system namespace.
	TKGSystemNS = "tkg-system"

	// DiscoveryCacheInvalidateInterval is the interval for invalidating cache
	DiscoveryCacheInvalidateInterval = time.Minute * 10

	// AntreaAddonName is the name of Antrea Addon Controller
	AntreaAddonName = "antrea"

	// KappControllerAddonName is the addon name of Kapp Controller
	KappControllerAddonName = "kapp-controller"

	// SecretNameLogKey is the log key for Secrets
	SecretNameLogKey = "secret-name"

	// ClusterBootstrapManagedSecret is the name for the secrets that are managed by ClusterBootstrapController
	ClusterBootstrapManagedSecret = "clusterbootstrap-secret"

	// DefaultIPFamilyClusterClassVarName is the default cluster variable name for ip family
	DefaultIPFamilyClusterClassVarName = "tkg.tanzu.vmware.com/tkg-ip-family"

	// PackageInstallServiceAccount is service account name used for PackageInstall
	PackageInstallServiceAccount = "tanzu-cluster-bootstrap-sa"

	// PackageInstallClusterRole is cluster role name used for PackageInstall
	PackageInstallClusterRole = "tanzu-cluster-bootstrap-clusterrole"

	// PackageInstallClusterRoleBinding is cluster role binding name used for PackageInstall
	PackageInstallClusterRoleBinding = "tanzu-cluster-bootstrap-clusterrolebinding"

	// PackageInstallSyncPeriod is the sync period for kapp-controller to periodically reconcile a PackageInstall
	PackageInstallSyncPeriod = time.Minute * 10

	// RequeueAfterDuration determines the duration after which the Controller should requeue the reconcile key
	RequeueAfterDuration = time.Second * 10

	// WebhookCertDir is the directory where the certificate and key are stored for webhook server TLS handshake
	WebhookCertDir = "/tmp/k8s-webhook-server/serving-certs"

	// WebhookCertManagementFrequency is how often the certificates for webhook server TLS are managed
	WebhookCertManagementFrequency = time.Minute * 60

	// WebhookCertLifeTime is how long the webhook server TLS certificates are good for
	WebhookCertLifeTime = time.Hour * 24 * 7

	// WebhookServiceName is the name of the k8s service that serves the admission requests
	WebhookServiceName = "tanzu-addons-manager-webhook-service"

	// WebhookScrtName is the name of secret that holds certificates and key for webhook service
	WebhookScrtName = "webhook-tls"

	// AddonWebhookLabelKey is the key for the label for addon admission webhooks
	AddonWebhookLabelKey = "tkg.tanzu.vmware.com/addon-webhooks"

	// AddonWebhookLabelValue is the value for the label for addon admission webhooks
	AddonWebhookLabelValue = "addon-webhooks"

	// LocalObjectRefSuffix is the suffix of a field within the provider's CR. This suffix indicates that the field is a
	// K8S typed local object reference
	LocalObjectRefSuffix = "LocalObjRef"

	// AddCBMissingFieldsAnnotationKey is the annotation key used by ClusterBootstrap webhook to implement its defaulting
	// logic
	AddCBMissingFieldsAnnotationKey = "tkg.tanzu.vmware.com/add-missing-fields-from-tkr"

	// VsphereCPIProviderServiceAccountAggregatedClusterRole is the name of ClusterRole created by controllers that use ProviderServiceAccount
	VsphereCPIProviderServiceAccountAggregatedClusterRole = "addons-vsphere-cpi-providerserviceaccount-aggregatedrole"

	// VsphereCSIProviderServiceAccountAggregatedClusterRole is the name of ClusterRole created by controllers that use ProviderServiceAccount
	VsphereCSIProviderServiceAccountAggregatedClusterRole = "addons-vsphere-csi-providerserviceaccount-aggregatedrole"

	// CAPVClusterRoleAggregationRuleLabelSelectorKey is the label selector key used by aggregation rule in CAPV ClusterRole
	CAPVClusterRoleAggregationRuleLabelSelectorKey = "capv.infrastucture.cluster.x-k8s.io/aggregate-to-manager"

	// CAPVClusterRoleAggregationRuleLabelSelectorValue is the label selector value used by aggregation rule in CAPV ClusterRole
	CAPVClusterRoleAggregationRuleLabelSelectorValue = "true"

	// PackageInstallStatusControllerRateLimitBaseDelay is the base delay for rate limiting error requeues in PackageInstallStatusController
	PackageInstallStatusControllerRateLimitBaseDelay = time.Second * 10

	// PackageInstallStatusControllerRateLimitMaxDelay is the maximum delay for rate limiting error requeues in PackageInstallStatusController
	PackageInstallStatusControllerRateLimitMaxDelay = time.Minute * 30

	// ClusterPauseLabel is the label on the Cluster Object to indicate the cluster is paused by TKG
	ClusterPauseLabel = "tkg.tanzu.vmware.com/paused"

	// ManagementClusterRoleLabel is the management cluster role label
	// It indicates the cluster object represents a mgmt cluster
	ManagementClusterRoleLabel = "cluster-role.tkg.tanzu.vmware.com/management"

	// InfrastructureProviderVSphere is the key for vsphere infrastructure
	InfrastructureProviderVSphere = "vsphere"

	// InfrastructureProviderTkgs is the key for vsphere infrastructure with supervisor
	InfrastructureProviderTkgs = "tkgs"

	// InfrastructureProviderAWS is the key for aws infrastructure
	InfrastructureProviderAWS = "aws"

	// InfrastructureProviderAzure is the key for azure infrastructure
	InfrastructureProviderAzure = "azure"

	// InfrastructureProviderDocker is the key for docker infrastructure
	InfrastructureProviderDocker = "docker"

	// InfrastructureRefVSphere is the vSphere infrastructure
	InfrastructureRefVSphere = "VSphereCluster"

	// InfrastructureRefAWS is the AWS infrastructure
	InfrastructureRefAWS = "AWSCluster"

	// InfrastructureRefAzure is the Azure infrastructure
	InfrastructureRefAzure = "AzureCluster"

	// InfrastructureRefDocker is the docker infrastructure
	InfrastructureRefDocker = "DockerCluster"

	// CAPVClusterSelectorKey is the selector key used by capv
	CAPVClusterSelectorKey = "capv.vmware.com/cluster.name"
)

var (
	// ClusterKind is the Kind for cluster-api Cluster object
	ClusterKind = reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name()

	// AntreaConfigKind is the Kind for cni AntreaConfig object
	AntreaConfigKind = reflect.TypeOf(cniv1alpha1.AntreaConfig{}).Name()

	// CalicoConfigKind is the Kind for cni CalicoConfig object
	CalicoConfigKind = reflect.TypeOf(cniv1alpha1.CalicoConfig{}).Name()

	// VSphereCSIConfigKind is the Kind for csi VSphereCSIConfig object
	VSphereCSIConfigKind = reflect.TypeOf(csiv1alpha1.VSphereCSIConfig{}).Name()

	// VSphereCPIConfigKind is the Kind for cpi VSphereCPIConfig object
	VSphereCPIConfigKind = reflect.TypeOf(cpiv1alpha1.VSphereCPIConfig{}).Name()

	// KappControllerConfigKind is the Kind for KappControllerConfig object
	KappControllerConfigKind = reflect.TypeOf(runv1alpha3.KappControllerConfig{}).Name()

	// VSphereCSIConfigKind is the Kind for csi AwsEbsCSIConfig object
	AwsEbsCSIConfigKind = reflect.TypeOf(csiv1alpha1.AwsEbsCSIConfig{}).Name()
)
