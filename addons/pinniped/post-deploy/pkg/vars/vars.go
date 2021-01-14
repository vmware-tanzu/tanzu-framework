package vars

var (
	SupervisorNamespace          = "pinniped-supervisor"
	SupervisorSvcName            = ""
	ConciergeNamespace           = "pinniped-concierge"
	DexNamespace                 = "tanzu-system-auth"
	DexSvcName                   = "dexsvc"
	DexCertName                  = "dex-cert-tls"
	DexConfigMapName             = "dex"
	PinnipedOIDCProviderName     = "dex-oidc-identity-provider"
	PinnipedOIDCClientSecretName = "dex-client-credentials"
	SupervisorSvcEndpoint        = ""
	FederationDomainName         = ""
	JWTAuthenticatorName         = ""
	JWTAuthenticatorAudience     = ""
	SupervisorCertName           = ""
	SupervisorCABundleData       = ""
)
