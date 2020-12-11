package v1alpha1

const (
	// Secrets of type Addon
	AddonSecretType = "tkg.tanzu.vmware.com/addon"

	// Addon Name Label on Secret to indicate the name of addon to be installed
	AddonNameLabel = "tkg.tanzu.vmware.com/addon-name"

	// Cluster Name label on Secret to indicate the cluster on which addon is to be installed
	ClusterNameLabel = "tkg.tanzu.vmware.com/cluster-name"

	// AddonFinalizer
	AddonFinalizer = "tkg.tanzu.vmware.com/addon"

	// AddonType anotation
	AddonTypeAnnotation = "tkg.tanzu.vmware.com/addon-type"
)
