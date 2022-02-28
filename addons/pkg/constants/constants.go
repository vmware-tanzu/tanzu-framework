// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants defines various constants used in the code.
package constants

import (
	"reflect"
	"time"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	/* Addon constants section */

	// CalicoAddonName is name of the Calico addon
	CalicoAddonName = "calico"

	// TKGBomNamespace is the TKG add on BOM namespace.
	TKGBomNamespace = "tkr-system"

	// TKRLabel is the TKR label.
	TKRLabel = "tanzuKubernetesRelease"

	// TKRLabelClassyClusters is the TKR label for the clusters created using cluster-class
	TKRLabelClassyClusters = "run.tanzu.vmware.com/tkr"

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

	// TKGCorePackageRepositoryComponentName is the name of component that includes the package and repository images
	TKGCorePackageRepositoryComponentName = "tkg-core-packages"

	// TKGCorePackageRepositoryImageName is the name of core package repository image
	TKGCorePackageRepositoryImageName = "tanzuCorePackageRepositoryImage"

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

	// InfrastructureRefDocker is the docker infrastructure
	InfrastructureRefDocker = "DockerCluster"

	// SecretNameLogKey is the log key for Secrets
	SecretNameLogKey = "secret-name"

	// ClusterBootstrapManagedSecret is the name for the secrets that are managed by ClusterBootstrapController
	ClusterBootstrapManagedSecret = "clusterbootstrap-secret"

	// DefaultCNISelectionClusterVariableName is the default cluster variable name for cni selection
	DefaultCNISelectionClusterVariableName = "tkg.tanzu.vmware.com/cni"

	// DefaultHTTPProxyClusterClassVarName is the default cluster variable name for HTTP proxy setting
	DefaultHTTPProxyClusterClassVarName = "tkg.tanzu.vmware.com/tkg-http-proxy"

	// DefaultHTTPSProxyClusterClassVarName is the default cluster variable name for HTTPS proxy setting
	DefaultHTTPSProxyClusterClassVarName = "tkg.tanzu.vmware.com/tkg-https-proxy"

	// DefaultNoProxyClusterClassVarName is the default cluster variable name for no proxy setting
	DefaultNoProxyClusterClassVarName = "tkg.tanzu.vmware.com/tkg-no-proxy"

	// DefaultProxyCaCertClusterClassVarName is the default cluster variable name for proxy CA cert
	DefaultProxyCaCertClusterClassVarName = "tkg.tanzu.vmware.com/tkg-proxy-ca-cert"

	// DefaultIPFamilyClusterClassVarName is the default cluster variable name for ip family
	DefaultIPFamilyClusterClassVarName = "tkg.tanzu.vmware.com/tkg-ip-family"

	// NSXTPodRoutingEnabledClassVarName is the cluster variable that indicates if NSXT pod routing is enabled or not
	NSXTPodRoutingEnabledClassVarName = "NSXT_POD_ROUTING_ENABLED"
)

// ClusterKind is the Kind for cluster-api Cluster object
var ClusterKind = reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name()
