// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

// IDPConfig struct defining properties for identity provider configuration
type IDPConfig struct {
	IdentityManagementType string `yaml:"IDENTITY_MANAGEMENT_TYPE"`
	OIDCConfig             `yaml:",inline"`
	LDAPConfig             `yaml:",inline"`
}

// OIDCConfig struct defining properties for OIDC configuration
type OIDCConfig struct {
	OIDCProviderName  string `yaml:"OIDC_IDENTITY_PROVIDER_NAME"`
	OIDCIssuerURL     string `yaml:"OIDC_IDENTITY_PROVIDER_ISSUER_URL"`
	OIDCClientID      string `yaml:"OIDC_IDENTITY_PROVIDER_CLIENT_ID"`
	OIDCClientSecret  string `yaml:"OIDC_IDENTITY_PROVIDER_CLIENT_SECRET"`
	OIDCScopes        string `yaml:"OIDC_IDENTITY_PROVIDER_SCOPES"`
	OIDCGroupsClaim   string `yaml:"OIDC_IDENTITY_PROVIDER_GROUPS_CLAIM"`
	OIDCUsernameClaim string `yaml:"OIDC_IDENTITY_PROVIDER_USERNAME_CLAIM"`
}

// LDAPConfig struct defining properties for OIDC configuration
type LDAPConfig struct {
	LDAPBindDN               string `yaml:"LDAP_BIND_DN"`
	LDAPBindPassword         string `yaml:"LDAP_BIND_PASSWORD"`
	LDAPHost                 string `yaml:"LDAP_HOST"`
	LDAPUserSearchBaseDN     string `yaml:"LDAP_USER_SEARCH_BASE_DN"`
	LDAPUserSearchFilter     string `yaml:"LDAP_USER_SEARCH_FILTER"`
	LDAPUserSearchUsername   string `yaml:"LDAP_USER_SEARCH_USERNAME"`
	LDAPUserSearchNameAttr   string `yaml:"LDAP_USER_SEARCH_NAME_ATTRIBUTE"`
	LDAPGroupSearchBaseDN    string `yaml:"LDAP_GROUP_SEARCH_BASE_DN"`
	LDAPGroupSearchFilter    string `yaml:"LDAP_GROUP_SEARCH_FILTER"`
	LDAPGroupSearchUserAttr  string `yaml:"LDAP_GROUP_SEARCH_USER_ATTRIBUTE"`
	LDAPGroupSearchGroupAttr string `yaml:"LDAP_GROUP_SEARCH_GROUP_ATTRIBUTE"`
	LDAPGroupSearchNameAttr  string `yaml:"LDAP_GROUP_SEARCH_NAME_ATTRIBUTE"`
	LDAPRootCAData           string `yaml:"LDAP_ROOT_CA_DATA_B64"`
}
