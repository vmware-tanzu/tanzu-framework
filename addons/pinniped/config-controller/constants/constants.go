// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package constants defines constants used throughout the codebase.
package constants

import (
	"reflect"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	//TODO: I kinda don't like that we're copying a lot of these from addons/pkg/constants

	// TKGDataValueFileName is the default name of YTT data value field
	TKGDataValueFieldName = "values.yaml"

	// TKGDataOverlayFileName is the default name of YTT data overlay field
	TKGDataOverlayFieldName = "overlays.yaml"

	// SecretNamespaceLogKey is the log key for "secret namespace"
	SecretNamespaceLogKey = "secret namespace"

	// SecretNameLogKey is the log key for "secret name"
	SecretNameLogKey = "secret name"

	// ClusterNamespaceLogKey is the log key for "cluster namespace"
	ClusterNamespaceLogKey = "cluster namespace"

	// ClusterNameLogKey is the log key for "cluster name"
	ClusterNameLogKey = "cluster name"

	// TKGClusterNameLabel is the label on the Secret to indicate the cluster on which addon is to be installed
	TKGClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"

	// ClusterBootstrapManagedSecret is the name for the secrets that are managed by ClusterBootstrapController
	ClusterBootstrapManagedSecret = "clusterbootstrap-secret"

	// TKGAddonType is the label associated with a TKG addon secret
	TKGAddonLabel = "tkg.tanzu.vmware.com/addon-name"

	// PackageNameLabel is the label on the cloned objects namely Secrets and Providers by "TanzuClusterBootstrap" Reconciler to indicate the package name
	PackageNameLabel = "tkg.tanzu.vmware.com/package-name"

	// TKRLabelClassyClusters is the TKR label for the clusters created using cluster-class
	TKRLabelClassyClusters = "run.tanzu.vmware.com/tkr"

	// Pinniped is the package label value for pinniped
	PinnipedPackageLabel = "pinniped"

	// PinnipedInfoConfigMapName is the name of the Pinniped Info Configmap
	PinnipedInfoConfigMapName = "pinniped-info"

	// Issuer is the key for "issuer" field in the Pinniped Info Configmap
	IssuerKey = "issuer"

	// IssuerCABundleKey is the key for "issuer_ca_bundle_data" field in the Pinniped Info Configmap
	IssuerCABundleKey = "issuer_ca_bundle_data"

	// SupervisorCABundleKey is the key for "supervisor_ca_bundle_data" field in the Pinniped ClusterBootstrap secret
	SupervisorCABundleKey = "supervisor_ca_bundle_data"

	// SupervisorEndpointKey is the key for "supervisor_svc_endpoint" field in the Pinniped ClusterBootstrap secret
	SupervisorEndpointKey = "supervisor_svc_endpoint"

	// IdentityManagementTypeKey is the key for "identity_management_type" field in the Pinniped ClusterBootstrap secret
	IdentityManagementTypeKey = "identity_management_type"

	// KubePublicNamespace is the `kube-public` namespace
	KubePublicNamespace = "kube-public"

	// TKGManagementLabel is the label associated with a TKG management cluster
	TKGManagementLabel = "cluster-role.tkg.tanzu.vmware.com/management"

	// InfrastructureRefDocker is the Docker infrastructure
	InfrastructureRefDocker = "DockerCluster"

	// OIDC is the string for OIDC value for identity_management_type field in the Pinniped ClusterBootstrap secret
	OIDC = "oidc"

	// None is the string for none value for identity_management_type field in the Pinniped ClusterBootstrap secret
	None = "none"
)

// ClusterKind is the Kind for cluster-api Cluster object
var ClusterKind = reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name()
