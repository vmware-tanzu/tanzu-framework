// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

const (
	CascadeControllerV1alpha1Name = "🐙 v1alpha1 tanzu auth cascade controller"
	CascadeControllerV1alpha3Name = "🐠 v1alpha3 tanzu auth cascade controller"
)

const (
	// TODO: I kinda don't like that we're copying a lot of these from addons/pkg/constants

	// tkgDataValueFileName is the default name of YTT data value field
	tkgDataValueFieldName = "values.yaml"

	// tkgDataOverlayFileName is the default name of YTT data overlay field
	tkgDataOverlayFieldName = "overlays.yaml"

	// secretNamespaceLogKey is the log key for "secret namespace"
	secretNamespaceLogKey = "secret namespace"

	// secretNameLogKey is the log key for "secret name"
	secretNameLogKey = "secret name"

	// clusterNamespaceLogKey is the log key for "cluster namespace"
	clusterNamespaceLogKey = "cluster namespace"

	// clusterNameLogKey is the log key for "cluster name"
	clusterNameLogKey = "cluster name"

	// tkgClusterNameLabel is the label on the Secret to indicate the cluster on which addon is to be installed
	tkgClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"

	// clusterBootstrapManagedSecret is the name for the secrets that are managed by ClusterBootstrapController
	clusterBootstrapManagedSecret = "clusterbootstrap-secret"

	// tkgAddonType is the type associated with a TKG addon secret
	tkgAddonType = "tkg.tanzu.vmware.com/addon"

	// tkgAddonTypeAnnotation is the addon type annotation
	tkgAddonTypeAnnotation = "tkg.tanzu.vmware.com/addon-type"

	// pinnipedAddonTypeAnnotation is the addon type annotation for Pinniped
	pinnipedAddonTypeAnnotation = "authentication/pinniped"

	// tkgAddonLabel is the label associated with a TKG addon secret
	tkgAddonLabel = "tkg.tanzu.vmware.com/addon-name"

	// pinnipedAddonLabel is the addon label for pinniped
	pinnipedAddonLabel = "pinniped"

	// packageNameLabel is the label on the cloned objects namely Secrets and Providers by "TanzuClusterBootstrap" Reconciler to indicate the package name
	packageNameLabel = "tkg.tanzu.vmware.com/package-name"

	// tkrLabel is the TKR label.
	tkrLabel = "tanzuKubernetesRelease"

	// tkrLabelClassyClusters is the TKR label for the clusters created using cluster-class
	tkrLabelClassyClusters = "run.tanzu.vmware.com/tkr"

	// Pinniped is the package label value for pinniped
	pinnipedPackageLabel = "pinniped"

	// pinnipedInfoConfigMapName is the name of the Pinniped Info Configmap
	pinnipedInfoConfigMapName = "pinniped-info"

	// Issuer is the key for "issuer" field in the Pinniped Info Configmap
	issuerKey = "issuer"

	// issuerCABundleKey is the key for "issuer_ca_bundle_data" field in the Pinniped Info Configmap
	issuerCABundleKey = "issuer_ca_bundle_data"

	// supervisorCABundleKey is the key for "supervisor_ca_bundle_data" field in the Pinniped ClusterBootstrap secret
	supervisorCABundleKey = "supervisor_ca_bundle_data"

	// supervisorEndpointKey is the key for "supervisor_svc_endpoint" field in the Pinniped ClusterBootstrap secret
	supervisorEndpointKey = "supervisor_svc_endpoint"

	// identityManagementTypeKey is the key for "identity_management_type" field in the Pinniped ClusterBootstrap secret
	identityManagementTypeKey = "identity_management_type"

	// kubePublicNamespace is the `kube-public` namespace
	kubePublicNamespace = "kube-public"

	// tkgManagementLabel is the label associated with a TKG management cluster
	tkgManagementLabel = "cluster-role.tkg.tanzu.vmware.com/management"

	// infrastructureRefDocker is the Docker infrastructure
	infrastructureRefDocker = "DockerCluster"

	// oidc is the string for oidc value for identity_management_type field in the Pinniped ClusterBootstrap secret
	oidc = "oidc"

	// none is the string for none value for identity_management_type field in the Pinniped ClusterBootstrap secret
	none = "none"

	// valuesYAMLPrefix is the data values prefix necessary in the addon secret
	valuesYAMLPrefix = `#@data/values
#@overlay/match-child-defaults missing_ok=True
---
`
)
