// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"

	"github.com/go-openapi/strfmt"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

func createIdentityManagementConfig(config interface{}) *models.IdentityManagementConfig {
	if getFieldFromConfig(config, "IdentityManagementType") == "" {
		return nil
	}

	idmType := getFieldFromConfig(config, "IdentityManagementType")
	ldapRootCa, _ := base64.StdEncoding.DecodeString(getFieldFromConfig(config, "LDAPRootCAData"))

	return &models.IdentityManagementConfig{
		IdmType:                  &idmType,
		LdapBindDn:               getFieldFromConfig(config, "LDAPBindDN"),
		LdapBindPassword:         getFieldFromConfig(config, "LDAPBindPassword"),
		LdapGroupSearchBaseDn:    getFieldFromConfig(config, "LDAPGroupSearchBaseDN"),
		LdapGroupSearchFilter:    getFieldFromConfig(config, "LDAPGroupSearchFilter"),
		LdapGroupSearchGroupAttr: getFieldFromConfig(config, "LDAPGroupSearchGroupAttr"),
		LdapGroupSearchNameAttr:  getFieldFromConfig(config, "LDAPGroupSearchNameAttr"),
		LdapGroupSearchUserAttr:  getFieldFromConfig(config, "LDAPGroupSearchUserAttr"),
		LdapRootCa:               string(ldapRootCa),
		LdapURL:                  getFieldFromConfig(config, "LDAPHost"),
		LdapUserSearchBaseDn:     getFieldFromConfig(config, "LDAPUserSearchBaseDN"),
		LdapUserSearchEmailAttr:  "",
		LdapUserSearchFilter:     getFieldFromConfig(config, "LDAPUserSearchFilter"),
		LdapUserSearchIDAttr:     "",
		LdapUserSearchNameAttr:   getFieldFromConfig(config, "LDAPUserSearchNameAttr"),
		LdapUserSearchUsername:   getFieldFromConfig(config, "LDAPUserSearchUsername"),
		OidcClaimMappings: map[string]string{
			"groups":   getFieldFromConfig(config, "OIDCGroupsClaim"),
			"username": getFieldFromConfig(config, "OIDCUsernameClaim"),
		},
		OidcClientID:       getFieldFromConfig(config, "OIDCClientID"),
		OidcClientSecret:   getFieldFromConfig(config, "OIDCClientSecret"),
		OidcProviderName:   getFieldFromConfig(config, "OIDCProviderName"),
		OidcProviderURL:    strfmt.URI(getFieldFromConfig(config, "OIDCIssuerURL")),
		OidcScope:          getFieldFromConfig(config, "OIDCScopes"),
		OidcSkipVerifyCert: false,
	}
}

func createNetworkingConfig(config interface{}) *models.TKGNetwork {
	if getFieldFromConfig(config, "Network") == "" {
		return nil
	}

	return &models.TKGNetwork{
		ClusterDNSName:         "",
		ClusterNodeCIDR:        "",
		ClusterPodCIDR:         getFieldFromConfig(config, "ClusterCIDR"),
		ClusterServiceCIDR:     getFieldFromConfig(config, "ServiceCIDR"),
		CniType:                "",
		HTTPProxyConfiguration: createHTTPProxyConfig(config),
		NetworkName:            "",
	}
}

func createHTTPProxyConfig(config interface{}) *models.HTTPProxyConfiguration {
	var httpProxyConfig *models.HTTPProxyConfiguration
	if getFieldFromConfig(config, "HTTPProxyEnabled") == trueConst {
		httpURL, _ := url.Parse(getFieldFromConfig(config, "ClusterHTTPProxy"))
		httpPassword, _ := httpURL.User.Password()
		httpsURL, _ := url.Parse(getFieldFromConfig(config, "ClusterHTTPSProxy"))
		httpsPassword, _ := httpsURL.User.Password()

		httpProxyConfig = &models.HTTPProxyConfiguration{
			HTTPProxyPassword:  httpPassword,
			HTTPProxyURL:       httpURL.Scheme + "://" + httpURL.Hostname() + httpURL.RequestURI(),
			HTTPProxyUsername:  httpURL.User.Username(),
			HTTPSProxyPassword: httpsPassword,
			HTTPSProxyURL:      httpsURL.Scheme + "://" + httpsURL.Hostname() + httpsURL.RequestURI(),
			HTTPSProxyUsername: httpsURL.User.Username(),
			Enabled:            true,
			NoProxy:            getFieldFromConfig(config, "ClusterNoProxy"),
		}
	}
	return httpProxyConfig
}

func getFieldFromConfig(config interface{}, fieldName string) string {
	field := reflect.ValueOf(config).FieldByName(fieldName)
	if !field.IsValid() {
		fmt.Printf("getFieldFromConfig() is unable to find field %s in object %v\n", fieldName, config)
		return ""
	}
	return field.String()
}
