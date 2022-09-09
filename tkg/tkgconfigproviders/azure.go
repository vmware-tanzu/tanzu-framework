// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"encoding/base64"
	"strconv"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// AzureConfig is the tkg config for Azure
type AzureConfig struct {
	ClusterName               string `yaml:"CLUSTER_NAME,omitempty"`
	ClusterLabels             string `yaml:"CLUSTER_LABELS,omitempty"`
	ClusterAnnotations        string `yaml:"CLUSTER_ANNOTATIONS,omitempty"`
	InfrastructureProvider    string `yaml:"INFRASTRUCTURE_PROVIDER,omitempty"`
	ClusterPlan               string `yaml:"CLUSTER_PLAN,omitempty"`
	CeipParticipation         string `yaml:"ENABLE_CEIP_PARTICIPATION,omitempty"`
	Region                    string `yaml:"AZURE_LOCATION,omitempty"`
	SubscriptionID            string `yaml:"AZURE_SUBSCRIPTION_ID,omitempty"`
	Environment               string `yaml:"AZURE_ENVIRONMENT,omitempty"`
	TenantID                  string `yaml:"AZURE_TENANT_ID,omitempty"`
	ClientID                  string `yaml:"AZURE_CLIENT_ID,omitempty"`
	ClientSecret              string `yaml:"AZURE_CLIENT_SECRET,omitempty"`
	SSHKeyB64                 string `yaml:"AZURE_SSH_PUBLIC_KEY_B64,omitempty"`
	ControlPlaneMachineType   string `yaml:"AZURE_CONTROL_PLANE_MACHINE_TYPE,omitempty"`
	NodeMachineType           string `yaml:"AZURE_NODE_MACHINE_TYPE,omitempty"`
	ResourceGroup             string `yaml:"AZURE_RESOURCE_GROUP,omitempty"`
	VNetResourceGroup         string `yaml:"AZURE_VNET_RESOURCE_GROUP,omitempty"`
	VNetName                  string `yaml:"AZURE_VNET_NAME,omitempty"`
	ControlPlaneSubnet        string `yaml:"AZURE_CONTROL_PLANE_SUBNET_NAME,omitempty"`
	WorkerNodeSubnet          string `yaml:"AZURE_NODE_SUBNET_NAME,omitempty"`
	VNetCIDR                  string `yaml:"AZURE_VNET_CIDR,omitempty"`
	ControlPlaneSubnetCIDR    string `yaml:"AZURE_CONTROL_PLANE_SUBNET_CIDR,omitempty"`
	ControlPlaneSubnetSG      string `yaml:"AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP,omitempty"`
	WorkerNodeSubnetCIDR      string `yaml:"AZURE_NODE_SUBNET_CIDR,omitempty"`
	WorkerNodeSubnetSG        string `yaml:"AZURE_NODE_SUBNET_SECURITY_GROUP,omitempty"`
	MachineHealthCheckEnabled string `yaml:"ENABLE_MHC,omitempty"`
	EnableAuditLogging        string `yaml:"ENABLE_AUDIT_LOGGING"`
	ServiceCIDR               string `yaml:"SERVICE_CIDR,omitempty"`
	ClusterCIDR               string `yaml:"CLUSTER_CIDR,omitempty"`
	ClusterHTTPProxy          string `yaml:"TKG_HTTP_PROXY,omitempty"`
	ClusterHTTPSProxy         string `yaml:"TKG_HTTPS_PROXY,omitempty"`
	ClusterNoProxy            string `yaml:"TKG_NO_PROXY,omitempty"`
	HTTPProxyEnabled          string `yaml:"TKG_HTTP_PROXY_ENABLED"`
	EnablePrivateCluster      string `yaml:"AZURE_ENABLE_PRIVATE_CLUSTER"`
	FrontendPrivateIP         string `yaml:"AZURE_FRONTEND_PRIVATE_IP"`
	IDPConfig                 `yaml:",inline"`
	OsInfo                    `yaml:",inline"`
}

// NewAzureConfig generates TKG config for Azure
func (c *client) NewAzureConfig(params *models.AzureRegionalClusterParams) (*AzureConfig, error) { // nolint:funlen
	var err error
	res := &AzureConfig{
		ClusterName:             params.ClusterName,
		ClusterLabels:           mapToConfigString(params.Labels),
		ClusterAnnotations:      mapToConfigString(params.Annotations),
		InfrastructureProvider:  constants.InfrastructureProviderAzure,
		ClusterPlan:             params.ControlPlaneFlavor,
		Region:                  params.Location,
		Environment:             params.AzureAccountParams.AzureCloud,
		SubscriptionID:          params.AzureAccountParams.SubscriptionID,
		TenantID:                params.AzureAccountParams.TenantID,
		ClientID:                params.AzureAccountParams.ClientID,
		ClientSecret:            params.AzureAccountParams.ClientSecret,
		SSHKeyB64:               base64.StdEncoding.EncodeToString([]byte(params.SSHPublicKey)),
		ControlPlaneMachineType: params.ControlPlaneMachineType,
		NodeMachineType:         params.WorkerMachineType,
		ClusterCIDR:             params.Networking.ClusterPodCIDR,
		ServiceCIDR:             params.Networking.ClusterServiceCIDR,
		HTTPProxyEnabled:        falseConst,
	}

	if params.CeipOptIn != nil {
		res.CeipParticipation = strconv.FormatBool(*params.CeipOptIn)
	}

	if params.Os != nil && params.Os.OsInfo != nil {
		res.OsInfo.Name = params.Os.OsInfo.Name
		res.OsInfo.Version = params.Os.OsInfo.Version
		res.OsInfo.Arch = params.Os.OsInfo.Arch
	}

	if params.EnableAuditLogging {
		res.EnableAuditLogging = trueConst
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

	res.ResourceGroup = params.ResourceGroup
	res.VNetResourceGroup = params.VnetResourceGroup
	res.VNetName = params.VnetName
	res.ControlPlaneSubnet = params.ControlPlaneSubnet
	res.WorkerNodeSubnet = params.WorkerNodeSubnet

	if params.VnetCidr != "" { // create new vnet
		res.VNetCIDR = params.VnetCidr
		res.ControlPlaneSubnetCIDR = params.ControlPlaneSubnetCidr
		res.WorkerNodeSubnetCIDR = params.WorkerNodeSubnetCidr
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

	if params.IsPrivateCluster {
		res.EnablePrivateCluster = trueConst
		res.FrontendPrivateIP = params.FrontendPrivateIP
	}

	return res, nil
}

// CreateAzureParams generates a Params object from an AzureConfig, used for importing configuration files
func (c *client) CreateAzureParams(azureConfig *AzureConfig) (params *models.AzureRegionalClusterParams, err error) {
	ceipOptIn := azureConfig.CeipParticipation == trueConst
	sshKey, err := base64.StdEncoding.DecodeString(azureConfig.SSHKeyB64)
	if err != nil {
		return nil, err
	}

	return &models.AzureRegionalClusterParams{
		Annotations: configStringToMap(azureConfig.ClusterAnnotations),
		AzureAccountParams: &models.AzureAccountParams{
			AzureCloud:     "",
			ClientID:       azureConfig.ClientID,
			ClientSecret:   azureConfig.ClientSecret,
			SubscriptionID: azureConfig.SubscriptionID,
			TenantID:       azureConfig.TenantID,
		},
		CeipOptIn:               &ceipOptIn,
		ClusterName:             azureConfig.ClusterName,
		ControlPlaneFlavor:      azureConfig.ClusterPlan,
		ControlPlaneMachineType: azureConfig.ControlPlaneMachineType,
		ControlPlaneSubnet:      azureConfig.ControlPlaneSubnet,
		ControlPlaneSubnetCidr:  azureConfig.ControlPlaneSubnetCIDR,
		EnableAuditLogging:      azureConfig.EnableAuditLogging == trueConst,
		FrontendPrivateIP:       azureConfig.FrontendPrivateIP,
		IdentityManagement:      createIdentityManagementConfig(azureConfig),
		IsPrivateCluster:        azureConfig.EnablePrivateCluster == trueConst,
		//KubernetesVersion:         "",
		Labels:                    configStringToMap(azureConfig.ClusterLabels),
		Location:                  azureConfig.Region,
		MachineHealthCheckEnabled: azureConfig.MachineHealthCheckEnabled == trueConst,
		Networking:                createNetworkingConfig(azureConfig),
		NumOfWorkerNodes:          "",
		Os:                        createOsInfo(azureConfig),
		ResourceGroup:             azureConfig.ResourceGroup,
		SSHPublicKey:              string(sshKey),
		VnetCidr:                  azureConfig.VNetCIDR,
		VnetName:                  azureConfig.VNetName,
		VnetResourceGroup:         azureConfig.VNetResourceGroup,
		WorkerMachineType:         azureConfig.NodeMachineType,
		WorkerNodeSubnet:          azureConfig.WorkerNodeSubnet,
		WorkerNodeSubnetCidr:      azureConfig.WorkerNodeSubnetCIDR,
	}, nil
}

func createOsInfo(azureConfig *AzureConfig) *models.AzureVirtualMachine {
	return &models.AzureVirtualMachine{
		Name: "",
		OsInfo: &models.OSInfo{
			Arch:    azureConfig.OsInfo.Arch,
			Name:    azureConfig.OsInfo.Name,
			Version: azureConfig.OsInfo.Version,
		},
	}
}
