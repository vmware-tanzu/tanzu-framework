// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package schemas defines the YAML schemas for various objects.
package schemas

// DexConfig contains the Dex configuration settings.
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
	EnablePasswordDB bool            `yaml:"enablePasswordDB"`
}

type frontEnd struct {
	Theme string `yaml:"theme,omitempty"`
}

type web struct {
	HTTPS   string `yaml:"https,omitempty"`
	TLSCert string `yaml:"tlsCert,omitempty"`
	TLSKey  string `yaml:"tlsKey,omitempty"`
}

type expiry struct {
	SigningKeys    string `yaml:"signingKeys,omitempty"`
	IDTokens       string `yaml:"idTokens,omitempty"`
	AuthRequests   string `yaml:"authRequests,omitempty"`
	DeviceRequests string `yaml:"deviceRequests,omitempty"`
}

type logger struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
}

// StaticClient contains client information.
type StaticClient struct {
	ID           string   `yaml:"id,omitempty"`
	Name         string   `yaml:"name,omitempty"`
	RedirectURIs []string `yaml:"redirectURIs,omitempty"`
	Secret       string   `yaml:"secret,omitempty"`
}

type connector struct {
	Type   string `yaml:"type,omitempty"`
	ID     string `yaml:"id,omitempty"`
	Name   string `yaml:"name,omitempty"`
	Config struct {
		UserSearch struct {
			BaseDN    string `yaml:"baseDN,omitempty"`
			Filter    string `yaml:"filter,omitempty"`
			Username  string `yaml:"username,omitempty"`
			IDAttr    string `yaml:"idAttr,omitempty"`
			EmailAttr string `yaml:"emailAttr,omitempty"`
			NameAttr  string `yaml:"nameAttr,omitempty"`
			Scope     string `yaml:"scope,omitempty"`
		} `yaml:"userSearch,omitempty"`
		GroupSearch struct {
			BaseDN       string `yaml:"baseDN,omitempty"`
			Filter       string `yaml:"filter,omitempty"`
			NameAttr     string `yaml:"nameAttr,omitempty"`
			Scope        string `yaml:"scope,omitempty"`
			UserMatchers []struct {
				UserAttr  string `yaml:"userAttr,omitempty"`
				GroupAttr string `yaml:"groupAttr,omitempty"`
			} `yaml:"userMatchers,omitempty"`
		} `yaml:"groupSearch,omitempty"`
		HostedDomains             []string          `yaml:"hostedDomains,omitempty"`
		Scopes                    []string          `yaml:"scopes,omitempty"`
		Issuer                    string            `yaml:"issuer,omitempty"`
		ClientID                  string            `yaml:"clientID,omitempty"`
		ClientSecret              string            `yaml:"clientSecret,omitempty"`
		RedirectURI               string            `yaml:"redirectURI,omitempty"`
		BasicAuthUnsupported      string            `yaml:"basicAuthUnsupported,omitempty"`
		UserIDKey                 string            `yaml:"userIDKey,omitempty"`
		UserNameKey               string            `yaml:"userNameKey,omitempty"`
		Host                      string            `yaml:"host,omitempty"`
		BindDN                    string            `yaml:"bindDN,omitempty"`
		BindPW                    string            `yaml:"bindPW,omitempty"`
		UsernamePrompt            string            `yaml:"usernamePrompt,omitempty"`
		RootCA                    string            `yaml:"rootCA,omitempty"`
		RootCAData                string            `yaml:"rootCAData,omitempty"`
		ClaimMapping              map[string]string `yaml:"claimMapping,omitempty"`
		InsecureSkipEmailVerified bool              `yaml:"insecureSkipEmailVerified,omitempty"`
		InsecureEnableGroups      bool              `yaml:"insecureEnableGroups,omitempty"`
		GetUserInfo               bool              `yaml:"getUserInfo,omitempty"`
		InsecureSkipVerify        bool              `yaml:"insecureSkipVerify"`
		InsecureNoSSL             bool              `yaml:"insecureNoSSL,omitempty"`
		StartTLS                  bool              `yaml:"startTLS,omitempty"`
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
