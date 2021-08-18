// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"encoding/base64"
	"strconv"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	utils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

const (
	trueConst  = "true"
	falseConst = "false"
	azCount    = 3
)

// AWSConfig is the tkg config for aws
type AWSConfig struct {
	AccessKeyID           string `yaml:"AWS_ACCESS_KEY_ID,omitempty"`
	AMIID                 string `yaml:"AWS_AMI_ID,omitempty"`
	AWSPrivateSubnetID    string `yaml:"AWS_PRIVATE_SUBNET_ID"`
	AWSPrivateSubnetID2   string `yaml:"AWS_PRIVATE_SUBNET_ID_1"`
	AWSPrivateSubnetID3   string `yaml:"AWS_PRIVATE_SUBNET_ID_2"`
	AWSPublicSubnetID     string `yaml:"AWS_PUBLIC_SUBNET_ID"`
	AWSPublicSubnetID2    string `yaml:"AWS_PUBLIC_SUBNET_ID_1"`
	AWSPublicSubnetID3    string `yaml:"AWS_PUBLIC_SUBNET_ID_2"`
	AWSVPCID              string `yaml:"AWS_VPC_ID"`
	B64EncodedCredentials string `yaml:"AWS_B64ENCODED_CREDENTIALS,omitempty"`
	BastionHostEnabled    string `yaml:"BASTION_HOST_ENABLED"`
	CeipParticipation     string `yaml:"ENABLE_CEIP_PARTICIPATION,omitempty"`
	ClusterCidr           string `yaml:"CLUSTER_CIDR"`
	ClusterHTTPProxy      string `yaml:"TKG_HTTP_PROXY,omitempty"`
	ClusterHTTPSProxy     string `yaml:"TKG_HTTPS_PROXY,omitempty"`
	ClusterName           string `yaml:"CLUSTER_NAME,omitempty"`
	ClusterNoProxy        string `yaml:"TKG_NO_PROXY,omitempty"`
	ClusterPlan           string `yaml:"CLUSTER_PLAN,omitempty"`
	ControlPlaneNodeType  string `yaml:"CONTROL_PLANE_MACHINE_TYPE,omitempty"`
	// ControlPlaneOSDiskSizeGiB is the size of the root volume of the control plane instances of a cluster
	ControlPlaneOSDiskSizeGiB string `yaml:"AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB,omitempty"`
	CredentialProfile         string `yaml:"AWS_PROFILE,omitempty"`
	EnableAuditLogging        string `yaml:"ENABLE_AUDIT_LOGGING"`
	HTTPProxyEnabled          string `yaml:"TKG_HTTP_PROXY_ENABLED"`
	InfrastructureProvider    string `yaml:"INFRASTRUCTURE_PROVIDER,omitempty"`
	MachineHealthCheckEnabled string `yaml:"ENABLE_MHC"`
	Node2Az                   string `yaml:"AWS_NODE_AZ_1"`
	Node3Az                   string `yaml:"AWS_NODE_AZ_2"`
	NodeAz                    string `yaml:"AWS_NODE_AZ"`
	NodeMachineType           string `yaml:"NODE_MACHINE_TYPE,omitempty"`
	NodeMachineType1          string `yaml:"NODE_MACHINE_TYPE_1,omitempty"`
	NodeMachineType2          string `yaml:"NODE_MACHINE_TYPE_2,omitempty"`
	// NodeOSDiskSizeGiB is the size of the root volume of the node instances of a cluster
	NodeOSDiskSizeGiB      string                    `yaml:"AWS_NODE_OS_DISK_SIZE_GIB,omitempty"`
	PrivateNode2Cidr       string                    `yaml:"AWS_PRIVATE_NODE_CIDR_1"`
	PrivateNode3Cidr       string                    `yaml:"AWS_PRIVATE_NODE_CIDR_2"`
	PrivateNodeCidr        string                    `yaml:"AWS_PRIVATE_NODE_CIDR"`
	PublicNode2Cidr        string                    `yaml:"AWS_PUBLIC_NODE_CIDR_1"`
	PublicNode3Cidr        string                    `yaml:"AWS_PUBLIC_NODE_CIDR_2"`
	PublicNodeCidr         string                    `yaml:"AWS_PUBLIC_NODE_CIDR"`
	Region                 string                    `yaml:"AWS_REGION,omitempty"`
	SecretAcessKey         string                    `yaml:"AWS_SECRET_ACCESS_KEY,omitempty"`
	ServiceCidr            string                    `yaml:"SERVICE_CIDR"`
	SessionToken           string                    `yaml:"AWS_SESSION_TOKEN,omitempty"`
	SSHKeyName             string                    `yaml:"AWS_SSH_KEY_NAME"`
	TmcRegistrationURL     string                    `yaml:"TMC_REGISTRATION_URL,omitempty"`
	VPCCidr                string                    `yaml:"AWS_VPC_CIDR"`
	IdentityReference      AWSIdentityReference      `yaml:",inline"`
	SecurityGroupOverrides AWSSecurityGroupOverrides `yaml:",inline"`
	IDPConfig              `yaml:",inline"`
	OsInfo                 `yaml:",inline"`
}

// AWSIdentityReference defines an optional reference to a AWS Identity Reference resource.
type AWSIdentityReference struct {
	// Kind is an optional kind of a Kubernetes resource containing  an identity to be used for a cluster.
	// Defaults to AWSClusterRoleIdentity if Name is set
	Kind string `yaml:"AWS_IDENTITY_REF_KIND,omitempty"`
	// Name is an optional name of a Kubernetes resource containing an identity to be used for a cluster.
	Name string `yaml:"AWS_IDENTITY_REF_NAME,omitempty"`
}

// AWSSecurityGroupOverrides can be used in conjunction with Bring Your Own Infrastructure to define specific security group
// IDs to use for the cluster
type AWSSecurityGroupOverrides struct {
	// APIServerLoadBalancer is an optional security group ID of a pre-created security group that will be used for Kubernetes
	// API Server ELB, and will control inbound access to the the control plane endpoint
	APIServerLoadBalancer string `yaml:"AWS_SECURITY_GROUP_APISERVER_LB,omitempty"`
	// Bastion is an optional security group ID of a pre-created security group that will be used to control in-bound access
	// to the bastion
	Bastion string `yaml:"AWS_SECURITY_GROUP_BASTION,omitempty"`
	// ControlPlane is an optional security group ID of a pre-created security group that will be used to control in-bound
	// access to the control plane nodes
	ControlPlane string `yaml:"AWS_SECURITY_GROUP_CONTROLPLANE,omitempty"`
	// CloudProviderLoadBalancer is an optional security group ID for use by the Kubernetes AWS Cloud Provider for setting rules
	// for ELBs
	CloudProviderLoadBalancer string `yaml:"AWS_SECURITY_GROUP_LB,omitempty"`
	// Node is an optional security group ID that will be used to to control in-bound acceess to all nodes
	Node string `yaml:"AWS_SECURITY_GROUP_NODE,omitempty"`
}

type newSubnetPair struct {
	Az                string
	PrivateSubnetCidr string
	PublicSubnetCidr  string
}

// K8sVersionAMIMap represents map of k8s version to aws AMI ID to use for that k8s version
type K8sVersionAMIMap map[string]string

// NewAWSConfig generates TKG config for aws
func (c *client) NewAWSConfig(params *models.AWSRegionalClusterParams, encodedCredentials string) (*AWSConfig, error) { //nolint:funlen,gocyclo
	bomConfiguration, err := c.tkgBomClient.GetDefaultTkrBOMConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get default TKr BoM configuration")
	}
	amiID := getAMIId(bomConfiguration, params)
	if amiID == "" {
		return nil, errors.Errorf("No AMI found in region %s for TKr version %s", params.AwsAccountParams.Region, bomConfiguration.Release.Version)
	}

	nodeMachineType1 := ""
	nodeMachineType2 := ""

	if len(params.Vpc.Azs) == 0 {
		return nil, errors.New("AWS node availability zone cannot be empty")
	}

	if params.ControlPlaneFlavor == constants.PlanProd && len(params.Vpc.Azs) < 3 {
		return nil, errors.Errorf("number of Availability Zones less than 3 for production cluster, actual %d", len(params.Vpc.Azs))
	}

	if params.ControlPlaneFlavor == constants.PlanProd {
		nodeMachineType1 = params.Vpc.Azs[1].WorkerNodeType
		nodeMachineType2 = params.Vpc.Azs[2].WorkerNodeType
	}

	res := &AWSConfig{
		ClusterName:            params.ClusterName,
		InfrastructureProvider: constants.InfrastructureProviderAWS,
		ClusterPlan:            params.ControlPlaneFlavor,
		TmcRegistrationURL:     params.TmcRegistrationURL,
		Region:                 params.AwsAccountParams.Region,
		B64EncodedCredentials:  encodedCredentials,
		ControlPlaneNodeType:   params.ControlPlaneNodeType,
		NodeMachineType:        params.Vpc.Azs[0].WorkerNodeType,
		NodeMachineType1:       nodeMachineType1,
		NodeMachineType2:       nodeMachineType2,
		AMIID:                  amiID,
		SSHKeyName:             params.SSHKeyName,
		ClusterCidr:            params.Networking.ClusterPodCIDR,
		ServiceCidr:            params.Networking.ClusterServiceCIDR,
		HTTPProxyEnabled:       falseConst,
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

	if params.AwsAccountParams.AccessKeyID != "" && params.AwsAccountParams.SecretAccessKey != "" {
		res.AccessKeyID = params.AwsAccountParams.AccessKeyID
		res.SecretAcessKey = params.AwsAccountParams.SecretAccessKey
		res.SessionToken = params.AwsAccountParams.SessionToken
	} else if params.AwsAccountParams.ProfileName != "" {
		res.CredentialProfile = params.AwsAccountParams.ProfileName
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

	if len(params.Vpc.Azs) == 0 {
		return res, errors.New("AWS node availability zone cannot be empty")
	}

	if params.ControlPlaneFlavor == constants.PlanProd && len(params.Vpc.Azs) < 3 {
		return nil, errors.Errorf("number of Availability Zones less than 3 for production cluster, actual %d", len(params.Vpc.Azs))
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

	if params.Vpc.VpcID == "" {
		// Create new VPC
		res.VPCCidr = params.Vpc.Cidr

		newSubnets, err := appendSubnets(params.Vpc, params.ControlPlaneFlavor == constants.PlanProd)
		if err != nil {
			return nil, errors.Wrap(err, "unable to divide cidrs for new VPC's subnets")
		}
		res.NodeAz = newSubnets[0].Az
		res.PrivateNodeCidr = newSubnets[0].PrivateSubnetCidr
		res.PublicNodeCidr = newSubnets[0].PublicSubnetCidr

		if params.ControlPlaneFlavor == constants.PlanProd {
			res.Node2Az = newSubnets[1].Az
			res.PrivateNode2Cidr = newSubnets[1].PrivateSubnetCidr
			res.PublicNode2Cidr = newSubnets[1].PublicSubnetCidr

			res.Node3Az = newSubnets[2].Az
			res.PrivateNode3Cidr = newSubnets[2].PrivateSubnetCidr
			res.PublicNode3Cidr = newSubnets[2].PublicSubnetCidr
		}
	} else {
		// Use existing VPC
		res.AWSVPCID = params.Vpc.VpcID

		res.NodeAz = params.Vpc.Azs[0].Name
		res.AWSPublicSubnetID = params.Vpc.Azs[0].PublicSubnetID
		res.AWSPrivateSubnetID = params.Vpc.Azs[0].PrivateSubnetID

		if params.ControlPlaneFlavor == constants.PlanProd {
			res.Node2Az = params.Vpc.Azs[1].Name
			res.AWSPublicSubnetID2 = params.Vpc.Azs[1].PublicSubnetID
			res.AWSPrivateSubnetID2 = params.Vpc.Azs[1].PrivateSubnetID

			res.Node3Az = params.Vpc.Azs[2].Name
			res.AWSPublicSubnetID3 = params.Vpc.Azs[2].PublicSubnetID
			res.AWSPrivateSubnetID3 = params.Vpc.Azs[2].PrivateSubnetID
		}
	}

	if params.BastionHostEnabled {
		res.BastionHostEnabled = trueConst
	} else {
		res.BastionHostEnabled = falseConst
	}

	if params.MachineHealthCheckEnabled {
		res.MachineHealthCheckEnabled = trueConst
	} else {
		res.MachineHealthCheckEnabled = falseConst
	}

	return res, nil
}

func getAMIId(bomConfiguration *tkgconfigbom.BOMConfiguration, params *models.AWSRegionalClusterParams) string {
	amiID := ""
	if amis, ok := bomConfiguration.AMI[params.AwsAccountParams.Region]; ok {
		if params.Os != nil && params.Os.OsInfo != nil {
			for _, ami := range amis {
				if ami.OSInfo.Name == params.Os.OsInfo.Name {
					amiID = ami.ID
					break
				}
			}
		} else {
			amiID = amis[0].ID
		}
	}

	return amiID
}

// AppendSubnets append subnet information in providerConfig to paramsVpc
func appendSubnets(paramsVpc *models.AWSVpc, highAvailability bool) ([]*newSubnetPair, error) {
	ProdSubnetCount := 6
	ExtendedBits := 4
	NumberOfSubnetsForDivProd := 8
	NumberOfSubnetsForDivDev := 2

	azNames := make([]string, 0)
	for i := range paramsVpc.Azs {
		azNames = append(azNames, paramsVpc.Azs[i].Name)
	}

	newSubnets := make([]*newSubnetPair, 0)

	if highAvailability {
		if len(azNames) < azCount {
			return nil, errors.Errorf("number of Availability Zones less than 3 for production cluster, actual %d", len(azNames))
		}
		subnetCidrs, err := utils.DivideVPCCidr(paramsVpc.Cidr, ExtendedBits, NumberOfSubnetsForDivProd)
		if err != nil {
			return nil, err
		}
		for i := 1; i < ProdSubnetCount; i += 2 {
			newSubnets = append(newSubnets, &newSubnetPair{
				PublicSubnetCidr:  subnetCidrs[i-1],
				PrivateSubnetCidr: subnetCidrs[i],
				Az:                azNames[i/2],
			})
		}
	} else {
		if len(azNames) != 1 {
			return nil, errors.Errorf("number of Availability Zones not 1 for developer cluster, actual %d", len(azNames))
		}
		subnetCidrs, err := utils.DivideVPCCidr(paramsVpc.Cidr, ExtendedBits, NumberOfSubnetsForDivDev)
		if err != nil {
			return nil, err
		}
		newSubnets = append(newSubnets, &newSubnetPair{
			PublicSubnetCidr:  subnetCidrs[0],
			PrivateSubnetCidr: subnetCidrs[1],
			Az:                azNames[0],
		})
	}

	return newSubnets, nil
}
