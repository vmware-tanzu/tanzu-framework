package tkgauth

// TODO: this likely should live elsewhere
// looking at the GetClusterInfoFromCluster() func in this directory,
// it returns a *clientcmdapi.Cluster which is generated or something from
// some other location.  Need to look into this and confirm
// if these objects should exist in the same place.
type GetPinnipedSupervisorDiscoveryOptions struct {
	// the .well-known/openid-configuration discovery endpoint for a pinniped supervisor
	Endpoint string
	// a certificate bundle to trust in order to communicate with the pinniped supervisor
	CABundle string
}

type PinnipedSupervisorDiscoveryV1Alpha1 struct {
	PinnipedIdentityProvidersEndpoint string `json:"pinniped_identity_providers_endpoint,omitempty"`
}

type PinnipedSupervisorDiscoveryInfo struct {
	Issuer                           string                              `json:"issuer,omitempty"`
	AuthorizationEndpoint            string                              `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                    string                              `json:"token_endpoint,omitempty"`
	JWKSUri                          string                              `json:"jwks_uri,omitempty"`
	ResponseTypesSupported           []string                            `json:"response_types_supported,omitempty"`
	ResponseModesSupported           []string                            `json:"response_modes_supported,omitempty"`
	SubjectTypesSupported            []string                            `json:"subject_types_supported,omitempty"`
	IDTokenSigningALGValuesSupported []string                            `json:"id_token_signing_alg_values_supported,omitempty"`
	ScopesSupported                  []string                            `json:"scopes_supported,omitempty"`
	ClaimsSupported                  []string                            `json:"claims_supported,omitempty"`
	DiscoveryV1Aplha1                PinnipedSupervisorDiscoveryV1Alpha1 `json:"discovery.supervisor.pinniped.dev/v1alpha1,omitempty"`
}
