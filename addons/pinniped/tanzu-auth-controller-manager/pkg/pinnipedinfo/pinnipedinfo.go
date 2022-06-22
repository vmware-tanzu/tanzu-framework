// Package pinniped-info contains constants related to the pinniped-info ConfigMap.
//
// The pinniped-info ConfigMap is essentially part of our API right now. In the future, we intend to
// deprecated using this ConfigMap as part of our API and move to a CRD-based API.
package pinnipedinfo

const (
	// ConfigMapName is the namespace of the Pinniped Info Configmap
	ConfigMapNamespace = "kube-public"
	// ConfigMapName is the name of the Pinniped Info Configmap
	ConfigMapName = "pinniped-info"
)

const (
	// IssuerKey is the key for "issuer" field in the Pinniped Info Configmap
	IssuerKey = "issuer"
	// IssuerCABundleKey is the key for "issuer_ca_bundle_data" field in the Pinniped Info Configmap
	IssuerCABundleKey = "issuer_ca_bundle_data"
)
