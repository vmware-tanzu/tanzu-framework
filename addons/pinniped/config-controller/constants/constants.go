package constants

const(
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
	TKGAddonLabel = "tkg.tanzu.vmware.com/addon-name"

	// PinnipedAddonLabel is the addon label for pinniped
	PinnipedAddonLabel = "pinniped"

	// TKGManagementLabel is the label associated with a TKG management cluster
	TKGManagementLabel = "cluster-role.tkg.tanzu.vmware.com/management"

	// PinnipedAddonTypeAnnotation is the addon type annotation for Pinniped
	PinnipedAddonTypeAnnotation = "authentication/pinniped"

	// TKGAddonTypeAnnotation is the addon type annotation
	TKGAddonTypeAnnotation = "tkg.tanzu.vmware.com/addon-type"
)
