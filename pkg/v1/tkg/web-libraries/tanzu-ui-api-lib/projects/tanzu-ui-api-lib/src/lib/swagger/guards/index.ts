/* tslint:disable */

import * as models from '../models';

/* pre-prepared guards for build in complex types */

function _isBlob(arg: any): arg is Blob {
  return arg != null && typeof arg.size === 'number' && typeof arg.type === 'string' && typeof arg.slice === 'function';
}

export function isFile(arg: any): arg is File {
return arg != null && typeof arg.lastModified === 'number' && typeof arg.name === 'string' && _isBlob(arg);
}

/* generated type guards */

export function isAviCloud(arg: any): arg is models.AviCloud {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // location?: string
    ( typeof arg.location === 'undefined' || typeof arg.location === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // uuid?: string
    ( typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string' ) &&

  true
  );
  }

export function isAviConfig(arg: any): arg is models.AviConfig {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // ca_cert?: string
    ( typeof arg.ca_cert === 'undefined' || typeof arg.ca_cert === 'string' ) &&
    // cloud?: string
    ( typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string' ) &&
    // controller?: string
    ( typeof arg.controller === 'undefined' || typeof arg.controller === 'string' ) &&
    // controlPlaneHaProvider?: boolean
    ( typeof arg.controlPlaneHaProvider === 'undefined' || typeof arg.controlPlaneHaProvider === 'boolean' ) &&
    // labels?: { [key: string]: string }
    ( typeof arg.labels === 'undefined' || typeof arg.labels === 'string' ) &&
    // managementClusterVipNetworkCidr?: string
    ( typeof arg.managementClusterVipNetworkCidr === 'undefined' || typeof arg.managementClusterVipNetworkCidr === 'string' ) &&
    // managementClusterVipNetworkName?: string
    ( typeof arg.managementClusterVipNetworkName === 'undefined' || typeof arg.managementClusterVipNetworkName === 'string' ) &&
    // network?: AviNetworkParams
    ( typeof arg.network === 'undefined' || isAviNetworkParams(arg.network) ) &&
    // password?: string
    ( typeof arg.password === 'undefined' || typeof arg.password === 'string' ) &&
    // service_engine?: string
    ( typeof arg.service_engine === 'undefined' || typeof arg.service_engine === 'string' ) &&
    // username?: string
    ( typeof arg.username === 'undefined' || typeof arg.username === 'string' ) &&

  true
  );
  }

export function isAviControllerParams(arg: any): arg is models.AviControllerParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // CAData?: string
    ( typeof arg.CAData === 'undefined' || typeof arg.CAData === 'string' ) &&
    // host?: string
    ( typeof arg.host === 'undefined' || typeof arg.host === 'string' ) &&
    // password?: string
    ( typeof arg.password === 'undefined' || typeof arg.password === 'string' ) &&
    // tenant?: string
    ( typeof arg.tenant === 'undefined' || typeof arg.tenant === 'string' ) &&
    // username?: string
    ( typeof arg.username === 'undefined' || typeof arg.username === 'string' ) &&

  true
  );
  }

export function isAviNetworkParams(arg: any): arg is models.AviNetworkParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // cidr?: string
    ( typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isAviServiceEngineGroup(arg: any): arg is models.AviServiceEngineGroup {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // location?: string
    ( typeof arg.location === 'undefined' || typeof arg.location === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // uuid?: string
    ( typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string' ) &&

  true
  );
  }

export function isAviSubnet(arg: any): arg is models.AviSubnet {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // family?: string
    ( typeof arg.family === 'undefined' || typeof arg.family === 'string' ) &&
    // subnet?: string
    ( typeof arg.subnet === 'undefined' || typeof arg.subnet === 'string' ) &&

  true
  );
  }

export function isAviVipNetwork(arg: any): arg is models.AviVipNetwork {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // cloud?: string
    ( typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string' ) &&
    // configedSubnets?: AviSubnet[]
    ( typeof arg.configedSubnets === 'undefined' || (Array.isArray(arg.configedSubnets) && arg.configedSubnets.every((item: unknown) => isAviSubnet(item))) ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // uuid?: string
    ( typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string' ) &&

  true
  );
  }

export function isAWSAccountParams(arg: any): arg is models.AWSAccountParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // accessKeyID?: string
    ( typeof arg.accessKeyID === 'undefined' || typeof arg.accessKeyID === 'string' ) &&
    // profileName?: string
    ( typeof arg.profileName === 'undefined' || typeof arg.profileName === 'string' ) &&
    // region?: string
    ( typeof arg.region === 'undefined' || typeof arg.region === 'string' ) &&
    // secretAccessKey?: string
    ( typeof arg.secretAccessKey === 'undefined' || typeof arg.secretAccessKey === 'string' ) &&
    // sessionToken?: string
    ( typeof arg.sessionToken === 'undefined' || typeof arg.sessionToken === 'string' ) &&

  true
  );
  }

export function isAWSAvailabilityZone(arg: any): arg is models.AWSAvailabilityZone {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // id?: string
    ( typeof arg.id === 'undefined' || typeof arg.id === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isAWSNodeAz(arg: any): arg is models.AWSNodeAz {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // privateSubnetID?: string
    ( typeof arg.privateSubnetID === 'undefined' || typeof arg.privateSubnetID === 'string' ) &&
    // publicSubnetID?: string
    ( typeof arg.publicSubnetID === 'undefined' || typeof arg.publicSubnetID === 'string' ) &&
    // workerNodeType?: string
    ( typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string' ) &&

  true
  );
  }

export function isAWSRegionalClusterParams(arg: any): arg is models.AWSRegionalClusterParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // annotations?: { [key: string]: string }
    ( typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string' ) &&
    // awsAccountParams?: AWSAccountParams
    ( typeof arg.awsAccountParams === 'undefined' || isAWSAccountParams(arg.awsAccountParams) ) &&
    // bastionHostEnabled?: boolean
    ( typeof arg.bastionHostEnabled === 'undefined' || typeof arg.bastionHostEnabled === 'boolean' ) &&
    // ceipOptIn?: boolean
    ( typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean' ) &&
    // clusterName?: string
    ( typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string' ) &&
    // controlPlaneFlavor?: string
    ( typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string' ) &&
    // controlPlaneNodeType?: string
    ( typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string' ) &&
    // createCloudFormationStack?: boolean
    ( typeof arg.createCloudFormationStack === 'undefined' || typeof arg.createCloudFormationStack === 'boolean' ) &&
    // enableAuditLogging?: boolean
    ( typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean' ) &&
    // identityManagement?: IdentityManagementConfig
    ( typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement) ) &&
    // kubernetesVersion?: string
    ( typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string' ) &&
    // labels?: { [key: string]: string }
    ( typeof arg.labels === 'undefined' || typeof arg.labels === 'string' ) &&
    // loadbalancerSchemeInternal?: boolean
    ( typeof arg.loadbalancerSchemeInternal === 'undefined' || typeof arg.loadbalancerSchemeInternal === 'boolean' ) &&
    // machineHealthCheckEnabled?: boolean
    ( typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean' ) &&
    // networking?: TKGNetwork
    ( typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking) ) &&
    // numOfWorkerNode?: number
    ( typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number' ) &&
    // os?: AWSVirtualMachine
    ( typeof arg.os === 'undefined' || isAWSVirtualMachine(arg.os) ) &&
    // sshKeyName?: string
    ( typeof arg.sshKeyName === 'undefined' || typeof arg.sshKeyName === 'string' ) &&
    // vpc?: AWSVpc
    ( typeof arg.vpc === 'undefined' || isAWSVpc(arg.vpc) ) &&
    // workerNodeType?: string
    ( typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string' ) &&

  true
  );
  }

export function isAWSRoute(arg: any): arg is models.AWSRoute {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // DestinationCidrBlock?: string
    ( typeof arg.DestinationCidrBlock === 'undefined' || typeof arg.DestinationCidrBlock === 'string' ) &&
    // GatewayId?: string
    ( typeof arg.GatewayId === 'undefined' || typeof arg.GatewayId === 'string' ) &&
    // State?: string
    ( typeof arg.State === 'undefined' || typeof arg.State === 'string' ) &&

  true
  );
  }

export function isAWSRouteTable(arg: any): arg is models.AWSRouteTable {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // id?: string
    ( typeof arg.id === 'undefined' || typeof arg.id === 'string' ) &&
    // routes?: AWSRoute[]
    ( typeof arg.routes === 'undefined' || (Array.isArray(arg.routes) && arg.routes.every((item: unknown) => isAWSRoute(item))) ) &&
    // vpcId?: string
    ( typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string' ) &&

  true
  );
  }

export function isAWSSubnet(arg: any): arg is models.AWSSubnet {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // availabilityZoneId?: string
    ( typeof arg.availabilityZoneId === 'undefined' || typeof arg.availabilityZoneId === 'string' ) &&
    // availabilityZoneName?: string
    ( typeof arg.availabilityZoneName === 'undefined' || typeof arg.availabilityZoneName === 'string' ) &&
    // cidr?: string
    ( typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string' ) &&
    // id?: string
    ( typeof arg.id === 'undefined' || typeof arg.id === 'string' ) &&
    // isPublic: boolean
    ( typeof arg.isPublic === 'boolean' ) &&
    // state?: string
    ( typeof arg.state === 'undefined' || typeof arg.state === 'string' ) &&
    // vpcId?: string
    ( typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string' ) &&

  true
  );
  }

export function isAWSVirtualMachine(arg: any): arg is models.AWSVirtualMachine {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // osInfo?: OSInfo
    ( typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo) ) &&

  true
  );
  }

export function isAWSVpc(arg: any): arg is models.AWSVpc {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // azs?: AWSNodeAz[]
    ( typeof arg.azs === 'undefined' || (Array.isArray(arg.azs) && arg.azs.every((item: unknown) => isAWSNodeAz(item))) ) &&
    // cidr?: string
    ( typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string' ) &&
    // vpcID?: string
    ( typeof arg.vpcID === 'undefined' || typeof arg.vpcID === 'string' ) &&

  true
  );
  }

export function isAzureAccountParams(arg: any): arg is models.AzureAccountParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // azureCloud?: string
    ( typeof arg.azureCloud === 'undefined' || typeof arg.azureCloud === 'string' ) &&
    // clientId?: string
    ( typeof arg.clientId === 'undefined' || typeof arg.clientId === 'string' ) &&
    // clientSecret?: string
    ( typeof arg.clientSecret === 'undefined' || typeof arg.clientSecret === 'string' ) &&
    // subscriptionId?: string
    ( typeof arg.subscriptionId === 'undefined' || typeof arg.subscriptionId === 'string' ) &&
    // tenantId?: string
    ( typeof arg.tenantId === 'undefined' || typeof arg.tenantId === 'string' ) &&

  true
  );
  }

export function isAzureInstanceType(arg: any): arg is models.AzureInstanceType {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // family?: string
    ( typeof arg.family === 'undefined' || typeof arg.family === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // size?: string
    ( typeof arg.size === 'undefined' || typeof arg.size === 'string' ) &&
    // tier?: string
    ( typeof arg.tier === 'undefined' || typeof arg.tier === 'string' ) &&
    // zones?: string[]
    ( typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every((item: unknown) => typeof item === 'string')) ) &&

  true
  );
  }

export function isAzureLocation(arg: any): arg is models.AzureLocation {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // displayName?: string
    ( typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isAzureRegionalClusterParams(arg: any): arg is models.AzureRegionalClusterParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // annotations?: { [key: string]: string }
    ( typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string' ) &&
    // azureAccountParams?: AzureAccountParams
    ( typeof arg.azureAccountParams === 'undefined' || isAzureAccountParams(arg.azureAccountParams) ) &&
    // ceipOptIn?: boolean
    ( typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean' ) &&
    // clusterName?: string
    ( typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string' ) &&
    // controlPlaneFlavor?: string
    ( typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string' ) &&
    // controlPlaneMachineType?: string
    ( typeof arg.controlPlaneMachineType === 'undefined' || typeof arg.controlPlaneMachineType === 'string' ) &&
    // controlPlaneSubnet?: string
    ( typeof arg.controlPlaneSubnet === 'undefined' || typeof arg.controlPlaneSubnet === 'string' ) &&
    // controlPlaneSubnetCidr?: string
    ( typeof arg.controlPlaneSubnetCidr === 'undefined' || typeof arg.controlPlaneSubnetCidr === 'string' ) &&
    // enableAuditLogging?: boolean
    ( typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean' ) &&
    // frontendPrivateIp?: string
    ( typeof arg.frontendPrivateIp === 'undefined' || typeof arg.frontendPrivateIp === 'string' ) &&
    // identityManagement?: IdentityManagementConfig
    ( typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement) ) &&
    // isPrivateCluster?: boolean
    ( typeof arg.isPrivateCluster === 'undefined' || typeof arg.isPrivateCluster === 'boolean' ) &&
    // kubernetesVersion?: string
    ( typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string' ) &&
    // labels?: { [key: string]: string }
    ( typeof arg.labels === 'undefined' || typeof arg.labels === 'string' ) &&
    // location?: string
    ( typeof arg.location === 'undefined' || typeof arg.location === 'string' ) &&
    // machineHealthCheckEnabled?: boolean
    ( typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean' ) &&
    // networking?: TKGNetwork
    ( typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking) ) &&
    // numOfWorkerNodes?: string
    ( typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string' ) &&
    // os?: AzureVirtualMachine
    ( typeof arg.os === 'undefined' || isAzureVirtualMachine(arg.os) ) &&
    // resourceGroup?: string
    ( typeof arg.resourceGroup === 'undefined' || typeof arg.resourceGroup === 'string' ) &&
    // sshPublicKey?: string
    ( typeof arg.sshPublicKey === 'undefined' || typeof arg.sshPublicKey === 'string' ) &&
    // vnetCidr?: string
    ( typeof arg.vnetCidr === 'undefined' || typeof arg.vnetCidr === 'string' ) &&
    // vnetName?: string
    ( typeof arg.vnetName === 'undefined' || typeof arg.vnetName === 'string' ) &&
    // vnetResourceGroup?: string
    ( typeof arg.vnetResourceGroup === 'undefined' || typeof arg.vnetResourceGroup === 'string' ) &&
    // workerMachineType?: string
    ( typeof arg.workerMachineType === 'undefined' || typeof arg.workerMachineType === 'string' ) &&
    // workerNodeSubnet?: string
    ( typeof arg.workerNodeSubnet === 'undefined' || typeof arg.workerNodeSubnet === 'string' ) &&
    // workerNodeSubnetCidr?: string
    ( typeof arg.workerNodeSubnetCidr === 'undefined' || typeof arg.workerNodeSubnetCidr === 'string' ) &&

  true
  );
  }

export function isAzureResourceGroup(arg: any): arg is models.AzureResourceGroup {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // id?: string
    ( typeof arg.id === 'undefined' || typeof arg.id === 'string' ) &&
    // location: string
    ( typeof arg.location === 'string' ) &&
    // name: string
    ( typeof arg.name === 'string' ) &&

  true
  );
  }

export function isAzureSubnet(arg: any): arg is models.AzureSubnet {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // cidr?: string
    ( typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isAzureVirtualMachine(arg: any): arg is models.AzureVirtualMachine {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // osInfo?: OSInfo
    ( typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo) ) &&

  true
  );
  }

export function isAzureVirtualNetwork(arg: any): arg is models.AzureVirtualNetwork {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // cidrBlock: string
    ( typeof arg.cidrBlock === 'string' ) &&
    // id?: string
    ( typeof arg.id === 'undefined' || typeof arg.id === 'string' ) &&
    // location: string
    ( typeof arg.location === 'string' ) &&
    // name: string
    ( typeof arg.name === 'string' ) &&
    // subnets?: AzureSubnet[]
    ( typeof arg.subnets === 'undefined' || (Array.isArray(arg.subnets) && arg.subnets.every((item: unknown) => isAzureSubnet(item))) ) &&

  true
  );
  }

export function isConfigFile(arg: any): arg is models.ConfigFile {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // filecontents?: string
    ( typeof arg.filecontents === 'undefined' || typeof arg.filecontents === 'string' ) &&

  true
  );
  }

export function isConfigFileInfo(arg: any): arg is models.ConfigFileInfo {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // path?: string
    ( typeof arg.path === 'undefined' || typeof arg.path === 'string' ) &&

  true
  );
  }

export function isDockerDaemonStatus(arg: any): arg is models.DockerDaemonStatus {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // status?: boolean
    ( typeof arg.status === 'undefined' || typeof arg.status === 'boolean' ) &&

  true
  );
  }

export function isDockerRegionalClusterParams(arg: any): arg is models.DockerRegionalClusterParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // annotations?: { [key: string]: string }
    ( typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string' ) &&
    // ceipOptIn?: boolean
    ( typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean' ) &&
    // clusterName?: string
    ( typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string' ) &&
    // controlPlaneFlavor?: string
    ( typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string' ) &&
    // identityManagement?: IdentityManagementConfig
    ( typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement) ) &&
    // kubernetesVersion?: string
    ( typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string' ) &&
    // labels?: { [key: string]: string }
    ( typeof arg.labels === 'undefined' || typeof arg.labels === 'string' ) &&
    // machineHealthCheckEnabled?: boolean
    ( typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean' ) &&
    // networking?: TKGNetwork
    ( typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking) ) &&
    // numOfWorkerNodes?: string
    ( typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string' ) &&

  true
  );
  }

export function isError(arg: any): arg is models.Error {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // message?: string
    ( typeof arg.message === 'undefined' || typeof arg.message === 'string' ) &&

  true
  );
  }

export function isFeatureMap(arg: any): arg is models.FeatureMap {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // [key: string]: string
    ( Object.values(arg).every((value: unknown) => typeof value === 'string') ) &&

  true
  );
  }

export function isFeatures(arg: any): arg is models.Features {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // [key: string]: FeatureMap
    ( Object.values(arg).every((value: unknown) => isFeatureMap(value)) ) &&

  true
  );
  }

export function isHTTPProxyConfiguration(arg: any): arg is models.HTTPProxyConfiguration {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // enabled?: boolean
    ( typeof arg.enabled === 'undefined' || typeof arg.enabled === 'boolean' ) &&
    // HTTPProxyPassword?: string
    ( typeof arg.HTTPProxyPassword === 'undefined' || typeof arg.HTTPProxyPassword === 'string' ) &&
    // HTTPProxyURL?: string
    ( typeof arg.HTTPProxyURL === 'undefined' || typeof arg.HTTPProxyURL === 'string' ) &&
    // HTTPProxyUsername?: string
    ( typeof arg.HTTPProxyUsername === 'undefined' || typeof arg.HTTPProxyUsername === 'string' ) &&
    // HTTPSProxyPassword?: string
    ( typeof arg.HTTPSProxyPassword === 'undefined' || typeof arg.HTTPSProxyPassword === 'string' ) &&
    // HTTPSProxyURL?: string
    ( typeof arg.HTTPSProxyURL === 'undefined' || typeof arg.HTTPSProxyURL === 'string' ) &&
    // HTTPSProxyUsername?: string
    ( typeof arg.HTTPSProxyUsername === 'undefined' || typeof arg.HTTPSProxyUsername === 'string' ) &&
    // noProxy?: string
    ( typeof arg.noProxy === 'undefined' || typeof arg.noProxy === 'string' ) &&

  true
  );
  }

export function isIdentityManagementConfig(arg: any): arg is models.IdentityManagementConfig {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // idm_type: 'oidc' | 'ldap' | 'none'
    ( ['oidc', 'ldap', 'none'].includes(arg.idm_type) ) &&
    // ldap_bind_dn?: string
    ( typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string' ) &&
    // ldap_bind_password?: string
    ( typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string' ) &&
    // ldap_group_search_base_dn?: string
    ( typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string' ) &&
    // ldap_group_search_filter?: string
    ( typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string' ) &&
    // ldap_group_search_group_attr?: string
    ( typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string' ) &&
    // ldap_group_search_name_attr?: string
    ( typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string' ) &&
    // ldap_group_search_user_attr?: string
    ( typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string' ) &&
    // ldap_root_ca?: string
    ( typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string' ) &&
    // ldap_url?: string
    ( typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string' ) &&
    // ldap_user_search_base_dn?: string
    ( typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string' ) &&
    // ldap_user_search_email_attr?: string
    ( typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string' ) &&
    // ldap_user_search_filter?: string
    ( typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string' ) &&
    // ldap_user_search_id_attr?: string
    ( typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string' ) &&
    // ldap_user_search_name_attr?: string
    ( typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string' ) &&
    // ldap_user_search_username?: string
    ( typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string' ) &&
    // oidc_claim_mappings?: { [key: string]: string }
    ( typeof arg.oidc_claim_mappings === 'undefined' || typeof arg.oidc_claim_mappings === 'string' ) &&
    // oidc_client_id?: string
    ( typeof arg.oidc_client_id === 'undefined' || typeof arg.oidc_client_id === 'string' ) &&
    // oidc_client_secret?: string
    ( typeof arg.oidc_client_secret === 'undefined' || typeof arg.oidc_client_secret === 'string' ) &&
    // oidc_provider_name?: string
    ( typeof arg.oidc_provider_name === 'undefined' || typeof arg.oidc_provider_name === 'string' ) &&
    // oidc_provider_url?: string
    ( typeof arg.oidc_provider_url === 'undefined' || typeof arg.oidc_provider_url === 'string' ) &&
    // oidc_scope?: string
    ( typeof arg.oidc_scope === 'undefined' || typeof arg.oidc_scope === 'string' ) &&
    // oidc_skip_verify_cert?: boolean
    ( typeof arg.oidc_skip_verify_cert === 'undefined' || typeof arg.oidc_skip_verify_cert === 'boolean' ) &&

  true
  );
  }

export function isLdapParams(arg: any): arg is models.LdapParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // ldap_bind_dn?: string
    ( typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string' ) &&
    // ldap_bind_password?: string
    ( typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string' ) &&
    // ldap_group_search_base_dn?: string
    ( typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string' ) &&
    // ldap_group_search_filter?: string
    ( typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string' ) &&
    // ldap_group_search_group_attr?: string
    ( typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string' ) &&
    // ldap_group_search_name_attr?: string
    ( typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string' ) &&
    // ldap_group_search_user_attr?: string
    ( typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string' ) &&
    // ldap_root_ca?: string
    ( typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string' ) &&
    // ldap_test_group?: string
    ( typeof arg.ldap_test_group === 'undefined' || typeof arg.ldap_test_group === 'string' ) &&
    // ldap_test_user?: string
    ( typeof arg.ldap_test_user === 'undefined' || typeof arg.ldap_test_user === 'string' ) &&
    // ldap_url?: string
    ( typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string' ) &&
    // ldap_user_search_base_dn?: string
    ( typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string' ) &&
    // ldap_user_search_email_attr?: string
    ( typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string' ) &&
    // ldap_user_search_filter?: string
    ( typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string' ) &&
    // ldap_user_search_id_attr?: string
    ( typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string' ) &&
    // ldap_user_search_name_attr?: string
    ( typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string' ) &&
    // ldap_user_search_username?: string
    ( typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string' ) &&

  true
  );
  }

export function isLdapTestResult(arg: any): arg is models.LdapTestResult {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // code?: number
    ( typeof arg.code === 'undefined' || typeof arg.code === 'number' ) &&
    // desc?: string
    ( typeof arg.desc === 'undefined' || typeof arg.desc === 'string' ) &&

  true
  );
  }

export function isNodeType(arg: any): arg is models.NodeType {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // cpu?: number
    ( typeof arg.cpu === 'undefined' || typeof arg.cpu === 'number' ) &&
    // disk?: number
    ( typeof arg.disk === 'undefined' || typeof arg.disk === 'number' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // ram?: number
    ( typeof arg.ram === 'undefined' || typeof arg.ram === 'number' ) &&

  true
  );
  }

export function isOSInfo(arg: any): arg is models.OSInfo {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // arch?: string
    ( typeof arg.arch === 'undefined' || typeof arg.arch === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // version?: string
    ( typeof arg.version === 'undefined' || typeof arg.version === 'string' ) &&

  true
  );
  }

export function isProviderInfo(arg: any): arg is models.ProviderInfo {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // provider?: string
    ( typeof arg.provider === 'undefined' || typeof arg.provider === 'string' ) &&
    // tkrVersion?: string
    ( typeof arg.tkrVersion === 'undefined' || typeof arg.tkrVersion === 'string' ) &&

  true
  );
  }

export function isTKGNetwork(arg: any): arg is models.TKGNetwork {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // clusterDNSName?: string
    ( typeof arg.clusterDNSName === 'undefined' || typeof arg.clusterDNSName === 'string' ) &&
    // clusterNodeCIDR?: string
    ( typeof arg.clusterNodeCIDR === 'undefined' || typeof arg.clusterNodeCIDR === 'string' ) &&
    // clusterPodCIDR?: string
    ( typeof arg.clusterPodCIDR === 'undefined' || typeof arg.clusterPodCIDR === 'string' ) &&
    // clusterServiceCIDR?: string
    ( typeof arg.clusterServiceCIDR === 'undefined' || typeof arg.clusterServiceCIDR === 'string' ) &&
    // cniType?: string
    ( typeof arg.cniType === 'undefined' || typeof arg.cniType === 'string' ) &&
    // httpProxyConfiguration?: HTTPProxyConfiguration
    ( typeof arg.httpProxyConfiguration === 'undefined' || isHTTPProxyConfiguration(arg.httpProxyConfiguration) ) &&
    // networkName?: string
    ( typeof arg.networkName === 'undefined' || typeof arg.networkName === 'string' ) &&

  true
  );
  }

export function isVpc(arg: any): arg is models.Vpc {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // cidr?: string
    ( typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string' ) &&
    // id?: string
    ( typeof arg.id === 'undefined' || typeof arg.id === 'string' ) &&

  true
  );
  }

export function isVSphereAvailabilityZone(arg: any): arg is models.VSphereAvailabilityZone {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isVSphereCredentials(arg: any): arg is models.VSphereCredentials {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // host?: string
    ( typeof arg.host === 'undefined' || typeof arg.host === 'string' ) &&
    // insecure?: boolean
    ( typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean' ) &&
    // password?: string
    ( typeof arg.password === 'undefined' || typeof arg.password === 'string' ) &&
    // thumbprint?: string
    ( typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string' ) &&
    // username?: string
    ( typeof arg.username === 'undefined' || typeof arg.username === 'string' ) &&

  true
  );
  }

export function isVSphereDatacenter(arg: any): arg is models.VSphereDatacenter {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isVSphereDatastore(arg: any): arg is models.VSphereDatastore {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isVSphereFolder(arg: any): arg is models.VSphereFolder {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isVsphereInfo(arg: any): arg is models.VsphereInfo {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // hasPacific?: string
    ( typeof arg.hasPacific === 'undefined' || typeof arg.hasPacific === 'string' ) &&
    // version?: string
    ( typeof arg.version === 'undefined' || typeof arg.version === 'string' ) &&

  true
  );
  }

export function isVSphereManagementObject(arg: any): arg is models.VSphereManagementObject {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // parentMoid?: string
    ( typeof arg.parentMoid === 'undefined' || typeof arg.parentMoid === 'string' ) &&
    // path?: string
    ( typeof arg.path === 'undefined' || typeof arg.path === 'string' ) &&
    // resourceType?: 'datacenter' | 'cluster' | 'hostgroup' | 'folder' | 'respool' | 'vm' | 'datastore' | 'host' | 'network'
    ( typeof arg.resourceType === 'undefined' || ['datacenter', 'cluster', 'hostgroup', 'folder', 'respool', 'vm', 'datastore', 'host', 'network'].includes(arg.resourceType) ) &&

  true
  );
  }

export function isVSphereNetwork(arg: any): arg is models.VSphereNetwork {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // displayName?: string
    ( typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string' ) &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isVSphereRegion(arg: any): arg is models.VSphereRegion {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // zones?: VSphereAvailabilityZone[]
    ( typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every((item: unknown) => isVSphereAvailabilityZone(item))) ) &&

  true
  );
  }

export function isVsphereRegionalClusterParams(arg: any): arg is models.VsphereRegionalClusterParams {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // annotations?: { [key: string]: string }
    ( typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string' ) &&
    // aviConfig?: AviConfig
    ( typeof arg.aviConfig === 'undefined' || isAviConfig(arg.aviConfig) ) &&
    // ceipOptIn?: boolean
    ( typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean' ) &&
    // clusterName?: string
    ( typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string' ) &&
    // controlPlaneEndpoint?: string
    ( typeof arg.controlPlaneEndpoint === 'undefined' || typeof arg.controlPlaneEndpoint === 'string' ) &&
    // controlPlaneFlavor?: string
    ( typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string' ) &&
    // controlPlaneNodeType?: string
    ( typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string' ) &&
    // datacenter?: string
    ( typeof arg.datacenter === 'undefined' || typeof arg.datacenter === 'string' ) &&
    // datastore?: string
    ( typeof arg.datastore === 'undefined' || typeof arg.datastore === 'string' ) &&
    // enableAuditLogging?: boolean
    ( typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean' ) &&
    // folder?: string
    ( typeof arg.folder === 'undefined' || typeof arg.folder === 'string' ) &&
    // identityManagement?: IdentityManagementConfig
    ( typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement) ) &&
    // ipFamily?: string
    ( typeof arg.ipFamily === 'undefined' || typeof arg.ipFamily === 'string' ) &&
    // kubernetesVersion?: string
    ( typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string' ) &&
    // labels?: { [key: string]: string }
    ( typeof arg.labels === 'undefined' || typeof arg.labels === 'string' ) &&
    // machineHealthCheckEnabled?: boolean
    ( typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean' ) &&
    // networking?: TKGNetwork
    ( typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking) ) &&
    // numOfWorkerNode?: number
    ( typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number' ) &&
    // os?: VSphereVirtualMachine
    ( typeof arg.os === 'undefined' || isVSphereVirtualMachine(arg.os) ) &&
    // resourcePool?: string
    ( typeof arg.resourcePool === 'undefined' || typeof arg.resourcePool === 'string' ) &&
    // ssh_key?: string
    ( typeof arg.ssh_key === 'undefined' || typeof arg.ssh_key === 'string' ) &&
    // vsphereCredentials?: VSphereCredentials
    ( typeof arg.vsphereCredentials === 'undefined' || isVSphereCredentials(arg.vsphereCredentials) ) &&
    // workerNodeType?: string
    ( typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string' ) &&

  true
  );
  }

export function isVSphereResourcePool(arg: any): arg is models.VSphereResourcePool {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&

  true
  );
  }

export function isVSphereThumbprint(arg: any): arg is models.VSphereThumbprint {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // insecure?: boolean
    ( typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean' ) &&
    // thumbprint?: string
    ( typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string' ) &&

  true
  );
  }

export function isVSphereVirtualMachine(arg: any): arg is models.VSphereVirtualMachine {
  return (
  arg != null &&
  typeof arg === 'object' &&
    // isTemplate: boolean
    ( typeof arg.isTemplate === 'boolean' ) &&
    // k8sVersion?: string
    ( typeof arg.k8sVersion === 'undefined' || typeof arg.k8sVersion === 'string' ) &&
    // moid?: string
    ( typeof arg.moid === 'undefined' || typeof arg.moid === 'string' ) &&
    // name?: string
    ( typeof arg.name === 'undefined' || typeof arg.name === 'string' ) &&
    // osInfo?: OSInfo
    ( typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo) ) &&

  true
  );
  }


