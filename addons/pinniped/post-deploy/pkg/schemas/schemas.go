package schemas

type DexConfig struct {
	Issuer           string          `yaml:"issuer,omitempty"`
	FrontEnd         *frontEnd       `yaml:"frontend,omitempty"`
	Web              *web            `yaml:"web,omitempty"`
	Expiry           *expiry         `yaml:"expiry,omitempty"`
	Logger           *logger         `yaml:"logger,omitempty"`
	StaticClients    []*StaticClient `yaml:"staticClients,omitempty"`
	Connectors       []*connector    `yaml:"connectors,omitempty"`
	Oauth2           *oauth2         `yaml:"oauth2,omitempty"`
	Storage          *storage        `yaml:"storage,omitempty"`
	EnablePasswordDb bool            `yaml:"enablePasswordDB"`
}

type frontEnd struct {
	Theme string `yaml:"theme,omitempty"`
}

type web struct {
	HTTPS   string `yaml:"https,omitempty"`
	TlsCert string `yaml:"tlsCert,omitempty"`
	TlsKey  string `yaml:"tlsKey,omitempty"`
}

type expiry struct {
	SigningKeys string `yaml:"signingKeys,omitempty"`
	IdTokens    string `yaml:"idTokens,omitempty"`
}

type logger struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
}

type StaticClient struct {
	Id           string   `yaml:"id,omitempty"`
	Name         string   `yaml:"name,omitempty"`
	RedirectURIs []string `yaml:"redirectURIs,omitempty"`
	Secret       string   `yaml:"secret,omitempty"`
}

type connector struct {
	Type   string `yaml:"type,omitempty"`
	Id     string `yaml:"id,omitempty"`
	Name   string `yaml:"name,omitempty"`
	Config struct {
		Issuer                    string            `yaml:"issuer,omitempty"`
		ClientID                  string            `yaml:"clientID,omitempty"`
		ClientSecret              string            `yaml:"clientSecret,omitempty"`
		RedirectURI               string            `yaml:"redirectURI,omitempty"`
		BasicAuthUnsupported      string            `yaml:"basicAuthUnsupported,omitempty"`
		HostedDomains             []string          `yaml:"hostedDomains,omitempty"`
		Scopes                    []string          `yaml:"scopes,omitempty"`
		InsecureSkipEmailVerified bool              `yaml:"insecureSkipEmailVerified,omitempty"`
		InsecureEnableGroups      bool              `yaml:"insecureEnableGroups,omitempty"`
		GetUserInfo               bool              `yaml:"getUserInfo,omitempty"`
		UserIDKey                 string            `yaml:"userIDKey,omitempty"`
		UserNameKey               string            `yaml:"userNameKey,omitempty"`
		ClaimMapping              map[string]string `yaml:"claimMapping,omitempty"`
		Host                      string            `yaml:"host,omitempty"`
		InsecureSkipVerify        bool              `yaml:"insecureSkipVerify"`
		BindDN                    string            `yaml:"bindDN,omitempty"`
		BindPW                    string            `yaml:"bindPW,omitempty"`
		UsernamePrompt            string            `yaml:"usernamePrompt,omitempty"`
		InsecureNoSSL             string            `yaml:"insecureNoSSL,omitempty"`
		StartTLS                  bool              `yaml:"startTLS,omitempty"`
		RootCA                    string            `yaml:"rootCA,omitempty"`
		RootCAData                string            `yaml:"rootCAData,omitempty"`
		UserSearch                struct {
			BaseDN    string `yaml:"baseDN,omitempty"`
			Filter    string `yaml:"filter,omitempty"`
			Username  string `yaml:"username,omitempty"`
			IdAttr    string `yaml:"idAttr,omitempty"`
			EmailAttr string `yaml:"emailAttr,omitempty"`
			NameAttr  string `yaml:"nameAttr,omitempty"`
		} `yaml:"userSearch,omitempty"`
		GroupSearch struct {
			BaseDN    string `yaml:"baseDN,omitempty"`
			Filter    string `yaml:"filter,omitempty"`
			UserAttr  string `yaml:"userAttr,omitempty"`
			GroupAttr string `yaml:"groupAttr,omitempty"`
			NameAttr  string `yaml:"nameAttr,omitempty"`
		} `yaml:"groupSearch,omitempty"`
	} `yaml:"config"`
}

type oauth2 struct {
	SkipApprovalScreen bool     `yaml:"skipApprovalScreen"`
	ResponseTypes      []string `yaml:"responseTypes,omitempty"`
}

type storage struct {
	Type   string `yaml:"type,omitempty"`
	Config struct {
		InCluster bool `yaml:"inCluster"`
	} `yaml:"config"`
}
