package constants

import (
	"reflect"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const(
	//TODO: I kinda don't like that we're copying a lot of these from addons/pkg/constants

	// TKGDataValueFileName is the default name of YTT data value field
	TKGDataValueFieldName = "values.yaml"

	// NamespaceLogKey is the log key for "namespace"
	NamespaceLogKey = "namespace"

	// NameLogKey is the log key for "name"
	NameLogKey = "name"

	// TKGClusterNameLabel is the label on the Secret to indicate the cluster on which addon is to be installed
	TKGClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"

	// TKGAddonType is the type associated with a TKG addon secret
	TKGAddonType = "tkg.tanzu.vmware.com/addon"

	// TKGAddonType is the label associated with a TKG addon secret
	// change to package-name for v1alpha3label
	// what will happen to old v1alpha1 secrets?
	// will mgmt cluster that we pull info from still be v1alpha1?
	// patch CB CR for workload cluster
	TKGAddonLabel = "tkg.tanzu.vmware.com/addon-name"

	// TKRLabelClassyClusters is the TKR label for the clusters created using cluster-class
	TKRLabelClassyClusters = "run.tanzu.vmware.com/tkr"

	// PinnipedAddonLabel is the addon label for pinniped
	PinnipedAddonLabel = "pinniped"

	// PinnipedInfoConfigMapName is the name of the Pinniped Info Configmap
	PinnipedInfoConfigMapName = "pinniped-info"

	// KubePublicNamespace is the `kube-public` namespace
	KubePublicNamespace = "kube-public"

	// TKGManagementLabel is the label associated with a TKG management cluster
	TKGManagementLabel = "cluster-role.tkg.tanzu.vmware.com/management"

	// PinnipedAddonTypeAnnotation is the addon type annotation for Pinniped
	PinnipedAddonTypeAnnotation = "authentication/pinniped"

	// TKGAddonTypeAnnotation is the addon type annotation
	TKGAddonTypeAnnotation = "tkg.tanzu.vmware.com/addon-type"

	// InfrastructureRefDocker is the Docker infrastructure
	InfrastructureRefDocker = "DockerCluster"
)

// ClusterKind is the Kind for cluster-api Cluster object
var ClusterKind = reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name()
