// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

type nodeType struct {
	Cpus   string
	Memory string
	Disk   string
}

// NodeTypes defines a struct of nodeType map
var NodeTypes map[string]nodeType

func init() {
	NodeTypes = make(map[string]nodeType)
	NodeTypes["small"] = nodeType{Cpus: "2", Memory: "4096", Disk: "20"}
	NodeTypes["medium"] = nodeType{Cpus: "2", Memory: "8192", Disk: "40"}
	NodeTypes["large"] = nodeType{Cpus: "4", Memory: "16384", Disk: "40"}
	NodeTypes["extra-large"] = nodeType{Cpus: "8", Memory: "32768", Disk: "80"}
}

// VSphereConfig is the tkg config file for vsphere
type VSphereConfig struct { // nolint:dupl
	ClusterName            string `yaml:"CLUSTER_NAME,omitempty"`
	InfrastructureProvider string `yaml:"INFRASTRUCTURE_PROVIDER,omitempty"`
	ClusterPlan            string `yaml:"CLUSTER_PLAN,omitempty"`
	CeipParticipation      string `yaml:"ENABLE_CEIP_PARTICIPATION,omitempty"`
	TmcRegistrationURL     string `yaml:"TMC_REGISTRATION_URL,omitempty"`

	K8sVersion                         string `yaml:"KUBERNETES_VERSION,omitempty"`
	Server                             string `yaml:"VSPHERE_SERVER,omitempty"`
	Username                           string `yaml:"VSPHERE_USERNAME,omitempty"`
	Password                           string `yaml:"VSPHERE_PASSWORD,omitempty"`
	Datacenter                         string `yaml:"VSPHERE_DATACENTER,omitempty"`
	Datastore                          string `yaml:"VSPHERE_DATASTORE,omitempty"`
	Network                            string `yaml:"VSPHERE_NETWORK,omitempty"`
	ResourcePool                       string `yaml:"VSPHERE_RESOURCE_POOL,omitempty"`
	Folder                             string `yaml:"VSPHERE_FOLDER,omitempty"`
	ControlPlaneDiskGIB                string `yaml:"VSPHERE_CONTROL_PLANE_DISK_GIB,omitempty"`
	ControlPlaneCPUs                   string `yaml:"VSPHERE_CONTROL_PLANE_NUM_CPUS,omitempty"`
	ControlPlaneMemory                 string `yaml:"VSPHERE_CONTROL_PLANE_MEM_MIB,omitempty"`
	WorkerDiskGIB                      string `yaml:"VSPHERE_WORKER_DISK_GIB,omitempty"`
	WorkerCPUs                         string `yaml:"VSPHERE_WORKER_NUM_CPUS,omitempty"`
	WorkerMemory                       string `yaml:"VSPHERE_WORKER_MEM_MIB,omitempty"`
	SSHKey                             string `yaml:"VSPHERE_SSH_AUTHORIZED_KEY,omitempty"`
	ServiceCIDR                        string `yaml:"SERVICE_CIDR,omitempty"`
	ClusterCIDR                        string `yaml:"CLUSTER_CIDR,omitempty"`
	ServiceDomain                      string `yaml:"SERVICE_DOMAIN,omitempty"`
	MachineHealthCheckEnabled          string `yaml:"ENABLE_MHC"`
	ControlPlaneEndpoint               string `yaml:"VSPHERE_CONTROL_PLANE_ENDPOINT"`
	VSphereTLSThumbprint               string `yaml:"VSPHERE_TLS_THUMBPRINT"`
	ClusterHTTPProxy                   string `yaml:"TKG_HTTP_PROXY,omitempty"`
	ClusterHTTPSProxy                  string `yaml:"TKG_HTTPS_PROXY,omitempty"`
	ClusterNoProxy                     string `yaml:"TKG_NO_PROXY,omitempty"`
	HTTPProxyEnabled                   string `yaml:"TKG_HTTP_PROXY_ENABLED"`
	AviController                      string `yaml:"AVI_CONTROLLER"`
	AviUsername                        string `yaml:"AVI_USERNAME"`
	AviPassword                        string `yaml:"AVI_PASSWORD"`
	AviCloudName                       string `yaml:"AVI_CLOUD_NAME"`
	AviServiceEngine                   string `yaml:"AVI_SERVICE_ENGINE_GROUP"`
	AviDataNetwork                     string `yaml:"AVI_DATA_NETWORK"`
	AviDataNetworkCIDR                 string `yaml:"AVI_DATA_NETWORK_CIDR"`
	AviCAData                          string `yaml:"AVI_CA_DATA_B64"`
	AviLabels                          string `yaml:"AVI_LABELS"`
	AviEnable                          string `yaml:"AVI_ENABLE"`
	EnableAuditLogging                 string `yaml:"ENABLE_AUDIT_LOGGING"`
	AviControlPlaneEndpointProvider    string `yaml:"AVI_CONTROL_PLANE_HA_PROVIDER"`
	AviManagementClusterVipNetworkName string `yaml:"AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_NAME"`
	AviManagementClusterVipNetworkCidr string `yaml:"AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_CIDR"`
	IDPConfig                          `yaml:",inline"`
	OsInfo                             `yaml:",inline"`
}

// NewVSphereConfig generates TKG config for vsphere
func (c *client) NewVSphereConfig(params *models.VsphereRegionalClusterParams) (*VSphereConfig, error) { //nolint:funlen,gocyclo
	var err error
	res := &VSphereConfig{
		ClusterName:            params.ClusterName,
		InfrastructureProvider: constants.InfrastructureProviderVSphere,
		ClusterPlan:            params.ControlPlaneFlavor,
		TmcRegistrationURL:     params.TmcRegistrationURL,

		Datacenter:           params.Datacenter,
		Datastore:            params.Datastore,
		Folder:               params.Folder,
		SSHKey:               params.SSHKey,
		ControlPlaneEndpoint: params.ControlPlaneEndpoint,
		HTTPProxyEnabled:     falseConst,
	}
	if params.Os != nil {
		if params.Os.OsInfo != nil {
			res.OsInfo.Name = params.Os.OsInfo.Name
			res.OsInfo.Version = params.Os.OsInfo.Version
			res.OsInfo.Arch = params.Os.OsInfo.Arch
		}
		c.tkgConfigReaderWriter.Set(constants.ConfigVariableVsphereTemplate, params.Os.Name)
	}

	if params.CeipOptIn != nil {
		res.CeipParticipation = strconv.FormatBool(*params.CeipOptIn)
	}

	res.EnableAuditLogging = falseConst
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

	res.ResourcePool = params.ResourcePool

	if params.VsphereCredentials != nil {
		res.Server = params.VsphereCredentials.Host
		res.Username = params.VsphereCredentials.Username
		res.Password = params.VsphereCredentials.Password
		res.VSphereTLSThumbprint = params.VsphereCredentials.Thumbprint
	}

	if params.Networking != nil {
		res.ServiceCIDR = params.Networking.ClusterServiceCIDR
		res.ClusterCIDR = params.Networking.ClusterPodCIDR
		res.Network = params.Networking.NetworkName

		if params.Networking.HTTPProxyConfiguration != nil && params.Networking.HTTPProxyConfiguration.Enabled {
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
	}

	cpNodeSize, ok := NodeTypes[params.ControlPlaneNodeType]
	if !ok {
		return res, errors.Errorf("control plane node size %s not defined", params.ControlPlaneNodeType)
	}
	res.ControlPlaneCPUs = cpNodeSize.Cpus
	res.ControlPlaneMemory = cpNodeSize.Memory
	res.ControlPlaneDiskGIB = cpNodeSize.Disk

	workerNodeSize, ok := NodeTypes[params.WorkerNodeType]
	if !ok {
		return res, errors.Errorf("worker node size %s not defined", params.WorkerNodeType)
	}
	res.WorkerCPUs = workerNodeSize.Cpus
	res.WorkerMemory = workerNodeSize.Memory
	res.WorkerDiskGIB = workerNodeSize.Disk

	if params.MachineHealthCheckEnabled {
		res.MachineHealthCheckEnabled = trueConst
	} else {
		res.MachineHealthCheckEnabled = falseConst
	}

	res.AviEnable = falseConst
	if isAviEnabled(params) {
		res.AviController = params.AviConfig.Controller
		res.AviUsername = params.AviConfig.Username
		res.AviPassword = params.AviConfig.Password
		res.AviCAData = base64.StdEncoding.EncodeToString([]byte(params.AviConfig.CaCert))
		res.AviServiceEngine = params.AviConfig.ServiceEngine
		res.AviCloudName = params.AviConfig.Cloud
		res.AviDataNetwork = params.AviConfig.Network.Name
		res.AviDataNetworkCIDR = params.AviConfig.Network.Cidr
		res.AviLabels = mapToYamlStr(params.AviConfig.Labels)
		res.AviEnable = trueConst

		res.AviControlPlaneEndpointProvider = falseConst
		if params.AviConfig.ControlPlaneHaProvider {
			res.AviControlPlaneEndpointProvider = trueConst
		}
		res.AviManagementClusterVipNetworkName = params.AviConfig.ManagementClusterVipNetworkName
		res.AviManagementClusterVipNetworkCidr = params.AviConfig.ManagementClusterVipNetworkCidr
	}

	return res, nil
}

// GetVsphereNodeSizeOptions returns the list of vSphere node size options
func GetVsphereNodeSizeOptions() string {
	nodeTypes := []string{}

	for k := range NodeTypes {
		nodeTypes = append(nodeTypes, k)
	}
	sort.Strings(nodeTypes)
	return strings.Join(nodeTypes, ",")
}

// CheckAndGetProxyURL validates and returns the proxy URL
func CheckAndGetProxyURL(username, password, proxyURL string) (string, error) {
	httpURL, err := url.Parse(proxyURL)
	if err != nil {
		return "", err
	}

	if httpURL.Scheme == "" {
		return "", errors.New("scheme is missing from the proxy URL")
	}

	if httpURL.Host == "" {
		return "", errors.New("hostname is missing from the proxy URL")
	}

	if username != "" && password != "" {
		httpURL.User = url.UserPassword(username, password)
	} else if username != "" {
		httpURL.User = url.User(username)
	}

	return httpURL.String(), nil
}

func mapToYamlStr(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	metadataStr := ""
	for key, value := range m {
		metadataStr += fmt.Sprintf("'%s': '%s'\n", key, value)
	}
	return metadataStr
}

func isAviEnabled(params *models.VsphereRegionalClusterParams) bool {
	return params.AviConfig != nil && params.AviConfig.Cloud != "" && params.AviConfig.ServiceEngine != ""
}
