// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"encoding/base64"
	"strconv"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// DockerConfig is the tkg config file for docker provider
type DockerConfig struct {
	ClusterName               string `yaml:"CLUSTER_NAME,omitempty"`
	ClusterLabels             string `yaml:"CLUSTER_LABELS,omitempty"`
	ClusterAnnotations        string `yaml:"CLUSTER_ANNOTATIONS,omitempty"`
	InfrastructureProvider    string `yaml:"INFRASTRUCTURE_PROVIDER,omitempty"`
	ClusterPlan               string `yaml:"CLUSTER_PLAN,omitempty"`
	CeipParticipation         string `yaml:"ENABLE_CEIP_PARTICIPATION,omitempty"`
	MachineHealthCheckEnabled string `yaml:"ENABLE_MHC,omitempty"`
	ServiceCIDR               string `yaml:"SERVICE_CIDR,omitempty"`
	ClusterCIDR               string `yaml:"CLUSTER_CIDR,omitempty"`
	ClusterHTTPProxy          string `yaml:"TKG_HTTP_PROXY,omitempty"`
	ClusterHTTPSProxy         string `yaml:"TKG_HTTPS_PROXY,omitempty"`
	ClusterNoProxy            string `yaml:"TKG_NO_PROXY,omitempty"`
	HTTPProxyEnabled          string `yaml:"TKG_HTTP_PROXY_ENABLED"`
	IDPConfig                 `yaml:",inline"`
	OsInfo                    `yaml:",inline"`
}

func (c *client) NewDockerConfig(params *models.DockerRegionalClusterParams) (*DockerConfig, error) {
	var err error
	res := &DockerConfig{
		ClusterName:            params.ClusterName,
		ClusterLabels:          mapToConfigString(params.Labels),
		ClusterAnnotations:     mapToConfigString(params.Annotations),
		InfrastructureProvider: constants.InfrastructureProviderDocker,
		ClusterPlan:            constants.PlanDev,
		ClusterCIDR:            params.Networking.ClusterPodCIDR,
		ServiceCIDR:            params.Networking.ClusterServiceCIDR,
		HTTPProxyEnabled:       falseConst,
	}

	if params.CeipOptIn != nil {
		res.CeipParticipation = strconv.FormatBool(*params.CeipOptIn)
	}

	if params.IdentityManagement != nil { //nolint:dupl
		res.IdentityManagementType = *params.IdentityManagement.IdmType
		res.OIDCProviderName = params.IdentityManagement.OidcProviderName
		res.OIDCIssuerURL = params.IdentityManagement.OidcProviderURL.String()
		res.OIDCClientID = params.IdentityManagement.OidcClientID
		res.OIDCClientSecret = params.IdentityManagement.OidcClientSecret
		res.OIDCScopes = params.IdentityManagement.OidcScope
		res.OIDCGroupsClaim = params.IdentityManagement.OidcClaimMappings["groups"]
		res.OIDCUsernameClaim = params.IdentityManagement.OidcClaimMappings["username"]
		res.LDAPBindDN = params.IdentityManagement.LdapBindDn
		res.LDAPBindPassword = params.IdentityManagement.LdapBindPassword
		res.LDAPHost = params.IdentityManagement.LdapURL
		res.LDAPUserSearchBaseDN = params.IdentityManagement.LdapUserSearchBaseDn
		res.LDAPUserSearchFilter = params.IdentityManagement.LdapUserSearchFilter
		res.LDAPUserSearchUsername = params.IdentityManagement.LdapUserSearchUsername
		res.LDAPUserSearchNameAttr = params.IdentityManagement.LdapUserSearchNameAttr
		res.LDAPGroupSearchBaseDN = params.IdentityManagement.LdapGroupSearchBaseDn
		res.LDAPGroupSearchFilter = params.IdentityManagement.LdapGroupSearchFilter
		res.LDAPGroupSearchUserAttr = params.IdentityManagement.LdapGroupSearchUserAttr
		res.LDAPGroupSearchGroupAttr = params.IdentityManagement.LdapGroupSearchGroupAttr
		res.LDAPGroupSearchNameAttr = params.IdentityManagement.LdapGroupSearchNameAttr
		res.LDAPRootCAData = base64.StdEncoding.EncodeToString([]byte(params.IdentityManagement.LdapRootCa))
	}

	if params.Networking != nil && params.Networking.HTTPProxyConfiguration != nil && params.Networking.HTTPProxyConfiguration.Enabled {
		res.HTTPProxyEnabled = trueConst
		conf := params.Networking.HTTPProxyConfiguration
		res.ClusterHTTPProxy, err = CheckAndGetProxyURL(conf.HTTPProxyUsername, conf.HTTPProxyPassword, conf.HTTPProxyURL)
		if err != nil {
			return res, err
		}
		res.ClusterHTTPSProxy, err = CheckAndGetProxyURL(conf.HTTPSProxyUsername, conf.HTTPSProxyPassword, conf.HTTPSProxyURL)
		if err != nil {
			return res, err
		}
		res.ClusterNoProxy = params.Networking.HTTPProxyConfiguration.NoProxy
	}

	if params.MachineHealthCheckEnabled {
		res.MachineHealthCheckEnabled = trueConst
	} else {
		res.MachineHealthCheckEnabled = falseConst
	}

	return res, nil
}

// CreateDockerParams generates a Params object from a DockerConfig, used for importing configuration files
func (c *client) CreateDockerParams(dockerConfig *DockerConfig) (params *models.DockerRegionalClusterParams, err error) {
	ceipOptIn := dockerConfig.CeipParticipation == trueConst

	params = &models.DockerRegionalClusterParams{
		Annotations:               configStringToMap(dockerConfig.ClusterAnnotations),
		ClusterName:               dockerConfig.ClusterName,
		Networking:                createDockerNetworkingConfig(dockerConfig),
		CeipOptIn:                 &ceipOptIn,
		ControlPlaneFlavor:        "",
		IdentityManagement:        createIdentityManagementConfig(dockerConfig),
		KubernetesVersion:         "",
		Labels:                    configStringToMap(dockerConfig.ClusterLabels),
		MachineHealthCheckEnabled: dockerConfig.MachineHealthCheckEnabled == trueConst,
		NumOfWorkerNodes:          "",
	}

	return params, nil
}

// createDockerNetworkingConfig() creates a TKGNetwork from a docker config. Note that we need a special method here,
// because the other providers have a Networking object that they use within their xxxConfig object,
// but Docker just has the fields at the DockerConfig level
func createDockerNetworkingConfig(conf *DockerConfig) *models.TKGNetwork {
	return &models.TKGNetwork{
		ClusterDNSName:         "",
		ClusterNodeCIDR:        "",
		ClusterPodCIDR:         conf.ClusterCIDR,
		ClusterServiceCIDR:     conf.ServiceCIDR,
		CniType:                "",
		HTTPProxyConfiguration: createHTTPProxyConfig(conf),
		NetworkName:            "",
	}
}
