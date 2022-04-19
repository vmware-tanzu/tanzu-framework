// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// ClusterAttributesToLegacyVariablesMapCommon has cluster class attributes to legacy variable names, are common for all infra providers.
var ClusterAttributesToLegacyVariablesMapCommon = map[string]string{
	"metadata.name":      ConfigVariableClusterName,
	"metadata.namespace": ConfigVariableNamespace,

	"spec.clusterNetwork.apiServerPort":       ConfigVariableClusterAPIServerPort,
	"spec.clusterNetwork.pods.cidrBlocks":     ConfigVariableClusterCIDR,
	"spec.clusterNetwork.services.cidrBlocks": ConfigVariableServiceCIDR,

	"spec.topology.class":                 ConfigVariableClusterClass,
	"spec.topology.version":               ConfigVariableKubernetesVersion,
	"spec.topology.controlPlane.replicas": ConfigVariableControlPlaneMachineCount,

	"spec.topology.variables.tkgClusterRole":              ConfigVariableClusterRole,
	"spec.topology.variables.clusterName":                 ConfigVariableClusterName,
	"spec.topology.variables.clusterPlan":                 ConfigVariableClusterPlan,
	"spec.topology.variables.tkgCustomImage.repository":   ConfigVariableCustomImageRepository,
	"spec.topology.variables.tkgCustomImage.caCert":       ConfigVariableCustomImageRepositoryCaCertificate,
	"spec.topology.variables.tkgCustomImage.skpTlsVerify": ConfigVariableCustomImageRepositorySkipTLSVerify,
	"spec.topology.variables.proxy":                       TKGHTTPProxyEnabled,
	"spec.topology.variables.proxy.httpProxy":             TKGHTTPProxy,
	"spec.topology.variables.proxy.httpsProxy":            TKGHTTPSProxy,
	"spec.topology.variables.proxy.noProxy":               TKGNoProxy,
	"spec.topology.variables.proxy.caCert":                TKGProxyCACert,
}

// ClusterAttributesHigherPrecedenceToLowerMap has cluster class input attributes which
// has higher precedence (as key's in this map) to attribute with lower precedence (as value's in this map)
// For some properties (NODE_MACHINE_TYPE_1, NODE_MACHINE_TYPE_2, etc), Cluster YAML object may have values twice, this map holds attributes paths, key attribute path has higher
// precedence than the value attribute path. So if the Cluster object has the values for both key and value attribute paths,
// then we need to consider key attribute path (which has higher precedence).
// eg: Key: "spec.topology.workers.machineDeployments.1.variables.overrides.NODE_MACHINE_TYPE", value: "spec.topology.variables.nodes.1.machineType"
// 	these two key/value attribute paths mapped to NODE_MACHINE_TYPE_1 legacy variable, if these two attribute paths has values in Cluster Object
// 	then need to consider higher precedence attribute path value which is key of ClusterAttributesHigherPrecedenceToLowerMap.

var ClusterAttributesHigherPrecedenceToLowerMap = map[string]string{
	"spec.topology.workers.machineDeployments.1.variables.overrides.NODE_MACHINE_TYPE": "spec.topology.variables.nodes.1.machineType",
	"spec.topology.workers.machineDeployments.2.variables.overrides.NODE_MACHINE_TYPE": "spec.topology.variables.nodes.2.machineType",
}

// ClusterAttributesToLegacyVariablesMapAws has, AWS Cluster object attributes path mapped to legacy variable names.
var ClusterAttributesToLegacyVariablesMapAws = map[string]string{

	TopologyWorkersMachineDeploymentsClass0:         "",
	TopologyWorkersMachineDeploymentsName0:          "",
	TopologyWorkersMachineDeploymentsReplicas0:      ConfigVariableWorkerMachineCount,
	TopologyWorkersMachineDeploymentsFailureDomain0: ConfigVariableAWSNodeAz,

	TopologyWorkersMachineDeploymentsClass1:                                            "",
	TopologyWorkersMachineDeploymentsName1:                                             "",
	TopologyWorkersMachineDeploymentsReplicas1:                                         ConfigVariableWorkerMachineCount1,
	TopologyWorkersMachineDeploymentsFailureDomain1:                                    ConfigVariableAWSNodeAz1,
	"spec.topology.workers.machineDeployments.1.variables.overrides.NODE_MACHINE_TYPE": ConfigVariableNodeMachineType1,

	TopologyWorkersMachineDeploymentsClass2:                                            "",
	TopologyWorkersMachineDeploymentsName2:                                             "",
	TopologyWorkersMachineDeploymentsReplicas2:                                         ConfigVariableWorkerMachineCount2,
	TopologyWorkersMachineDeploymentsFailureDomain2:                                    ConfigVariableAWSNodeAz2,
	"spec.topology.workers.machineDeployments.2.variables.overrides.NODE_MACHINE_TYPE": ConfigVariableNodeMachineType2,

	// spec.topology.variables.* mapped as per config_variable_association.star:get_aws_vars()
	"spec.topology.variables.region":                     ConfigVariableAWSRegion,
	"spec.topology.variables.sshKeyName":                 ConfigVariableAWSSSHKeyName,
	"spec.topology.variables.bastionHostEnabled":         ConfigVariableBastionHostEnabled,
	"spec.topology.variables.loadBalancerSchemeInternal": ConfigVariableAWSLoadBalancerSchemeInternal,
	"spec.topology.variables.vpc.cidr":                   ConfigVariableAWSVPCCIDR,
	"spec.topology.variables.vpc.id":                     ConfigVariableAWSVPCID,
	"spec.topology.variables.identityRef.kind":           ConfigVariableAWSIdentityRefKind,
	"spec.topology.variables.identityRef.name":           ConfigVariableAWSIdentityRefName,
	"spec.topology.variables.securityGroup.node":         ConfigVariableAWSSecurityGroupNode,
	"spec.topology.variables.securityGroup.apiServerLB":  ConfigVariableAWSSecurityGroupApiserverLb,
	"spec.topology.variables.securityGroup.bastion":      ConfigVariableAWSSecurityGroupBastion,
	"spec.topology.variables.securityGroup.controlPlane": ConfigVariableAWSSecurityGroupControlplane,
	"spec.topology.variables.securityGroup.lb":           ConfigVariableAWSSecurityGroupLb,

	"spec.topology.variables.subnets.0.private.cidr":      ConfigVariableAWSPrivateNodeCIDR,
	"spec.topology.variables.subnets.0.private.id":        ConfigVariableAWSPrivateSubnetID,
	"spec.topology.variables.subnets.0.public.cidr":       ConfigVariableAWSPublicNodeCIDR,
	"spec.topology.variables.subnets.0.public.id":         ConfigVariableAWSPublicSubnetID,
	"spec.topology.variables.subnets.1.private.cidr":      ConfigVariableAWSPrivateNodeCIDR1,
	"spec.topology.variables.subnets.1.private.id":        ConfigVariableAWSPrivateSubnetID1,
	"spec.topology.variables.subnets.1.public.cidr":       ConfigVariableAWSPublicNodeCIDR1,
	"spec.topology.variables.subnets.1.public.id":         ConfigVariableAWSPublicSubnetID1,
	"spec.topology.variables.subnets.2.private.cidr":      ConfigVariableAWSPrivateNodeCIDR2,
	"spec.topology.variables.subnets.2.private.id":        ConfigVariableAWSPrivateSubnetID2,
	"spec.topology.variables.subnets.2.public.cidr":       ConfigVariableAWSPublicNodeCIDR2,
	"spec.topology.variables.subnets.2.public.id":         ConfigVariableAWSPublicSubnetID2,
	"spec.topology.variables.controlPlane.machineType":    ConfigVariableCPMachineType,
	"spec.topology.variables.controlPlane.osDisk.sizeGiB": ConfigVariableAWSControlplaneOsDiskSizeGib,

	"spec.topology.variables.nodes.0.az":             ConfigVariableAWSNodeAz,
	"spec.topology.variables.nodes.0.machineType":    ConfigVariableNodeMachineType,
	"spec.topology.variables.nodes.0.osDisk.sizeGiB": ConfigVariableAWSNodeOsDiskSizeGib,
	"spec.topology.variables.nodes.1.az":             ConfigVariableAWSNodeAz1,
	"spec.topology.variables.nodes.1.machineType":    ConfigVariableNodeMachineType1,
	"spec.topology.variables.nodes.1.osDisk.sizeGiB": ConfigVariableAWSNodeOsDiskSizeGib,
	"spec.topology.variables.nodes.2.az":             ConfigVariableAWSNodeAz2,
	"spec.topology.variables.nodes.2.machineType":    ConfigVariableNodeMachineType2,
	"spec.topology.variables.nodes.2.osDisk.sizeGiB": ConfigVariableAWSNodeOsDiskSizeGib,
}

// ClusterAttributesToLegacyVariablesMapAzure has, Azure Cluster object attributes path mapped to legacy variable names.
var ClusterAttributesToLegacyVariablesMapAzure = map[string]string{

	TopologyWorkersMachineDeploymentsClass0:         "",
	TopologyWorkersMachineDeploymentsName0:          "",
	TopologyWorkersMachineDeploymentsReplicas0:      ConfigVariableWorkerMachineCount,
	TopologyWorkersMachineDeploymentsFailureDomain0: ConfigVariableAzureAZ,

	TopologyWorkersMachineDeploymentsClass1:         "",
	TopologyWorkersMachineDeploymentsName1:          "",
	TopologyWorkersMachineDeploymentsReplicas1:      ConfigVariableWorkerMachineCount1,
	TopologyWorkersMachineDeploymentsFailureDomain1: ConfigVariableAzureAZ1,

	TopologyWorkersMachineDeploymentsClass2:         "",
	TopologyWorkersMachineDeploymentsName2:          "",
	TopologyWorkersMachineDeploymentsReplicas2:      ConfigVariableWorkerMachineCount2,
	TopologyWorkersMachineDeploymentsFailureDomain2: ConfigVariableAzureAZ2,

	// spec.topology.variables.* mapped as per config_variable_association.star:get_azure_vars()
	"spec.topology.variables.location":                    ConfigVariableAzureLocation,
	"spec.topology.variables.resourceGroup":               ConfigVariableAzureResourceGroup,
	"spec.topology.variables.subscriptionID":              ConfigVariableAzureSubscriptionID,
	"spec.topology.variables.environment":                 ConfigVariableAzureEnvironment,
	"spec.topology.variables.sshPublicKey":                ConfigVariableAzureSSHPublicKeyB64,
	"spec.topology.variables.enableAcceleratedNetworking": ConfigVariableAzureEnableAcceleratedNetworking,
	"spec.topology.variables.enablePrivateCluster":        ConfigVariableAzureEnablePrivateCluster,
	"spec.topology.variables.frontendPrivateIP":           ConfigVariableAzureFrontendPrivateIP,
	"spec.topology.variables.customTags":                  ConfigVariableAzureCustomTags,

	"spec.topology.variables.vnet.cidr":          ConfigVariableAzureVnetCidr,
	"spec.topology.variables.vnet.name":          ConfigVariableAzureVnetName,
	"spec.topology.variables.vnet.resourceGroup": ConfigVariableAzureVnetResourceGroup,

	"spec.topology.variables.identity.name":      ConfigVariableAzureIdentityName,
	"spec.topology.variables.identity.namespace": ConfigVariableAzureIdentityNamespace,

	"spec.topology.variables.controlPlane.dataDisk.sizeGiB":           ConfigVariableAzureControlPlaneDataDiskSizeGib,
	"spec.topology.variables.controlPlane.machineType":                ConfigVariableAzureCPMachineType,
	"spec.topology.variables.controlPlane.osDisk.storageAccountType":  ConfigVariableAzureControlPlaneOsDiskStorageAccountType,
	"spec.topology.variables.controlPlane.osDisk.sizeGiB":             ConfigVariableAzureControlPlaneOsDiskSizeGib,
	"spec.topology.variables.controlPlane.outboundLB.frontendIPCount": ConfigVariableAzureControlPlaneOutboundLbFrontendIPCount,
	"spec.topology.variables.controlPlane.outboundLB.enabled":         ConfigVariableAzureControlPlaneOutboundLb,
	"spec.topology.variables.controlPlane.subnet.name":                ConfigVariableAzureControlPlaneSubnetName,
	"spec.topology.variables.controlPlane.subnet.securityGroup":       ConfigVariableAzureControlPlaneSubnetSecurityGroup,
	"spec.topology.variables.controlPlane.subnet.cidr":                ConfigVariableAzureControlPlaneSubnetCidr,

	"spec.topology.variables.node.machineType":                     ConfigVariableAzureNodeMachineType,
	"spec.topology.variables.node.osDisk.sizeGiB":                  ConfigVariableAzureNodeOsDiskSizeGib,
	"spec.topology.variables.node.osDisk.storageAccountType":       ConfigVariableAzureNodeOsDiskStorageAccountType,
	"spec.topology.variables.node.dataDisk.enabled":                ConfigVariableAzureEnableNodeDataDisk,
	"spec.topology.variables.node.dataDisk.sizeGiB":                ConfigVariableAzureNodeDataDiskSizeGib,
	"spec.topology.variables.node.subnet.name":                     ConfigVariableAzureWorkerSubnetName,
	"spec.topology.variables.node.subnet.cidr":                     ConfigVariableAzureWorkerNodeSubnetCidr,
	"spec.topology.variables.node.subnet.securityGroup":            ConfigVariableAzureNodeSubnetSecurityGroup,
	"spec.topology.variables.node.outboundLB.enabled":              ConfigVariableAzureEnableNodeOutboundLb,
	"spec.topology.variables.node.outboundLB.frontendIPCount":      ConfigVariableAzureNodeOutboundLbFrontendIPCount,
	"spec.topology.variables.node.outboundLB.idleTimeoutInMinutes": ConfigVariableAzureNodeOutboundLbIdleTimeoutInMinutes,
}

// ClusterAttributesToLegacyVariablesMapVsphere has, Vsphere Cluster object attributes path mapped to legacy variable names.
var ClusterAttributesToLegacyVariablesMapVsphere = map[string]string{

	TopologyWorkersMachineDeploymentsClass0:         "",
	TopologyWorkersMachineDeploymentsName0:          "",
	TopologyWorkersMachineDeploymentsReplicas0:      ConfigVariableWorkerMachineCount,
	TopologyWorkersMachineDeploymentsFailureDomain0: ConfigVariableVsphereAz0,

	TopologyWorkersMachineDeploymentsClass1:         "",
	TopologyWorkersMachineDeploymentsName1:          "",
	TopologyWorkersMachineDeploymentsReplicas1:      ConfigVariableWorkerMachineCount1,
	TopologyWorkersMachineDeploymentsFailureDomain1: ConfigVariableVsphereAz1,

	TopologyWorkersMachineDeploymentsClass2:         "",
	TopologyWorkersMachineDeploymentsName2:          "",
	TopologyWorkersMachineDeploymentsReplicas2:      ConfigVariableWorkerMachineCount2,
	TopologyWorkersMachineDeploymentsFailureDomain2: ConfigVariableVsphereAz2,

	// spec.topology.variables.* mapped as per config_variable_association.star:get_vsphere_vars()
	"spec.topology.variables.cloneMode":                 ConfigVariableVsphereCloneMode,
	"spec.topology.variables.network":                   ConfigVariableVsphereNetwork,
	"spec.topology.variables.datacenter":                ConfigVariableVsphereDatacenter,
	"spec.topology.variables.datastore":                 ConfigVariableVsphereDatastore,
	"spec.topology.variables.folder":                    ConfigVariableVsphereFolder,
	"spec.topology.variables.resourcePool":              ConfigVariableVsphereResourcePool,
	"spec.topology.variables.storagePolicyID":           ConfigVariableVsphereStoragePolicyID,
	"spec.topology.variables.server":                    ConfigVariableVsphereServer,
	"spec.topology.variables.tlsThumbprint":             ConfigVariableVsphereTLSThumbprint,
	"spec.topology.variables.template":                  ConfigVariableVsphereTemplate,
	"spec.topology.variables.controlPlaneEndpoint":      ConfigVariableVsphereControlPlaneEndpoint,
	"spec.topology.variables.vipNetworkInterface":       ConfigVariableVipNetworkInterface,
	"spec.topology.variables.clusterApiServerPort":      ConfigVariableClusterAPIServerPort,
	"spec.topology.variables.aviControlPlaneHAProvider": ConfigVariableVsphereHaProvider,
	"spec.topology.variables.serviceCidr":               ConfigVariableServiceCIDR,
	"spec.topology.variables.clusterCidr":               ConfigVariableClusterCIDR,

	"spec.topology.variables.controlPlane.count":               ConfigVariableControlPlaneMachineCount,
	"spec.topology.variables.controlPlane.machine.numCPUs":     ConfigVariableVsphereCPNumCpus,
	"spec.topology.variables.controlPlane.machine.diskGiB":     ConfigVariableVsphereCPDiskGib,
	"spec.topology.variables.controlPlane.machine.memoryMiB":   ConfigVariableVsphereCPMemMib,
	"spec.topology.variables.controlPlane.network.nameservers": ConfigVariableControlPlaneNodeNameservers,

	"spec.topology.variables.node.count":               "",
	"spec.topology.variables.node.machine.numCPUs":     ConfigVariableVsphereWorkerNumCpus,
	"spec.topology.variables.node.machine.diskGiB":     ConfigVariableVsphereWorkerDiskGib,
	"spec.topology.variables.node.machine.memoryMiB":   ConfigVariableVsphereWorkerMemMib,
	"spec.topology.variables.node.network.nameservers": ConfigVariableWorkerNodeNameservers,
}

// ClusterAttributesToLegacyVariablesMapDocker has, Docker Cluster object attributes path mapped to legacy variable names.
var ClusterAttributesToLegacyVariablesMapDocker = map[string]string{}

// Cluster class variables constants
const (
	TopologyVariablesSubnets          = "spec.topology.variables.subnets"
	TopologyVariablesNodes            = "spec.topology.variables.nodes"
	TopologyWorkersMachineDeployments = "spec.topology.workers.machineDeployments"
	RegexpMachineDeploymentsOverrides = `spec.topology.workers.machineDeployments.[0-9].variables.overrides`
	TopologyClass                     = "spec.topology.class"
	TopologyVariables                 = "spec.topology.variables"

	TopologyWorkersMachineDeploymentsClass0         = "spec.topology.workers.machineDeployments.0.class"
	TopologyWorkersMachineDeploymentsName0          = "spec.topology.workers.machineDeployments.0.name"
	TopologyWorkersMachineDeploymentsReplicas0      = "spec.topology.workers.machineDeployments.0.replicas"
	TopologyWorkersMachineDeploymentsFailureDomain0 = "spec.topology.workers.machineDeployments.0.failureDomain"

	TopologyWorkersMachineDeploymentsClass1         = "spec.topology.workers.machineDeployments.1.class"
	TopologyWorkersMachineDeploymentsName1          = "spec.topology.workers.machineDeployments.1.name"
	TopologyWorkersMachineDeploymentsReplicas1      = "spec.topology.workers.machineDeployments.1.replicas"
	TopologyWorkersMachineDeploymentsFailureDomain1 = "spec.topology.workers.machineDeployments.1.failureDomain"

	TopologyWorkersMachineDeploymentsClass2         = "spec.topology.workers.machineDeployments.2.class"
	TopologyWorkersMachineDeploymentsName2          = "spec.topology.workers.machineDeployments.2.name"
	TopologyWorkersMachineDeploymentsReplicas2      = "spec.topology.workers.machineDeployments.2.replicas"
	TopologyWorkersMachineDeploymentsFailureDomain2 = "spec.topology.workers.machineDeployments.2.failureDomain"

	SPEC = "spec"

	TopologyClassIncorrectValueErrMsg = "input cluster class file, attribute spec.topology.class has no value or incorrect value or not following correct naming convension"
)

// InfrastructureSpecificVariableMappingMap has, infra name to variable mapping map, which makes easy to get infra specific mapping map
var InfrastructureSpecificVariableMappingMap = map[string]map[string]string{
	InfrastructureProviderVSphere: ClusterAttributesToLegacyVariablesMapVsphere,
	InfrastructureProviderAWS:     ClusterAttributesToLegacyVariablesMapAws,
	InfrastructureProviderAzure:   ClusterAttributesToLegacyVariablesMapAzure,
	InfrastructureProviderDocker:  ClusterAttributesToLegacyVariablesMapDocker,
}

// During initialization, combine common map to infra specific map's, which reduces code where we use infra maps
func init() {
	for key, value := range ClusterAttributesToLegacyVariablesMapCommon {
		ClusterAttributesToLegacyVariablesMapAws[key] = value
		ClusterAttributesToLegacyVariablesMapAzure[key] = value
		ClusterAttributesToLegacyVariablesMapVsphere[key] = value
		ClusterAttributesToLegacyVariablesMapDocker[key] = value
	}
}
