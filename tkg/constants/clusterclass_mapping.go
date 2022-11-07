// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// ClusterAttributesToLegacyVariablesMapCommon has cluster class attributes to legacy variable names, are common for all infra providers.
var ClusterAttributesToLegacyVariablesMapCommon = map[string]string{
	"metadata.name":      ConfigVariableClusterName, // CLUSTER_NAME
	"metadata.namespace": ConfigVariableNamespace,   // NAMESPACE

	"spec.clusterNetwork.pods.cidrBlocks":     ConfigVariableClusterCIDR, // CLUSTER_CIDR
	"spec.clusterNetwork.services.cidrBlocks": ConfigVariableServiceCIDR, // SERVICE_CIDR

	"spec.topology.class":   ConfigVariableClusterClass,      // CLUSTER_CLASS
	"spec.topology.version": ConfigVariableKubernetesVersion, // KUBERNETES_VERSION

	"spec.topology.controlPlane.replicas": ConfigVariableControlPlaneMachineCount, // CONTROL_PLANE_MACHINE_COUNT

	"spec.topology.controlPlane.metadata.annotations.run.tanzu.vmware.com/resolve-os-image": "",

	"spec.topology.variables.network.ipv6Primary": TKGIPV6Primary,      // TKG_IPV6_PRIMARY
	"spec.topology.variables.proxy":               TKGHTTPProxyEnabled, // TKG_HTTP_PROXY_ENABLED
	"spec.topology.variables.proxy.httpProxy":     TKGHTTPProxy,        // TKG_HTTP_PROXY
	"spec.topology.variables.proxy.httpsProxy":    TKGHTTPSProxy,       // TKG_HTTPS_PROXY
	"spec.topology.variables.proxy.noProxy":       TKGNoProxy,          // TKG_NO_PROXY

	"spec.topology.variables.imageRepository.host":                     ConfigVariableCustomImageRepository,
	"spec.topology.variables.imageRepository.tlsCertificateValidation": ConfigVariableCustomImageRepositorySkipTLSVerify,

	"spec.topology.variables.clusterRole": ConfigVariableClusterRole, // TKG_CLUSTER_ROLE

	"spec.topology.variables.auditLogging.enabled": EnableAuditLogging, // ENABLE_AUDIT_LOGGING

	"spec.topology.variables.trust.proxy":           TKGProxyCACert,                                   // TKG_PROXY_CA_CERT
	"spec.topology.variables.trust.imageRepository": ConfigVariableCustomImageRepositoryCaCertificate, // TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE

	"spec.topology.variables.apiServerPort": ConfigVariableClusterAPIServerPort, // CLUSTER_API_SERVER_PORT
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
// spec.topology.variables.* mapped as per config_variable_association.star:get_aws_vars()
// other attributes mapped as per infrastructure-aws/v*.*.*/yttcc/overlay.yaml
var ClusterAttributesToLegacyVariablesMapAws = map[string]string{

	"spec.topology.variables.region":     ConfigVariableAWSRegion,     // AWS_REGION
	"spec.topology.variables.sshKeyName": ConfigVariableAWSSSHKeyName, // AWS_SSH_KEY_NAME

	"spec.topology.variables.loadBalancerSchemeInternal": ConfigVariableAWSLoadBalancerSchemeInternal, // AWS_LOAD_BALANCER_SCHEME_INTERNAL

	"spec.topology.variables.network.subnets.0.az":           ConfigVariableAWSNodeAz,           // AWS_NODE_AZ
	"spec.topology.variables.network.subnets.0.private.cidr": ConfigVariableAWSPrivateNodeCIDR,  // AWS_PRIVATE_NODE_CIDR
	"spec.topology.variables.network.subnets.0.private.id":   ConfigVariableAWSPrivateSubnetID,  // AWS_PRIVATE_SUBNET_ID
	"spec.topology.variables.network.subnets.0.public.cidr":  ConfigVariableAWSPublicNodeCIDR,   // AWS_PUBLIC_NODE_CIDR
	"spec.topology.variables.network.subnets.0.public.id":    ConfigVariableAWSPublicSubnetID,   // AWS_PUBLIC_SUBNET_ID
	"spec.topology.variables.network.subnets.1.az":           ConfigVariableAWSNodeAz1,          // AWS_NODE_AZ_1
	"spec.topology.variables.network.subnets.1.private.cidr": ConfigVariableAWSPrivateNodeCIDR1, // AWS_PRIVATE_NODE_CIDR_1
	"spec.topology.variables.network.subnets.1.private.id":   ConfigVariableAWSPrivateSubnetID1, // AWS_PRIVATE_SUBNET_ID_1
	"spec.topology.variables.network.subnets.1.public.cidr":  ConfigVariableAWSPublicNodeCIDR1,  // AWS_PUBLIC_NODE_CIDR_1
	"spec.topology.variables.network.subnets.1.public.id":    ConfigVariableAWSPublicSubnetID1,  // AWS_PUBLIC_SUBNET_ID_1
	"spec.topology.variables.network.subnets.2.az":           ConfigVariableAWSNodeAz2,          // AWS_NODE_AZ_2
	"spec.topology.variables.network.subnets.2.private.cidr": ConfigVariableAWSPrivateNodeCIDR2, // AWS_PRIVATE_NODE_CIDR_2
	"spec.topology.variables.network.subnets.2.private.id":   ConfigVariableAWSPrivateSubnetID2, // AWS_PRIVATE_SUBNET_ID_2
	"spec.topology.variables.network.subnets.2.public.cidr":  ConfigVariableAWSPublicNodeCIDR2,  // AWS_PUBLIC_NODE_CIDR_2
	"spec.topology.variables.network.subnets.2.public.id":    ConfigVariableAWSPublicSubnetID2,  // AWS_PUBLIC_SUBNET_ID_2

	"spec.topology.variables.network.vpc.cidr":       ConfigVariableAWSVPCCIDR, // AWS_VPC_CIDR
	"spec.topology.variables.network.vpc.existingID": ConfigVariableAWSVPCID,   // AWS_VPC_ID

	"spec.topology.variables.network.securityGroupOverrides.bastion":      ConfigVariableAWSSecurityGroupBastion,      // AWS_SECURITY_GROUP_BASTION
	"spec.topology.variables.network.securityGroupOverrides.apiServerLB":  ConfigVariableAWSSecurityGroupApiserverLb,  // AWS_SECURITY_GROUP_APISERVER_LB
	"spec.topology.variables.network.securityGroupOverrides.lb":           ConfigVariableAWSSecurityGroupLb,           // AWS_SECURITY_GROUP_LB
	"spec.topology.variables.network.securityGroupOverrides.controlPlane": ConfigVariableAWSSecurityGroupControlplane, // AWS_SECURITY_GROUP_CONTROLPLANE
	"spec.topology.variables.network.securityGroupOverrides.node":         ConfigVariableAWSSecurityGroupNode,         // AWS_SECURITY_GROUP_NODE

	"spec.topology.variables.bastion.enabled": ConfigVariableBastionHostEnabled, // BASTION_HOST_ENABLED

	"spec.topology.variables.identityRef.name": ConfigVariableAWSIdentityRefName, // AWS_IDENTITY_REF_NAME
	"spec.topology.variables.identityRef.kind": ConfigVariableAWSIdentityRefKind, // AWS_IDENTITY_REF_KIND

	"spec.topology.variables.worker.instanceType":       ConfigVariableNodeMachineType,      // NODE_MACHINE_TYPE
	"spec.topology.variables.worker.rootVolume.sizeGiB": ConfigVariableAWSNodeOsDiskSizeGib, // AWS_NODE_OS_DISK_SIZE_GIB

	"spec.topology.variables.controlPlane.instanceType":       ConfigVariableControlPlaneMachineType,      // CONTROL_PLANE_MACHINE_TYPE
	"spec.topology.variables.controlPlane.rootVolume.sizeGiB": ConfigVariableAWSControlplaneOsDiskSizeGib, // AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB

	TopologyWorkersMachineDeploymentsClass0:         "",
	TopologyWorkersMachineDeploymentsName0:          "",
	TopologyWorkersMachineDeploymentsReplicas0:      ConfigVariableWorkerMachineCount0, // WORKER_MACHINE_COUNT_0
	TopologyWorkersMachineDeploymentsFailureDomain0: "",                                // AWS_NODE_AZ_0,

	"spec.topology.workers.machineDeployments.0.metadata.annotations.run.tanzu.vmware.com/resolve-os-image": "",

	TopologyWorkersMachineDeploymentsClass1:         "",
	TopologyWorkersMachineDeploymentsName1:          "",
	TopologyWorkersMachineDeploymentsReplicas1:      ConfigVariableWorkerMachineCount1, // WORKER_MACHINE_COUNT_1
	TopologyWorkersMachineDeploymentsFailureDomain1: "",                                // AWS_NODE_AZ_1

	"spec.topology.workers.machineDeployments.1.variables.overrides.worker.instanceType": ConfigVariableNodeMachineType1, // NODE_MACHINE_TYPE_1

	TopologyWorkersMachineDeploymentsClass2:         "",
	TopologyWorkersMachineDeploymentsName2:          "",
	TopologyWorkersMachineDeploymentsReplicas2:      ConfigVariableWorkerMachineCount2, // WORKER_MACHINE_COUNT_2
	TopologyWorkersMachineDeploymentsFailureDomain2: "",                                // AWS_NODE_AZ_2

	"spec.topology.workers.machineDeployments.2.variables.overrides.worker.instanceType": ConfigVariableNodeMachineType2, // NODE_MACHINE_TYPE_2

}

// ClusterAttributesToLegacyVariablesMapAzure has, Azure Cluster object attributes path mapped to legacy variable names.
// spec.topology.variables.* mapped as per config_variable_association.star:get_azure_vars()
// other attributes mapped as per infrastructure-azure/v*.*.*/yttcc/overlay.yaml
var ClusterAttributesToLegacyVariablesMapAzure = map[string]string{

	"spec.topology.variables.network.vnet.cidrBlocks":    ConfigVariableAzureVnetCidr,          // AZURE_VNET_CIDR
	"spec.topology.variables.network.vnet.name":          ConfigVariableAzureVnetName,          // AZURE_VNET_NAME
	"spec.topology.variables.network.vnet.resourceGroup": ConfigVariableAzureVnetResourceGroup, // AZURE_VNET_RESOURCE_GROUP

	"spec.topology.variables.location":          ConfigVariableAzureLocation,          // AZURE_LOCATION
	"spec.topology.variables.resourceGroup":     ConfigVariableAzureResourceGroup,     // AZURE_RESOURCE_GROUP
	"spec.topology.variables.subscriptionID":    ConfigVariableAzureSubscriptionID,    // AZURE_SUBSCRIPTION_ID
	"spec.topology.variables.environment":       ConfigVariableAzureEnvironment,       // AZURE_ENVIRONMENT
	"spec.topology.variables.sshPublicKey":      ConfigVariableAzureSSHPublicKeyB64,   // AZURE_SSH_PUBLIC_KEY_B64
	"spec.topology.variables.frontendPrivateIP": ConfigVariableAzureFrontendPrivateIP, // AZURE_FRONTEND_PRIVATE_IP
	"spec.topology.variables.customTags":        ConfigVariableAzureCustomTags,        // AZURE_CUSTOM_TAGS

	"spec.topology.variables.acceleratedNetworking.enabled": ConfigVariableAzureEnableAcceleratedNetworking, // AZURE_ENABLE_ACCELERATED_NETWORKING
	"spec.topology.variables.privateCluster.enabled":        ConfigVariableAzureEnablePrivateCluster,        // AZURE_ENABLE_PRIVATE_CLUSTER

	"spec.topology.variables.identityRef.name":      ConfigVariableAzureIdentityName,      // AZURE_IDENTITY_NAME
	"spec.topology.variables.identityRef.namespace": ConfigVariableAzureIdentityNamespace, // AZURE_IDENTITY_NAMESPACE

	"spec.topology.variables.controlPlane.vmSize":            ConfigVariableAzureCPMachineType,               // AZURE_CONTROL_PLANE_MACHINE_TYPE
	"spec.topology.variables.controlPlane.dataDisks.sizeGiB": ConfigVariableAzureControlPlaneDataDiskSizeGib, // AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB

	"spec.topology.variables.controlPlane.osDisk.sizeGiB":            ConfigVariableAzureControlPlaneOsDiskSizeGib,            // AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB
	"spec.topology.variables.controlPlane.osDisk.storageAccountType": ConfigVariableAzureControlPlaneOsDiskStorageAccountType, // AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE

	"spec.topology.variables.controlPlane.subnet.name":          ConfigVariableAzureControlPlaneSubnetName,          // AZURE_CONTROL_PLANE_SUBNET_NAME
	"spec.topology.variables.controlPlane.subnet.cidr":          ConfigVariableAzureControlPlaneSubnetCidr,          // AZURE_CONTROL_PLANE_SUBNET_CIDR
	"spec.topology.variables.controlPlane.subnet.securityGroup": ConfigVariableAzureControlPlaneSubnetSecurityGroup, // AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP

	"spec.topology.variables.controlPlane.outboundLB.enabled":         ConfigVariableAzureControlPlaneOutboundLb,                // AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB
	"spec.topology.variables.controlPlane.outboundLB.frontendIPCount": ConfigVariableAzureControlPlaneOutboundLbFrontendIPCount, // AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT

	"spec.topology.variables.worker.vmSize":                    ConfigVariableAzureNodeMachineType,              // AZURE_NODE_MACHINE_TYPE
	"spec.topology.variables.worker.osDisk.sizeGiB":            ConfigVariableAzureNodeOsDiskSizeGib,            // AZURE_NODE_OS_DISK_SIZE_GIB
	"spec.topology.variables.worker.osDisk.storageAccountType": ConfigVariableAzureNodeOsDiskStorageAccountType, // AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE
	"spec.topology.variables.worker.dataDisks.sizeGiB":         ConfigVariableAzureNodeDataDiskSizeGib,          // AZURE_NODE_DATA_DISK_SIZE_GIB

	"spec.topology.variables.worker.subnet.cidr":          ConfigVariableAzureWorkerNodeSubnetCidr,    // AZURE_NODE_SUBNET_CIDR
	"spec.topology.variables.worker.subnet.name":          ConfigVariableAzureWorkerSubnetName,        // AZURE_NODE_SUBNET_NAME
	"spec.topology.variables.worker.subnet.securityGroup": ConfigVariableAzureNodeSubnetSecurityGroup, // AZURE_NODE_SUBNET_SECURITY_GROUP

	"spec.topology.variables.worker.outboundLB.enabled":              ConfigVariableAzureEnableNodeOutboundLb,               // AZURE_ENABLE_NODE_OUTBOUND_LB
	"spec.topology.variables.worker.outboundLB.frontendIPCount":      ConfigVariableAzureNodeOutboundLbFrontendIPCount,      // AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT
	"spec.topology.variables.worker.outboundLB.idleTimeoutInMinutes": ConfigVariableAzureNodeOutboundLbIdleTimeoutInMinutes, // AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES

	TopologyWorkersMachineDeploymentsClass0:         "",
	TopologyWorkersMachineDeploymentsName0:          "",
	TopologyWorkersMachineDeploymentsReplicas0:      ConfigVariableWorkerMachineCount0,
	TopologyWorkersMachineDeploymentsFailureDomain0: ConfigVariableAzureAZ,

	"spec.topology.workers.machineDeployments.0.metadata.annotations.run.tanzu.vmware.com/resolve-os-image": "",

	TopologyWorkersMachineDeploymentsClass1:         "",
	TopologyWorkersMachineDeploymentsName1:          "",
	TopologyWorkersMachineDeploymentsReplicas1:      ConfigVariableWorkerMachineCount1,
	TopologyWorkersMachineDeploymentsFailureDomain1: ConfigVariableAzureAZ1,

	"spec.topology.workers.machineDeployments.1.variables.overrides.worker.vmSize": ConfigVariableNodeMachineType1,

	TopologyWorkersMachineDeploymentsClass2:         "",
	TopologyWorkersMachineDeploymentsName2:          "",
	TopologyWorkersMachineDeploymentsReplicas2:      ConfigVariableWorkerMachineCount2,
	TopologyWorkersMachineDeploymentsFailureDomain2: ConfigVariableAzureAZ2,

	"spec.topology.workers.machineDeployments.2.variables.overrides.worker.vmSize": ConfigVariableNodeMachineType2,
}

// ClusterAttributesToLegacyVariablesMapVsphere has, Vsphere Cluster object attributes path mapped to legacy variable names.
// spec.topology.variables.* mapped as per config_variable_association.star:get_vsphere_vars()
// other attributes mapped as per infrastructure-vsphere/v*.*.*/yttcc/overlay.yaml
var ClusterAttributesToLegacyVariablesMapVsphere = map[string]string{

	"spec.topology.variables.apiServerEndpoint":      ConfigVariableVsphereControlPlaneEndpoint, // VSPHERE_CONTROL_PLANE_ENDPOINT
	"spec.topology.variables.vipNetworkInterface":    ConfigVariableVipNetworkInterface,         // VIP_NETWORK_INTERFACE
	"spec.topology.variables.aviAPIServerHAProvider": ConfigVariableVsphereHaProvider,           // AVI_CONTROL_PLANE_HA_PROVIDER

	"spec.topology.variables.vcenter.cloneMode":     ConfigVariableVsphereCloneMode,     // VSPHERE_CLONE_MODE
	"spec.topology.variables.vcenter.network":       ConfigVariableVsphereNetwork,       // VSPHERE_NETWORK
	"spec.topology.variables.vcenter.resourcePool":  ConfigVariableVsphereResourcePool,  // VSPHERE_RESOURCE_POOL
	"spec.topology.variables.vcenter.template":      ConfigVariableVsphereTemplate,      // VSPHERE_TEMPLATE
	"spec.topology.variables.vcenter.tlsThumbprint": ConfigVariableVsphereTLSThumbprint, // VSPHERE_TLS_THUMBPRINT
	"spec.topology.variables.vcenter.datacenter":    ConfigVariableVsphereDatacenter,    // VSPHERE_DATACENTER
	"spec.topology.variables.vcenter.datastore":     ConfigVariableVsphereDatastore,     // VSPHERE_DATASTORE
	"spec.topology.variables.vcenter.folder":        ConfigVariableVsphereFolder,        // VSPHERE_FOLDER
	"spec.topology.variables.vcenter.server":        ConfigVariableVsphereServer,        // VSPHERE_SERVER

	"spec.topology.variables.user.sshAuthorizedKeys": ConfigVariableVsphereSSHAuthorizedKey, // VSPHERE_SSH_AUTHORIZED_KEY

	"spec.topology.variables.controlPlane.machine.diskGiB":     ConfigVariableVsphereCPDiskGib,            // VSPHERE_CONTROL_PLANE_DISK_GIB
	"spec.topology.variables.controlPlane.machine.memoryMiB":   ConfigVariableVsphereCPMemMib,             // VSPHERE_CONTROL_PLANE_MEM_MIB
	"spec.topology.variables.controlPlane.machine.numCPUs":     ConfigVariableVsphereCPNumCpus,            // VSPHERE_CONTROL_PLANE_NUM_CPUS
	"spec.topology.variables.controlPlane.network.nameservers": ConfigVariableControlPlaneNodeNameservers, // CONTROL_PLANE_NODE_NAMESERVERS

	"spec.topology.variables.worker.machine.diskGiB":     ConfigVariableVsphereWorkerDiskGib,  // VSPHERE_WORKER_DISK_GIB
	"spec.topology.variables.worker.machine.memoryMiB":   ConfigVariableVsphereWorkerMemMib,   // VSPHERE_WORKER_MEM_MIB
	"spec.topology.variables.worker.machine.numCPUs":     ConfigVariableVsphereWorkerNumCpus,  // VSPHERE_WORKER_NUM_CPUS
	"spec.topology.variables.worker.network.nameservers": ConfigVariableWorkerNodeNameservers, // WORKER_NODE_NAMESERVERS

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
}

// ClusterAttributesToLegacyVariablesMapDocker has, Docker Cluster object attributes path mapped to legacy variable names.
var ClusterAttributesToLegacyVariablesMapDocker = map[string]string{}

// ClusterAttributesWithArrayTypeValue has, list of Cluster attributes paths, which value type is array list
var ClusterAttributesWithArrayTypeValue = map[string]bool{
	"spec.clusterNetwork.pods.cidrBlocks":             true,
	"spec.clusterNetwork.services.cidrBlocks":         true,
	"spec.topology.variables.proxy.noProxy":           true,
	"spec.topology.variables.user.sshAuthorizedKeys":  true,
	"spec.topology.variables.network.vnet.cidrBlocks": true,
}

// Cluster class variables constants
const (
	RegexpMachineDeploymentsOverrides = `spec.topology.workers.machineDeployments.[0-9].variables.overrides`
	RegexpTopologyClassValue          = `tkg-(aws|azure|vsphere)-default`

	TopologyVariablesNetworkSubnets   = "spec.topology.variables.network.subnets"
	TopologyVariablesNodes            = "spec.topology.variables.nodes"
	TopologyVariablesTrust            = "spec.topology.variables.trust"
	TopologyWorkersMachineDeployments = "spec.topology.workers.machineDeployments"
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

	TopologyClassIncorrectValueErrMsg                = "input cluster class file, attribute spec.topology.class has no value or incorrect value or not following correct naming convention"
	ClusterResourceWithoutTopologyNotSupportedErrMsg = "input file contains Cluster resource which doesn't have ClusterClass specified. Passing Cluster resource without ClusterClass specification is not supported"
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
