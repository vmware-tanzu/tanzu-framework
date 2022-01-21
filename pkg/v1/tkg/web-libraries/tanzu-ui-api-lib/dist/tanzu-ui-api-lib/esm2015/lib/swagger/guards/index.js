/* tslint:disable */
/* pre-prepared guards for build in complex types */
function _isBlob(arg) {
    return arg != null && typeof arg.size === 'number' && typeof arg.type === 'string' && typeof arg.slice === 'function';
}
export function isFile(arg) {
    return arg != null && typeof arg.lastModified === 'number' && typeof arg.name === 'string' && _isBlob(arg);
}
/* generated type guards */
export function isAviCloud(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // location?: string
        (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // uuid?: string
        (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
        true);
}
export function isAviConfig(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // ca_cert?: string
        (typeof arg.ca_cert === 'undefined' || typeof arg.ca_cert === 'string') &&
        // cloud?: string
        (typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string') &&
        // controller?: string
        (typeof arg.controller === 'undefined' || typeof arg.controller === 'string') &&
        // controlPlaneHaProvider?: boolean
        (typeof arg.controlPlaneHaProvider === 'undefined' || typeof arg.controlPlaneHaProvider === 'boolean') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // managementClusterVipNetworkCidr?: string
        (typeof arg.managementClusterVipNetworkCidr === 'undefined' || typeof arg.managementClusterVipNetworkCidr === 'string') &&
        // managementClusterVipNetworkName?: string
        (typeof arg.managementClusterVipNetworkName === 'undefined' || typeof arg.managementClusterVipNetworkName === 'string') &&
        // network?: AviNetworkParams
        (typeof arg.network === 'undefined' || isAviNetworkParams(arg.network)) &&
        // password?: string
        (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
        // service_engine?: string
        (typeof arg.service_engine === 'undefined' || typeof arg.service_engine === 'string') &&
        // username?: string
        (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
        true);
}
export function isAviControllerParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // CAData?: string
        (typeof arg.CAData === 'undefined' || typeof arg.CAData === 'string') &&
        // host?: string
        (typeof arg.host === 'undefined' || typeof arg.host === 'string') &&
        // password?: string
        (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
        // tenant?: string
        (typeof arg.tenant === 'undefined' || typeof arg.tenant === 'string') &&
        // username?: string
        (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
        true);
}
export function isAviNetworkParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isAviServiceEngineGroup(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // location?: string
        (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // uuid?: string
        (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
        true);
}
export function isAviSubnet(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // family?: string
        (typeof arg.family === 'undefined' || typeof arg.family === 'string') &&
        // subnet?: string
        (typeof arg.subnet === 'undefined' || typeof arg.subnet === 'string') &&
        true);
}
export function isAviVipNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cloud?: string
        (typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string') &&
        // configedSubnets?: AviSubnet[]
        (typeof arg.configedSubnets === 'undefined' || (Array.isArray(arg.configedSubnets) && arg.configedSubnets.every((item) => isAviSubnet(item)))) &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // uuid?: string
        (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
        true);
}
export function isAWSAccountParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // accessKeyID?: string
        (typeof arg.accessKeyID === 'undefined' || typeof arg.accessKeyID === 'string') &&
        // profileName?: string
        (typeof arg.profileName === 'undefined' || typeof arg.profileName === 'string') &&
        // region?: string
        (typeof arg.region === 'undefined' || typeof arg.region === 'string') &&
        // secretAccessKey?: string
        (typeof arg.secretAccessKey === 'undefined' || typeof arg.secretAccessKey === 'string') &&
        // sessionToken?: string
        (typeof arg.sessionToken === 'undefined' || typeof arg.sessionToken === 'string') &&
        true);
}
export function isAWSAvailabilityZone(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isAWSNodeAz(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // privateSubnetID?: string
        (typeof arg.privateSubnetID === 'undefined' || typeof arg.privateSubnetID === 'string') &&
        // publicSubnetID?: string
        (typeof arg.publicSubnetID === 'undefined' || typeof arg.publicSubnetID === 'string') &&
        // workerNodeType?: string
        (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
        true);
}
export function isAWSRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // awsAccountParams?: AWSAccountParams
        (typeof arg.awsAccountParams === 'undefined' || isAWSAccountParams(arg.awsAccountParams)) &&
        // bastionHostEnabled?: boolean
        (typeof arg.bastionHostEnabled === 'undefined' || typeof arg.bastionHostEnabled === 'boolean') &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // controlPlaneNodeType?: string
        (typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string') &&
        // createCloudFormationStack?: boolean
        (typeof arg.createCloudFormationStack === 'undefined' || typeof arg.createCloudFormationStack === 'boolean') &&
        // enableAuditLogging?: boolean
        (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // loadbalancerSchemeInternal?: boolean
        (typeof arg.loadbalancerSchemeInternal === 'undefined' || typeof arg.loadbalancerSchemeInternal === 'boolean') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNode?: number
        (typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number') &&
        // os?: AWSVirtualMachine
        (typeof arg.os === 'undefined' || isAWSVirtualMachine(arg.os)) &&
        // sshKeyName?: string
        (typeof arg.sshKeyName === 'undefined' || typeof arg.sshKeyName === 'string') &&
        // vpc?: AWSVpc
        (typeof arg.vpc === 'undefined' || isAWSVpc(arg.vpc)) &&
        // workerNodeType?: string
        (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
        true);
}
export function isAWSRoute(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // DestinationCidrBlock?: string
        (typeof arg.DestinationCidrBlock === 'undefined' || typeof arg.DestinationCidrBlock === 'string') &&
        // GatewayId?: string
        (typeof arg.GatewayId === 'undefined' || typeof arg.GatewayId === 'string') &&
        // State?: string
        (typeof arg.State === 'undefined' || typeof arg.State === 'string') &&
        true);
}
export function isAWSRouteTable(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // routes?: AWSRoute[]
        (typeof arg.routes === 'undefined' || (Array.isArray(arg.routes) && arg.routes.every((item) => isAWSRoute(item)))) &&
        // vpcId?: string
        (typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string') &&
        true);
}
export function isAWSSubnet(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // availabilityZoneId?: string
        (typeof arg.availabilityZoneId === 'undefined' || typeof arg.availabilityZoneId === 'string') &&
        // availabilityZoneName?: string
        (typeof arg.availabilityZoneName === 'undefined' || typeof arg.availabilityZoneName === 'string') &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // isPublic: boolean
        (typeof arg.isPublic === 'boolean') &&
        // state?: string
        (typeof arg.state === 'undefined' || typeof arg.state === 'string') &&
        // vpcId?: string
        (typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string') &&
        true);
}
export function isAWSVirtualMachine(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // osInfo?: OSInfo
        (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
        true);
}
export function isAWSVpc(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // azs?: AWSNodeAz[]
        (typeof arg.azs === 'undefined' || (Array.isArray(arg.azs) && arg.azs.every((item) => isAWSNodeAz(item)))) &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // vpcID?: string
        (typeof arg.vpcID === 'undefined' || typeof arg.vpcID === 'string') &&
        true);
}
export function isAzureAccountParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // azureCloud?: string
        (typeof arg.azureCloud === 'undefined' || typeof arg.azureCloud === 'string') &&
        // clientId?: string
        (typeof arg.clientId === 'undefined' || typeof arg.clientId === 'string') &&
        // clientSecret?: string
        (typeof arg.clientSecret === 'undefined' || typeof arg.clientSecret === 'string') &&
        // subscriptionId?: string
        (typeof arg.subscriptionId === 'undefined' || typeof arg.subscriptionId === 'string') &&
        // tenantId?: string
        (typeof arg.tenantId === 'undefined' || typeof arg.tenantId === 'string') &&
        true);
}
export function isAzureInstanceType(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // family?: string
        (typeof arg.family === 'undefined' || typeof arg.family === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // size?: string
        (typeof arg.size === 'undefined' || typeof arg.size === 'string') &&
        // tier?: string
        (typeof arg.tier === 'undefined' || typeof arg.tier === 'string') &&
        // zones?: string[]
        (typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every((item) => typeof item === 'string'))) &&
        true);
}
export function isAzureLocation(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // displayName?: string
        (typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isAzureRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // azureAccountParams?: AzureAccountParams
        (typeof arg.azureAccountParams === 'undefined' || isAzureAccountParams(arg.azureAccountParams)) &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // controlPlaneMachineType?: string
        (typeof arg.controlPlaneMachineType === 'undefined' || typeof arg.controlPlaneMachineType === 'string') &&
        // controlPlaneSubnet?: string
        (typeof arg.controlPlaneSubnet === 'undefined' || typeof arg.controlPlaneSubnet === 'string') &&
        // controlPlaneSubnetCidr?: string
        (typeof arg.controlPlaneSubnetCidr === 'undefined' || typeof arg.controlPlaneSubnetCidr === 'string') &&
        // enableAuditLogging?: boolean
        (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
        // frontendPrivateIp?: string
        (typeof arg.frontendPrivateIp === 'undefined' || typeof arg.frontendPrivateIp === 'string') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // isPrivateCluster?: boolean
        (typeof arg.isPrivateCluster === 'undefined' || typeof arg.isPrivateCluster === 'boolean') &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // location?: string
        (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNodes?: string
        (typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string') &&
        // os?: AzureVirtualMachine
        (typeof arg.os === 'undefined' || isAzureVirtualMachine(arg.os)) &&
        // resourceGroup?: string
        (typeof arg.resourceGroup === 'undefined' || typeof arg.resourceGroup === 'string') &&
        // sshPublicKey?: string
        (typeof arg.sshPublicKey === 'undefined' || typeof arg.sshPublicKey === 'string') &&
        // vnetCidr?: string
        (typeof arg.vnetCidr === 'undefined' || typeof arg.vnetCidr === 'string') &&
        // vnetName?: string
        (typeof arg.vnetName === 'undefined' || typeof arg.vnetName === 'string') &&
        // vnetResourceGroup?: string
        (typeof arg.vnetResourceGroup === 'undefined' || typeof arg.vnetResourceGroup === 'string') &&
        // workerMachineType?: string
        (typeof arg.workerMachineType === 'undefined' || typeof arg.workerMachineType === 'string') &&
        // workerNodeSubnet?: string
        (typeof arg.workerNodeSubnet === 'undefined' || typeof arg.workerNodeSubnet === 'string') &&
        // workerNodeSubnetCidr?: string
        (typeof arg.workerNodeSubnetCidr === 'undefined' || typeof arg.workerNodeSubnetCidr === 'string') &&
        true);
}
export function isAzureResourceGroup(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // location: string
        (typeof arg.location === 'string') &&
        // name: string
        (typeof arg.name === 'string') &&
        true);
}
export function isAzureSubnet(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isAzureVirtualMachine(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // osInfo?: OSInfo
        (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
        true);
}
export function isAzureVirtualNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidrBlock: string
        (typeof arg.cidrBlock === 'string') &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // location: string
        (typeof arg.location === 'string') &&
        // name: string
        (typeof arg.name === 'string') &&
        // subnets?: AzureSubnet[]
        (typeof arg.subnets === 'undefined' || (Array.isArray(arg.subnets) && arg.subnets.every((item) => isAzureSubnet(item)))) &&
        true);
}
export function isConfigFile(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // filecontents?: string
        (typeof arg.filecontents === 'undefined' || typeof arg.filecontents === 'string') &&
        true);
}
export function isConfigFileInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // path?: string
        (typeof arg.path === 'undefined' || typeof arg.path === 'string') &&
        true);
}
export function isDockerDaemonStatus(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // status?: boolean
        (typeof arg.status === 'undefined' || typeof arg.status === 'boolean') &&
        true);
}
export function isDockerRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNodes?: string
        (typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string') &&
        true);
}
export function isError(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // message?: string
        (typeof arg.message === 'undefined' || typeof arg.message === 'string') &&
        true);
}
export function isFeatureMap(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // [key: string]: string
        (Object.values(arg).every((value) => typeof value === 'string')) &&
        true);
}
export function isFeatures(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // [key: string]: FeatureMap
        (Object.values(arg).every((value) => isFeatureMap(value))) &&
        true);
}
export function isHTTPProxyConfiguration(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // enabled?: boolean
        (typeof arg.enabled === 'undefined' || typeof arg.enabled === 'boolean') &&
        // HTTPProxyPassword?: string
        (typeof arg.HTTPProxyPassword === 'undefined' || typeof arg.HTTPProxyPassword === 'string') &&
        // HTTPProxyURL?: string
        (typeof arg.HTTPProxyURL === 'undefined' || typeof arg.HTTPProxyURL === 'string') &&
        // HTTPProxyUsername?: string
        (typeof arg.HTTPProxyUsername === 'undefined' || typeof arg.HTTPProxyUsername === 'string') &&
        // HTTPSProxyPassword?: string
        (typeof arg.HTTPSProxyPassword === 'undefined' || typeof arg.HTTPSProxyPassword === 'string') &&
        // HTTPSProxyURL?: string
        (typeof arg.HTTPSProxyURL === 'undefined' || typeof arg.HTTPSProxyURL === 'string') &&
        // HTTPSProxyUsername?: string
        (typeof arg.HTTPSProxyUsername === 'undefined' || typeof arg.HTTPSProxyUsername === 'string') &&
        // noProxy?: string
        (typeof arg.noProxy === 'undefined' || typeof arg.noProxy === 'string') &&
        true);
}
export function isIdentityManagementConfig(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // idm_type: 'oidc' | 'ldap' | 'none'
        (['oidc', 'ldap', 'none'].includes(arg.idm_type)) &&
        // ldap_bind_dn?: string
        (typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string') &&
        // ldap_bind_password?: string
        (typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string') &&
        // ldap_group_search_base_dn?: string
        (typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string') &&
        // ldap_group_search_filter?: string
        (typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string') &&
        // ldap_group_search_group_attr?: string
        (typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string') &&
        // ldap_group_search_name_attr?: string
        (typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string') &&
        // ldap_group_search_user_attr?: string
        (typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string') &&
        // ldap_root_ca?: string
        (typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string') &&
        // ldap_url?: string
        (typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string') &&
        // ldap_user_search_base_dn?: string
        (typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string') &&
        // ldap_user_search_email_attr?: string
        (typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string') &&
        // ldap_user_search_filter?: string
        (typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string') &&
        // ldap_user_search_id_attr?: string
        (typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string') &&
        // ldap_user_search_name_attr?: string
        (typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string') &&
        // ldap_user_search_username?: string
        (typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string') &&
        // oidc_claim_mappings?: { [key: string]: string }
        (typeof arg.oidc_claim_mappings === 'undefined' || typeof arg.oidc_claim_mappings === 'string') &&
        // oidc_client_id?: string
        (typeof arg.oidc_client_id === 'undefined' || typeof arg.oidc_client_id === 'string') &&
        // oidc_client_secret?: string
        (typeof arg.oidc_client_secret === 'undefined' || typeof arg.oidc_client_secret === 'string') &&
        // oidc_provider_name?: string
        (typeof arg.oidc_provider_name === 'undefined' || typeof arg.oidc_provider_name === 'string') &&
        // oidc_provider_url?: string
        (typeof arg.oidc_provider_url === 'undefined' || typeof arg.oidc_provider_url === 'string') &&
        // oidc_scope?: string
        (typeof arg.oidc_scope === 'undefined' || typeof arg.oidc_scope === 'string') &&
        // oidc_skip_verify_cert?: boolean
        (typeof arg.oidc_skip_verify_cert === 'undefined' || typeof arg.oidc_skip_verify_cert === 'boolean') &&
        true);
}
export function isLdapParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // ldap_bind_dn?: string
        (typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string') &&
        // ldap_bind_password?: string
        (typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string') &&
        // ldap_group_search_base_dn?: string
        (typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string') &&
        // ldap_group_search_filter?: string
        (typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string') &&
        // ldap_group_search_group_attr?: string
        (typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string') &&
        // ldap_group_search_name_attr?: string
        (typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string') &&
        // ldap_group_search_user_attr?: string
        (typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string') &&
        // ldap_root_ca?: string
        (typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string') &&
        // ldap_test_group?: string
        (typeof arg.ldap_test_group === 'undefined' || typeof arg.ldap_test_group === 'string') &&
        // ldap_test_user?: string
        (typeof arg.ldap_test_user === 'undefined' || typeof arg.ldap_test_user === 'string') &&
        // ldap_url?: string
        (typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string') &&
        // ldap_user_search_base_dn?: string
        (typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string') &&
        // ldap_user_search_email_attr?: string
        (typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string') &&
        // ldap_user_search_filter?: string
        (typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string') &&
        // ldap_user_search_id_attr?: string
        (typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string') &&
        // ldap_user_search_name_attr?: string
        (typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string') &&
        // ldap_user_search_username?: string
        (typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string') &&
        true);
}
export function isLdapTestResult(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // code?: number
        (typeof arg.code === 'undefined' || typeof arg.code === 'number') &&
        // desc?: string
        (typeof arg.desc === 'undefined' || typeof arg.desc === 'string') &&
        true);
}
export function isNodeType(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cpu?: number
        (typeof arg.cpu === 'undefined' || typeof arg.cpu === 'number') &&
        // disk?: number
        (typeof arg.disk === 'undefined' || typeof arg.disk === 'number') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // ram?: number
        (typeof arg.ram === 'undefined' || typeof arg.ram === 'number') &&
        true);
}
export function isOSInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // arch?: string
        (typeof arg.arch === 'undefined' || typeof arg.arch === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // version?: string
        (typeof arg.version === 'undefined' || typeof arg.version === 'string') &&
        true);
}
export function isProviderInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // provider?: string
        (typeof arg.provider === 'undefined' || typeof arg.provider === 'string') &&
        // tkrVersion?: string
        (typeof arg.tkrVersion === 'undefined' || typeof arg.tkrVersion === 'string') &&
        true);
}
export function isTKGNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // clusterDNSName?: string
        (typeof arg.clusterDNSName === 'undefined' || typeof arg.clusterDNSName === 'string') &&
        // clusterNodeCIDR?: string
        (typeof arg.clusterNodeCIDR === 'undefined' || typeof arg.clusterNodeCIDR === 'string') &&
        // clusterPodCIDR?: string
        (typeof arg.clusterPodCIDR === 'undefined' || typeof arg.clusterPodCIDR === 'string') &&
        // clusterServiceCIDR?: string
        (typeof arg.clusterServiceCIDR === 'undefined' || typeof arg.clusterServiceCIDR === 'string') &&
        // cniType?: string
        (typeof arg.cniType === 'undefined' || typeof arg.cniType === 'string') &&
        // httpProxyConfiguration?: HTTPProxyConfiguration
        (typeof arg.httpProxyConfiguration === 'undefined' || isHTTPProxyConfiguration(arg.httpProxyConfiguration)) &&
        // networkName?: string
        (typeof arg.networkName === 'undefined' || typeof arg.networkName === 'string') &&
        true);
}
export function isVpc(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        true);
}
export function isVSphereAvailabilityZone(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isVSphereCredentials(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // host?: string
        (typeof arg.host === 'undefined' || typeof arg.host === 'string') &&
        // insecure?: boolean
        (typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean') &&
        // password?: string
        (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
        // thumbprint?: string
        (typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string') &&
        // username?: string
        (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
        true);
}
export function isVSphereDatacenter(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isVSphereDatastore(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isVSphereFolder(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isVsphereInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // hasPacific?: string
        (typeof arg.hasPacific === 'undefined' || typeof arg.hasPacific === 'string') &&
        // version?: string
        (typeof arg.version === 'undefined' || typeof arg.version === 'string') &&
        true);
}
export function isVSphereManagementObject(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // parentMoid?: string
        (typeof arg.parentMoid === 'undefined' || typeof arg.parentMoid === 'string') &&
        // path?: string
        (typeof arg.path === 'undefined' || typeof arg.path === 'string') &&
        // resourceType?: 'datacenter' | 'cluster' | 'hostgroup' | 'folder' | 'respool' | 'vm' | 'datastore' | 'host' | 'network'
        (typeof arg.resourceType === 'undefined' || ['datacenter', 'cluster', 'hostgroup', 'folder', 'respool', 'vm', 'datastore', 'host', 'network'].includes(arg.resourceType)) &&
        true);
}
export function isVSphereNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // displayName?: string
        (typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string') &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isVSphereRegion(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // zones?: VSphereAvailabilityZone[]
        (typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every((item) => isVSphereAvailabilityZone(item)))) &&
        true);
}
export function isVsphereRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // aviConfig?: AviConfig
        (typeof arg.aviConfig === 'undefined' || isAviConfig(arg.aviConfig)) &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneEndpoint?: string
        (typeof arg.controlPlaneEndpoint === 'undefined' || typeof arg.controlPlaneEndpoint === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // controlPlaneNodeType?: string
        (typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string') &&
        // datacenter?: string
        (typeof arg.datacenter === 'undefined' || typeof arg.datacenter === 'string') &&
        // datastore?: string
        (typeof arg.datastore === 'undefined' || typeof arg.datastore === 'string') &&
        // enableAuditLogging?: boolean
        (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
        // folder?: string
        (typeof arg.folder === 'undefined' || typeof arg.folder === 'string') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // ipFamily?: string
        (typeof arg.ipFamily === 'undefined' || typeof arg.ipFamily === 'string') &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNode?: number
        (typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number') &&
        // os?: VSphereVirtualMachine
        (typeof arg.os === 'undefined' || isVSphereVirtualMachine(arg.os)) &&
        // resourcePool?: string
        (typeof arg.resourcePool === 'undefined' || typeof arg.resourcePool === 'string') &&
        // ssh_key?: string
        (typeof arg.ssh_key === 'undefined' || typeof arg.ssh_key === 'string') &&
        // vsphereCredentials?: VSphereCredentials
        (typeof arg.vsphereCredentials === 'undefined' || isVSphereCredentials(arg.vsphereCredentials)) &&
        // workerNodeType?: string
        (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
        true);
}
export function isVSphereResourcePool(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
export function isVSphereThumbprint(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // insecure?: boolean
        (typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean') &&
        // thumbprint?: string
        (typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string') &&
        true);
}
export function isVSphereVirtualMachine(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // isTemplate: boolean
        (typeof arg.isTemplate === 'boolean') &&
        // k8sVersion?: string
        (typeof arg.k8sVersion === 'undefined' || typeof arg.k8sVersion === 'string') &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // osInfo?: OSInfo
        (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
        true);
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi8uLi8uLi8uLi8uLi9wcm9qZWN0cy90YW56dS11aS1hcGktbGliL3NyYy9saWIvc3dhZ2dlci9ndWFyZHMvaW5kZXgudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsb0JBQW9CO0FBSXBCLG9EQUFvRDtBQUVwRCxTQUFTLE9BQU8sQ0FBQyxHQUFRO0lBQ3ZCLE9BQU8sR0FBRyxJQUFJLElBQUksSUFBSSxPQUFPLEdBQUcsQ0FBQyxJQUFJLEtBQUssUUFBUSxJQUFJLE9BQU8sR0FBRyxDQUFDLElBQUksS0FBSyxRQUFRLElBQUksT0FBTyxHQUFHLENBQUMsS0FBSyxLQUFLLFVBQVUsQ0FBQztBQUN4SCxDQUFDO0FBRUQsTUFBTSxVQUFVLE1BQU0sQ0FBQyxHQUFRO0lBQy9CLE9BQU8sR0FBRyxJQUFJLElBQUksSUFBSSxPQUFPLEdBQUcsQ0FBQyxZQUFZLEtBQUssUUFBUSxJQUFJLE9BQU8sR0FBRyxDQUFDLElBQUksS0FBSyxRQUFRLElBQUksT0FBTyxDQUFDLEdBQUcsQ0FBQyxDQUFDO0FBQzNHLENBQUM7QUFFRCwyQkFBMkI7QUFFM0IsTUFBTSxVQUFVLFVBQVUsQ0FBQyxHQUFRO0lBQ2pDLE9BQU8sQ0FDUCxHQUFHLElBQUksSUFBSTtRQUNYLE9BQU8sR0FBRyxLQUFLLFFBQVE7UUFDckIsb0JBQW9CO1FBQ3BCLENBQUUsT0FBTyxHQUFHLENBQUMsUUFBUSxLQUFLLFdBQVcsSUFBSSxPQUFPLEdBQUcsQ0FBQyxRQUFRLEtBQUssUUFBUSxDQUFFO1FBQzNFLGdCQUFnQjtRQUNoQixDQUFFLE9BQU8sR0FBRyxDQUFDLElBQUksS0FBSyxXQUFXLElBQUksT0FBTyxHQUFHLENBQUMsSUFBSSxLQUFLLFFBQVEsQ0FBRTtRQUNuRSxnQkFBZ0I7UUFDaEIsQ0FBRSxPQUFPLEdBQUcsQ0FBQyxJQUFJLEtBQUssV0FBVyxJQUFJLE9BQU8sR0FBRyxDQUFDLElBQUksS0FBSyxRQUFR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QUMxQyxPQUFPLENBQ1AsR0FBRyxJQUFJLElBQUk7UUFDWCxPQUFPLEdBQUcsS0FBSyxRQUFRO1FBQ3JCLHFCQUFxQjtRQUNyQixDQUFFLE9BQU8sR0FBRyxDQUFDLFFBQVEsS0FBSyxXQUFXLElBQUksT0FBTyxHQUFHLENBQUMsUUFBUSxLQUFLLFNBQVMsQ0FBRTtRQUM1RSxzQkFBc0I7UUFDdEIsQ0FBRSxPQUFPLEdBQUcsQ0FBQyxVQUFVLEtBQUssV0FBVyxJQUFJLE9BQU8sR0FBRyxDQUFDLFVBQVUsS0FBSyxRQUFRLENBQUU7UUFFakYsSUFBSSxDQUNILENBQUM7QUFDRixDQUFDO0FBRUgsTUFBTSxVQUFVLHVCQUF1QixDQUFDLEdBQVE7SUFDOUMsT0FBTyxDQUNQLEdBQUcsSUFBSSxJQUFJO1FBQ1gsT0FBTyxHQUFHLEtBQUssUUFBUTtRQUNyQixzQkFBc0I7UUFDdEIsQ0FBRSxPQUFPLEdBQUcsQ0FBQyxVQUFVLEtBQUssU0FBUyxDQUFFO1FBQ3ZDLHNCQUFzQjtRQUN0QixDQUFFLE9BQU8sR0FBRyxDQUFDLFVBQVUsS0FBSyxXQUFXLElBQUksT0FBTyxHQUFHLENBQUMsVUFBVSxLQUFLLFFBQVEsQ0FBRTtRQUMvRSxnQkFBZ0I7UUFDaEIsQ0FBRSxPQUFPLEdBQUcsQ0FBQyxJQUFJLEtBQUssV0FBVyxJQUFJLE9BQU8sR0FBRyxDQUFDLElBQUksS0FBSyxRQUFRLENBQUU7UUFDbkUsZ0JBQWdCO1FBQ2hCLENBQUUsT0FBTyxHQUFHLENBQUMsSUFBSSxLQUFLLFdBQVcsSUFBSSxPQUFPLEdBQUcsQ0FBQyxJQUFJLEtBQUssUUFBUSxDQUFFO1FBQ25FLGtCQUFrQjtRQUNsQixDQUFFLE9BQU8sR0FBRyxDQUFDLE1BQU0sS0FBSyxXQUFXLElBQUksUUFBUSxDQUFDLEdBQUcsQ0FBQyxNQUFNLENBQUMsQ0FBRTtRQUUvRCxJQUFJLENBQ0gsQ0FBQztBQUNGLENBQUMiLCJzb3VyY2VzQ29udGVudCI6WyIvKiB0c2xpbnQ6ZGlzYWJsZSAqL1xuXG5pbXBvcnQgKiBhcyBtb2RlbHMgZnJvbSAnLi4vbW9kZWxzJztcblxuLyogcHJlLXByZXBhcmVkIGd1YXJkcyBmb3IgYnVpbGQgaW4gY29tcGxleCB0eXBlcyAqL1xuXG5mdW5jdGlvbiBfaXNCbG9iKGFyZzogYW55KTogYXJnIGlzIEJsb2Ige1xuICByZXR1cm4gYXJnICE9IG51bGwgJiYgdHlwZW9mIGFyZy5zaXplID09PSAnbnVtYmVyJyAmJiB0eXBlb2YgYXJnLnR5cGUgPT09ICdzdHJpbmcnICYmIHR5cGVvZiBhcmcuc2xpY2UgPT09ICdmdW5jdGlvbic7XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0ZpbGUoYXJnOiBhbnkpOiBhcmcgaXMgRmlsZSB7XG5yZXR1cm4gYXJnICE9IG51bGwgJiYgdHlwZW9mIGFyZy5sYXN0TW9kaWZpZWQgPT09ICdudW1iZXInICYmIHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgJiYgX2lzQmxvYihhcmcpO1xufVxuXG4vKiBnZW5lcmF0ZWQgdHlwZSBndWFyZHMgKi9cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQXZpQ2xvdWQoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkF2aUNsb3VkIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBsb2NhdGlvbj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sb2NhdGlvbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sb2NhdGlvbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5uYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gdXVpZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy51dWlkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnV1aWQgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQXZpQ29uZmlnKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BdmlDb25maWcge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGNhX2NlcnQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2FfY2VydCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jYV9jZXJ0ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY2xvdWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2xvdWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2xvdWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjb250cm9sbGVyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xsZXIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY29udHJvbGxlciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNvbnRyb2xQbGFuZUhhUHJvdmlkZXI/OiBib29sZWFuXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZUhhUHJvdmlkZXIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY29udHJvbFBsYW5lSGFQcm92aWRlciA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBsYWJlbHM/OiB7IFtrZXk6IHN0cmluZ106IHN0cmluZyB9XG4gICAgKCB0eXBlb2YgYXJnLmxhYmVscyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sYWJlbHMgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBtYW5hZ2VtZW50Q2x1c3RlclZpcE5ldHdvcmtDaWRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm1hbmFnZW1lbnRDbHVzdGVyVmlwTmV0d29ya0NpZHIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubWFuYWdlbWVudENsdXN0ZXJWaXBOZXR3b3JrQ2lkciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG1hbmFnZW1lbnRDbHVzdGVyVmlwTmV0d29ya05hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubWFuYWdlbWVudENsdXN0ZXJWaXBOZXR3b3JrTmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5tYW5hZ2VtZW50Q2x1c3RlclZpcE5ldHdvcmtOYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbmV0d29yaz86IEF2aU5ldHdvcmtQYXJhbXNcbiAgICAoIHR5cGVvZiBhcmcubmV0d29yayA9PT0gJ3VuZGVmaW5lZCcgfHwgaXNBdmlOZXR3b3JrUGFyYW1zKGFyZy5uZXR3b3JrKSApICYmXG4gICAgLy8gcGFzc3dvcmQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucGFzc3dvcmQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcucGFzc3dvcmQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBzZXJ2aWNlX2VuZ2luZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5zZXJ2aWNlX2VuZ2luZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5zZXJ2aWNlX2VuZ2luZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHVzZXJuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnVzZXJuYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnVzZXJuYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0F2aUNvbnRyb2xsZXJQYXJhbXMoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkF2aUNvbnRyb2xsZXJQYXJhbXMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIENBRGF0YT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5DQURhdGEgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuQ0FEYXRhID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gaG9zdD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5ob3N0ID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmhvc3QgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBwYXNzd29yZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5wYXNzd29yZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5wYXNzd29yZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHRlbmFudD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy50ZW5hbnQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcudGVuYW50ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gdXNlcm5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudXNlcm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcudXNlcm5hbWUgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQXZpTmV0d29ya1BhcmFtcyhhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQXZpTmV0d29ya1BhcmFtcyB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gY2lkcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jaWRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNpZHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBdmlTZXJ2aWNlRW5naW5lR3JvdXAoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkF2aVNlcnZpY2VFbmdpbmVHcm91cCB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbG9jYXRpb24/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubG9jYXRpb24gPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubG9jYXRpb24gPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHV1aWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudXVpZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy51dWlkID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0F2aVN1Ym5ldChhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQXZpU3VibmV0IHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBmYW1pbHk/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuZmFtaWx5ID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmZhbWlseSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHN1Ym5ldD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5zdWJuZXQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuc3VibmV0ID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0F2aVZpcE5ldHdvcmsoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkF2aVZpcE5ldHdvcmsge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGNsb3VkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNsb3VkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNsb3VkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY29uZmlnZWRTdWJuZXRzPzogQXZpU3VibmV0W11cbiAgICAoIHR5cGVvZiBhcmcuY29uZmlnZWRTdWJuZXRzID09PSAndW5kZWZpbmVkJyB8fCAoQXJyYXkuaXNBcnJheShhcmcuY29uZmlnZWRTdWJuZXRzKSAmJiBhcmcuY29uZmlnZWRTdWJuZXRzLmV2ZXJ5KChpdGVtOiB1bmtub3duKSA9PiBpc0F2aVN1Ym5ldChpdGVtKSkpICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHV1aWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudXVpZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy51dWlkID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0FXU0FjY291bnRQYXJhbXMoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkFXU0FjY291bnRQYXJhbXMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGFjY2Vzc0tleUlEPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmFjY2Vzc0tleUlEID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmFjY2Vzc0tleUlEID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gcHJvZmlsZU5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucHJvZmlsZU5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcucHJvZmlsZU5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyByZWdpb24/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucmVnaW9uID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnJlZ2lvbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHNlY3JldEFjY2Vzc0tleT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5zZWNyZXRBY2Nlc3NLZXkgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuc2VjcmV0QWNjZXNzS2V5ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gc2Vzc2lvblRva2VuPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnNlc3Npb25Ub2tlbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5zZXNzaW9uVG9rZW4gPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQVdTQXZhaWxhYmlsaXR5Wm9uZShhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQVdTQXZhaWxhYmlsaXR5Wm9uZSB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gaWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuaWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBV1NOb2RlQXooYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkFXU05vZGVBeiB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5uYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBwcml2YXRlU3VibmV0SUQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucHJpdmF0ZVN1Ym5ldElEID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnByaXZhdGVTdWJuZXRJRCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHB1YmxpY1N1Ym5ldElEPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnB1YmxpY1N1Ym5ldElEID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnB1YmxpY1N1Ym5ldElEID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gd29ya2VyTm9kZVR5cGU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcud29ya2VyTm9kZVR5cGUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcud29ya2VyTm9kZVR5cGUgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQVdTUmVnaW9uYWxDbHVzdGVyUGFyYW1zKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BV1NSZWdpb25hbENsdXN0ZXJQYXJhbXMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGFubm90YXRpb25zPzogeyBba2V5OiBzdHJpbmddOiBzdHJpbmcgfVxuICAgICggdHlwZW9mIGFyZy5hbm5vdGF0aW9ucyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5hbm5vdGF0aW9ucyA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGF3c0FjY291bnRQYXJhbXM/OiBBV1NBY2NvdW50UGFyYW1zXG4gICAgKCB0eXBlb2YgYXJnLmF3c0FjY291bnRQYXJhbXMgPT09ICd1bmRlZmluZWQnIHx8IGlzQVdTQWNjb3VudFBhcmFtcyhhcmcuYXdzQWNjb3VudFBhcmFtcykgKSAmJlxuICAgIC8vIGJhc3Rpb25Ib3N0RW5hYmxlZD86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuYmFzdGlvbkhvc3RFbmFibGVkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmJhc3Rpb25Ib3N0RW5hYmxlZCA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBjZWlwT3B0SW4/OiBib29sZWFuXG4gICAgKCB0eXBlb2YgYXJnLmNlaXBPcHRJbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jZWlwT3B0SW4gPT09ICdib29sZWFuJyApICYmXG4gICAgLy8gY2x1c3Rlck5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2x1c3Rlck5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2x1c3Rlck5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjb250cm9sUGxhbmVGbGF2b3I/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY29udHJvbFBsYW5lRmxhdm9yID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZUZsYXZvciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNvbnRyb2xQbGFuZU5vZGVUeXBlPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZU5vZGVUeXBlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZU5vZGVUeXBlID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY3JlYXRlQ2xvdWRGb3JtYXRpb25TdGFjaz86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuY3JlYXRlQ2xvdWRGb3JtYXRpb25TdGFjayA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jcmVhdGVDbG91ZEZvcm1hdGlvblN0YWNrID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIGVuYWJsZUF1ZGl0TG9nZ2luZz86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuZW5hYmxlQXVkaXRMb2dnaW5nID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmVuYWJsZUF1ZGl0TG9nZ2luZyA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBpZGVudGl0eU1hbmFnZW1lbnQ/OiBJZGVudGl0eU1hbmFnZW1lbnRDb25maWdcbiAgICAoIHR5cGVvZiBhcmcuaWRlbnRpdHlNYW5hZ2VtZW50ID09PSAndW5kZWZpbmVkJyB8fCBpc0lkZW50aXR5TWFuYWdlbWVudENvbmZpZyhhcmcuaWRlbnRpdHlNYW5hZ2VtZW50KSApICYmXG4gICAgLy8ga3ViZXJuZXRlc1ZlcnNpb24/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcua3ViZXJuZXRlc1ZlcnNpb24gPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcua3ViZXJuZXRlc1ZlcnNpb24gPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsYWJlbHM/OiB7IFtrZXk6IHN0cmluZ106IHN0cmluZyB9XG4gICAgKCB0eXBlb2YgYXJnLmxhYmVscyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sYWJlbHMgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsb2FkYmFsYW5jZXJTY2hlbWVJbnRlcm5hbD86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcubG9hZGJhbGFuY2VyU2NoZW1lSW50ZXJuYWwgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubG9hZGJhbGFuY2VyU2NoZW1lSW50ZXJuYWwgPT09ICdib29sZWFuJyApICYmXG4gICAgLy8gbWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZD86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcubWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5tYWNoaW5lSGVhbHRoQ2hlY2tFbmFibGVkID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIG5ldHdvcmtpbmc/OiBUS0dOZXR3b3JrXG4gICAgKCB0eXBlb2YgYXJnLm5ldHdvcmtpbmcgPT09ICd1bmRlZmluZWQnIHx8IGlzVEtHTmV0d29yayhhcmcubmV0d29ya2luZykgKSAmJlxuICAgIC8vIG51bU9mV29ya2VyTm9kZT86IG51bWJlclxuICAgICggdHlwZW9mIGFyZy5udW1PZldvcmtlck5vZGUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubnVtT2ZXb3JrZXJOb2RlID09PSAnbnVtYmVyJyApICYmXG4gICAgLy8gb3M/OiBBV1NWaXJ0dWFsTWFjaGluZVxuICAgICggdHlwZW9mIGFyZy5vcyA9PT0gJ3VuZGVmaW5lZCcgfHwgaXNBV1NWaXJ0dWFsTWFjaGluZShhcmcub3MpICkgJiZcbiAgICAvLyBzc2hLZXlOYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnNzaEtleU5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuc3NoS2V5TmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHZwYz86IEFXU1ZwY1xuICAgICggdHlwZW9mIGFyZy52cGMgPT09ICd1bmRlZmluZWQnIHx8IGlzQVdTVnBjKGFyZy52cGMpICkgJiZcbiAgICAvLyB3b3JrZXJOb2RlVHlwZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy53b3JrZXJOb2RlVHlwZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy53b3JrZXJOb2RlVHlwZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBV1NSb3V0ZShhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQVdTUm91dGUge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIERlc3RpbmF0aW9uQ2lkckJsb2NrPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLkRlc3RpbmF0aW9uQ2lkckJsb2NrID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLkRlc3RpbmF0aW9uQ2lkckJsb2NrID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gR2F0ZXdheUlkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLkdhdGV3YXlJZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5HYXRld2F5SWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBTdGF0ZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5TdGF0ZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5TdGF0ZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBV1NSb3V0ZVRhYmxlKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BV1NSb3V0ZVRhYmxlIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBpZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5pZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5pZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHJvdXRlcz86IEFXU1JvdXRlW11cbiAgICAoIHR5cGVvZiBhcmcucm91dGVzID09PSAndW5kZWZpbmVkJyB8fCAoQXJyYXkuaXNBcnJheShhcmcucm91dGVzKSAmJiBhcmcucm91dGVzLmV2ZXJ5KChpdGVtOiB1bmtub3duKSA9PiBpc0FXU1JvdXRlKGl0ZW0pKSkgKSAmJlxuICAgIC8vIHZwY0lkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnZwY0lkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnZwY0lkID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0FXU1N1Ym5ldChhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQVdTU3VibmV0IHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBhdmFpbGFiaWxpdHlab25lSWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuYXZhaWxhYmlsaXR5Wm9uZUlkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmF2YWlsYWJpbGl0eVpvbmVJZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGF2YWlsYWJpbGl0eVpvbmVOYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmF2YWlsYWJpbGl0eVpvbmVOYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmF2YWlsYWJpbGl0eVpvbmVOYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY2lkcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jaWRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNpZHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBpZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5pZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5pZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGlzUHVibGljOiBib29sZWFuXG4gICAgKCB0eXBlb2YgYXJnLmlzUHVibGljID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIHN0YXRlPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnN0YXRlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnN0YXRlID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gdnBjSWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudnBjSWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcudnBjSWQgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQVdTVmlydHVhbE1hY2hpbmUoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkFXU1ZpcnR1YWxNYWNoaW5lIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG9zSW5mbz86IE9TSW5mb1xuICAgICggdHlwZW9mIGFyZy5vc0luZm8gPT09ICd1bmRlZmluZWQnIHx8IGlzT1NJbmZvKGFyZy5vc0luZm8pICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQVdTVnBjKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BV1NWcGMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGF6cz86IEFXU05vZGVBeltdXG4gICAgKCB0eXBlb2YgYXJnLmF6cyA9PT0gJ3VuZGVmaW5lZCcgfHwgKEFycmF5LmlzQXJyYXkoYXJnLmF6cykgJiYgYXJnLmF6cy5ldmVyeSgoaXRlbTogdW5rbm93bikgPT4gaXNBV1NOb2RlQXooaXRlbSkpKSApICYmXG4gICAgLy8gY2lkcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jaWRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNpZHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB2cGNJRD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy52cGNJRCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy52cGNJRCA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBenVyZUFjY291bnRQYXJhbXMoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkF6dXJlQWNjb3VudFBhcmFtcyB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gYXp1cmVDbG91ZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5henVyZUNsb3VkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmF6dXJlQ2xvdWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjbGllbnRJZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jbGllbnRJZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jbGllbnRJZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNsaWVudFNlY3JldD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jbGllbnRTZWNyZXQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2xpZW50U2VjcmV0ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gc3Vic2NyaXB0aW9uSWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuc3Vic2NyaXB0aW9uSWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuc3Vic2NyaXB0aW9uSWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB0ZW5hbnRJZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy50ZW5hbnRJZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy50ZW5hbnRJZCA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBenVyZUluc3RhbmNlVHlwZShhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQXp1cmVJbnN0YW5jZVR5cGUge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGZhbWlseT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5mYW1pbHkgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuZmFtaWx5ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5uYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBzaXplPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnNpemUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuc2l6ZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHRpZXI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudGllciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy50aWVyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gem9uZXM/OiBzdHJpbmdbXVxuICAgICggdHlwZW9mIGFyZy56b25lcyA9PT0gJ3VuZGVmaW5lZCcgfHwgKEFycmF5LmlzQXJyYXkoYXJnLnpvbmVzKSAmJiBhcmcuem9uZXMuZXZlcnkoKGl0ZW06IHVua25vd24pID0+IHR5cGVvZiBpdGVtID09PSAnc3RyaW5nJykpICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQXp1cmVMb2NhdGlvbihhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQXp1cmVMb2NhdGlvbiB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gZGlzcGxheU5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuZGlzcGxheU5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuZGlzcGxheU5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBenVyZVJlZ2lvbmFsQ2x1c3RlclBhcmFtcyhhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQXp1cmVSZWdpb25hbENsdXN0ZXJQYXJhbXMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGFubm90YXRpb25zPzogeyBba2V5OiBzdHJpbmddOiBzdHJpbmcgfVxuICAgICggdHlwZW9mIGFyZy5hbm5vdGF0aW9ucyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5hbm5vdGF0aW9ucyA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGF6dXJlQWNjb3VudFBhcmFtcz86IEF6dXJlQWNjb3VudFBhcmFtc1xuICAgICggdHlwZW9mIGFyZy5henVyZUFjY291bnRQYXJhbXMgPT09ICd1bmRlZmluZWQnIHx8IGlzQXp1cmVBY2NvdW50UGFyYW1zKGFyZy5henVyZUFjY291bnRQYXJhbXMpICkgJiZcbiAgICAvLyBjZWlwT3B0SW4/OiBib29sZWFuXG4gICAgKCB0eXBlb2YgYXJnLmNlaXBPcHRJbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jZWlwT3B0SW4gPT09ICdib29sZWFuJyApICYmXG4gICAgLy8gY2x1c3Rlck5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2x1c3Rlck5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2x1c3Rlck5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjb250cm9sUGxhbmVGbGF2b3I/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY29udHJvbFBsYW5lRmxhdm9yID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZUZsYXZvciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNvbnRyb2xQbGFuZU1hY2hpbmVUeXBlPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZU1hY2hpbmVUeXBlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZU1hY2hpbmVUeXBlID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY29udHJvbFBsYW5lU3VibmV0Pzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZVN1Ym5ldCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jb250cm9sUGxhbmVTdWJuZXQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjb250cm9sUGxhbmVTdWJuZXRDaWRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZVN1Ym5ldENpZHIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY29udHJvbFBsYW5lU3VibmV0Q2lkciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGVuYWJsZUF1ZGl0TG9nZ2luZz86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuZW5hYmxlQXVkaXRMb2dnaW5nID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmVuYWJsZUF1ZGl0TG9nZ2luZyA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBmcm9udGVuZFByaXZhdGVJcD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5mcm9udGVuZFByaXZhdGVJcCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5mcm9udGVuZFByaXZhdGVJcCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGlkZW50aXR5TWFuYWdlbWVudD86IElkZW50aXR5TWFuYWdlbWVudENvbmZpZ1xuICAgICggdHlwZW9mIGFyZy5pZGVudGl0eU1hbmFnZW1lbnQgPT09ICd1bmRlZmluZWQnIHx8IGlzSWRlbnRpdHlNYW5hZ2VtZW50Q29uZmlnKGFyZy5pZGVudGl0eU1hbmFnZW1lbnQpICkgJiZcbiAgICAvLyBpc1ByaXZhdGVDbHVzdGVyPzogYm9vbGVhblxuICAgICggdHlwZW9mIGFyZy5pc1ByaXZhdGVDbHVzdGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmlzUHJpdmF0ZUNsdXN0ZXIgPT09ICdib29sZWFuJyApICYmXG4gICAgLy8ga3ViZXJuZXRlc1ZlcnNpb24/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcua3ViZXJuZXRlc1ZlcnNpb24gPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcua3ViZXJuZXRlc1ZlcnNpb24gPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsYWJlbHM/OiB7IFtrZXk6IHN0cmluZ106IHN0cmluZyB9XG4gICAgKCB0eXBlb2YgYXJnLmxhYmVscyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sYWJlbHMgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsb2NhdGlvbj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sb2NhdGlvbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sb2NhdGlvbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG1hY2hpbmVIZWFsdGhDaGVja0VuYWJsZWQ/OiBib29sZWFuXG4gICAgKCB0eXBlb2YgYXJnLm1hY2hpbmVIZWFsdGhDaGVja0VuYWJsZWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZCA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBuZXR3b3JraW5nPzogVEtHTmV0d29ya1xuICAgICggdHlwZW9mIGFyZy5uZXR3b3JraW5nID09PSAndW5kZWZpbmVkJyB8fCBpc1RLR05ldHdvcmsoYXJnLm5ldHdvcmtpbmcpICkgJiZcbiAgICAvLyBudW1PZldvcmtlck5vZGVzPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm51bU9mV29ya2VyTm9kZXMgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubnVtT2ZXb3JrZXJOb2RlcyA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG9zPzogQXp1cmVWaXJ0dWFsTWFjaGluZVxuICAgICggdHlwZW9mIGFyZy5vcyA9PT0gJ3VuZGVmaW5lZCcgfHwgaXNBenVyZVZpcnR1YWxNYWNoaW5lKGFyZy5vcykgKSAmJlxuICAgIC8vIHJlc291cmNlR3JvdXA/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucmVzb3VyY2VHcm91cCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5yZXNvdXJjZUdyb3VwID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gc3NoUHVibGljS2V5Pzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnNzaFB1YmxpY0tleSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5zc2hQdWJsaWNLZXkgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB2bmV0Q2lkcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy52bmV0Q2lkciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy52bmV0Q2lkciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHZuZXROYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnZuZXROYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnZuZXROYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gdm5ldFJlc291cmNlR3JvdXA/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudm5ldFJlc291cmNlR3JvdXAgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcudm5ldFJlc291cmNlR3JvdXAgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB3b3JrZXJNYWNoaW5lVHlwZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy53b3JrZXJNYWNoaW5lVHlwZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy53b3JrZXJNYWNoaW5lVHlwZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHdvcmtlck5vZGVTdWJuZXQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcud29ya2VyTm9kZVN1Ym5ldCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy53b3JrZXJOb2RlU3VibmV0ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gd29ya2VyTm9kZVN1Ym5ldENpZHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcud29ya2VyTm9kZVN1Ym5ldENpZHIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcud29ya2VyTm9kZVN1Ym5ldENpZHIgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQXp1cmVSZXNvdXJjZUdyb3VwKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BenVyZVJlc291cmNlR3JvdXAge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGlkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmlkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmlkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbG9jYXRpb246IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sb2NhdGlvbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG5hbWU6IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5uYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0F6dXJlU3VibmV0KGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BenVyZVN1Ym5ldCB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gY2lkcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jaWRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNpZHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNBenVyZVZpcnR1YWxNYWNoaW5lKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5BenVyZVZpcnR1YWxNYWNoaW5lIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG9zSW5mbz86IE9TSW5mb1xuICAgICggdHlwZW9mIGFyZy5vc0luZm8gPT09ICd1bmRlZmluZWQnIHx8IGlzT1NJbmZvKGFyZy5vc0luZm8pICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQXp1cmVWaXJ0dWFsTmV0d29yayhhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuQXp1cmVWaXJ0dWFsTmV0d29yayB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gY2lkckJsb2NrOiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2lkckJsb2NrID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gaWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuaWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsb2NhdGlvbjogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxvY2F0aW9uID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbmFtZTogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBzdWJuZXRzPzogQXp1cmVTdWJuZXRbXVxuICAgICggdHlwZW9mIGFyZy5zdWJuZXRzID09PSAndW5kZWZpbmVkJyB8fCAoQXJyYXkuaXNBcnJheShhcmcuc3VibmV0cykgJiYgYXJnLnN1Ym5ldHMuZXZlcnkoKGl0ZW06IHVua25vd24pID0+IGlzQXp1cmVTdWJuZXQoaXRlbSkpKSApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0NvbmZpZ0ZpbGUoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLkNvbmZpZ0ZpbGUge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGZpbGVjb250ZW50cz86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5maWxlY29udGVudHMgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuZmlsZWNvbnRlbnRzID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0NvbmZpZ0ZpbGVJbmZvKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5Db25maWdGaWxlSW5mbyB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gcGF0aD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5wYXRoID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnBhdGggPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzRG9ja2VyRGFlbW9uU3RhdHVzKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5Eb2NrZXJEYWVtb25TdGF0dXMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIHN0YXR1cz86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuc3RhdHVzID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnN0YXR1cyA9PT0gJ2Jvb2xlYW4nICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzRG9ja2VyUmVnaW9uYWxDbHVzdGVyUGFyYW1zKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5Eb2NrZXJSZWdpb25hbENsdXN0ZXJQYXJhbXMge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGFubm90YXRpb25zPzogeyBba2V5OiBzdHJpbmddOiBzdHJpbmcgfVxuICAgICggdHlwZW9mIGFyZy5hbm5vdGF0aW9ucyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5hbm5vdGF0aW9ucyA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNlaXBPcHRJbj86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuY2VpcE9wdEluID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNlaXBPcHRJbiA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBjbHVzdGVyTmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jbHVzdGVyTmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jbHVzdGVyTmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNvbnRyb2xQbGFuZUZsYXZvcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5jb250cm9sUGxhbmVGbGF2b3IgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY29udHJvbFBsYW5lRmxhdm9yID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gaWRlbnRpdHlNYW5hZ2VtZW50PzogSWRlbnRpdHlNYW5hZ2VtZW50Q29uZmlnXG4gICAgKCB0eXBlb2YgYXJnLmlkZW50aXR5TWFuYWdlbWVudCA9PT0gJ3VuZGVmaW5lZCcgfHwgaXNJZGVudGl0eU1hbmFnZW1lbnRDb25maWcoYXJnLmlkZW50aXR5TWFuYWdlbWVudCkgKSAmJlxuICAgIC8vIGt1YmVybmV0ZXNWZXJzaW9uPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmt1YmVybmV0ZXNWZXJzaW9uID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmt1YmVybmV0ZXNWZXJzaW9uID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGFiZWxzPzogeyBba2V5OiBzdHJpbmddOiBzdHJpbmcgfVxuICAgICggdHlwZW9mIGFyZy5sYWJlbHMgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGFiZWxzID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZD86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcubWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5tYWNoaW5lSGVhbHRoQ2hlY2tFbmFibGVkID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIG5ldHdvcmtpbmc/OiBUS0dOZXR3b3JrXG4gICAgKCB0eXBlb2YgYXJnLm5ldHdvcmtpbmcgPT09ICd1bmRlZmluZWQnIHx8IGlzVEtHTmV0d29yayhhcmcubmV0d29ya2luZykgKSAmJlxuICAgIC8vIG51bU9mV29ya2VyTm9kZXM/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubnVtT2ZXb3JrZXJOb2RlcyA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5udW1PZldvcmtlck5vZGVzID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0Vycm9yKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5FcnJvciB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbWVzc2FnZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5tZXNzYWdlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm1lc3NhZ2UgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzRmVhdHVyZU1hcChhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuRmVhdHVyZU1hcCB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gW2tleTogc3RyaW5nXTogc3RyaW5nXG4gICAgKCBPYmplY3QudmFsdWVzKGFyZykuZXZlcnkoKHZhbHVlOiB1bmtub3duKSA9PiB0eXBlb2YgdmFsdWUgPT09ICdzdHJpbmcnKSApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0ZlYXR1cmVzKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5GZWF0dXJlcyB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gW2tleTogc3RyaW5nXTogRmVhdHVyZU1hcFxuICAgICggT2JqZWN0LnZhbHVlcyhhcmcpLmV2ZXJ5KCh2YWx1ZTogdW5rbm93bikgPT4gaXNGZWF0dXJlTWFwKHZhbHVlKSkgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNIVFRQUHJveHlDb25maWd1cmF0aW9uKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5IVFRQUHJveHlDb25maWd1cmF0aW9uIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBlbmFibGVkPzogYm9vbGVhblxuICAgICggdHlwZW9mIGFyZy5lbmFibGVkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmVuYWJsZWQgPT09ICdib29sZWFuJyApICYmXG4gICAgLy8gSFRUUFByb3h5UGFzc3dvcmQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuSFRUUFByb3h5UGFzc3dvcmQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuSFRUUFByb3h5UGFzc3dvcmQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBIVFRQUHJveHlVUkw/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuSFRUUFByb3h5VVJMID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLkhUVFBQcm94eVVSTCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIEhUVFBQcm94eVVzZXJuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLkhUVFBQcm94eVVzZXJuYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLkhUVFBQcm94eVVzZXJuYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gSFRUUFNQcm94eVBhc3N3b3JkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLkhUVFBTUHJveHlQYXNzd29yZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5IVFRQU1Byb3h5UGFzc3dvcmQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBIVFRQU1Byb3h5VVJMPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLkhUVFBTUHJveHlVUkwgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuSFRUUFNQcm94eVVSTCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIEhUVFBTUHJveHlVc2VybmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5IVFRQU1Byb3h5VXNlcm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuSFRUUFNQcm94eVVzZXJuYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbm9Qcm94eT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5ub1Byb3h5ID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm5vUHJveHkgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzSWRlbnRpdHlNYW5hZ2VtZW50Q29uZmlnKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5JZGVudGl0eU1hbmFnZW1lbnRDb25maWcge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGlkbV90eXBlOiAnb2lkYycgfCAnbGRhcCcgfCAnbm9uZSdcbiAgICAoIFsnb2lkYycsICdsZGFwJywgJ25vbmUnXS5pbmNsdWRlcyhhcmcuaWRtX3R5cGUpICkgJiZcbiAgICAvLyBsZGFwX2JpbmRfZG4/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9iaW5kX2RuID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfYmluZF9kbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfYmluZF9wYXNzd29yZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX2JpbmRfcGFzc3dvcmQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF9iaW5kX3Bhc3N3b3JkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF9ncm91cF9zZWFyY2hfYmFzZV9kbj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX2dyb3VwX3NlYXJjaF9iYXNlX2RuID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX2Jhc2VfZG4gPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX2dyb3VwX3NlYXJjaF9maWx0ZXI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9ncm91cF9zZWFyY2hfZmlsdGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX2ZpbHRlciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfZ3JvdXBfc2VhcmNoX2dyb3VwX2F0dHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9ncm91cF9zZWFyY2hfZ3JvdXBfYXR0ciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX2dyb3VwX3NlYXJjaF9ncm91cF9hdHRyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF9ncm91cF9zZWFyY2hfbmFtZV9hdHRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX25hbWVfYXR0ciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX2dyb3VwX3NlYXJjaF9uYW1lX2F0dHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX2dyb3VwX3NlYXJjaF91c2VyX2F0dHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9ncm91cF9zZWFyY2hfdXNlcl9hdHRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX3VzZXJfYXR0ciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfcm9vdF9jYT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX3Jvb3RfY2EgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF9yb290X2NhID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF91cmw/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF91cmwgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF91cmwgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX3VzZXJfc2VhcmNoX2Jhc2VfZG4/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF9iYXNlX2RuID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfYmFzZV9kbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfdXNlcl9zZWFyY2hfZW1haWxfYXR0cj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX2VtYWlsX2F0dHIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF9lbWFpbF9hdHRyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF91c2VyX3NlYXJjaF9maWx0ZXI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF9maWx0ZXIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF9maWx0ZXIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX3VzZXJfc2VhcmNoX2lkX2F0dHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF9pZF9hdHRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfaWRfYXR0ciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfdXNlcl9zZWFyY2hfbmFtZV9hdHRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfbmFtZV9hdHRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfbmFtZV9hdHRyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF91c2VyX3NlYXJjaF91c2VybmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX3VzZXJuYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfdXNlcm5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBvaWRjX2NsYWltX21hcHBpbmdzPzogeyBba2V5OiBzdHJpbmddOiBzdHJpbmcgfVxuICAgICggdHlwZW9mIGFyZy5vaWRjX2NsYWltX21hcHBpbmdzID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm9pZGNfY2xhaW1fbWFwcGluZ3MgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBvaWRjX2NsaWVudF9pZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5vaWRjX2NsaWVudF9pZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5vaWRjX2NsaWVudF9pZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG9pZGNfY2xpZW50X3NlY3JldD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5vaWRjX2NsaWVudF9zZWNyZXQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcub2lkY19jbGllbnRfc2VjcmV0ID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gb2lkY19wcm92aWRlcl9uYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm9pZGNfcHJvdmlkZXJfbmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5vaWRjX3Byb3ZpZGVyX25hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBvaWRjX3Byb3ZpZGVyX3VybD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5vaWRjX3Byb3ZpZGVyX3VybCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5vaWRjX3Byb3ZpZGVyX3VybCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG9pZGNfc2NvcGU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcub2lkY19zY29wZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5vaWRjX3Njb3BlID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gb2lkY19za2lwX3ZlcmlmeV9jZXJ0PzogYm9vbGVhblxuICAgICggdHlwZW9mIGFyZy5vaWRjX3NraXBfdmVyaWZ5X2NlcnQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcub2lkY19za2lwX3ZlcmlmeV9jZXJ0ID09PSAnYm9vbGVhbicgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNMZGFwUGFyYW1zKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5MZGFwUGFyYW1zIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBsZGFwX2JpbmRfZG4/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9iaW5kX2RuID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfYmluZF9kbiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfYmluZF9wYXNzd29yZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX2JpbmRfcGFzc3dvcmQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF9iaW5kX3Bhc3N3b3JkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF9ncm91cF9zZWFyY2hfYmFzZV9kbj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX2dyb3VwX3NlYXJjaF9iYXNlX2RuID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX2Jhc2VfZG4gPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX2dyb3VwX3NlYXJjaF9maWx0ZXI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9ncm91cF9zZWFyY2hfZmlsdGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX2ZpbHRlciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfZ3JvdXBfc2VhcmNoX2dyb3VwX2F0dHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9ncm91cF9zZWFyY2hfZ3JvdXBfYXR0ciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX2dyb3VwX3NlYXJjaF9ncm91cF9hdHRyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF9ncm91cF9zZWFyY2hfbmFtZV9hdHRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX25hbWVfYXR0ciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX2dyb3VwX3NlYXJjaF9uYW1lX2F0dHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX2dyb3VwX3NlYXJjaF91c2VyX2F0dHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF9ncm91cF9zZWFyY2hfdXNlcl9hdHRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfZ3JvdXBfc2VhcmNoX3VzZXJfYXR0ciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfcm9vdF9jYT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX3Jvb3RfY2EgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGRhcF9yb290X2NhID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF90ZXN0X2dyb3VwPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfdGVzdF9ncm91cCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX3Rlc3RfZ3JvdXAgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX3Rlc3RfdXNlcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX3Rlc3RfdXNlciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX3Rlc3RfdXNlciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfdXJsPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfdXJsID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXJsID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF91c2VyX3NlYXJjaF9iYXNlX2RuPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfYmFzZV9kbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX2Jhc2VfZG4gPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX3VzZXJfc2VhcmNoX2VtYWlsX2F0dHI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF9lbWFpbF9hdHRyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfZW1haWxfYXR0ciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfdXNlcl9zZWFyY2hfZmlsdGVyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfZmlsdGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfZmlsdGVyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGRhcF91c2VyX3NlYXJjaF9pZF9hdHRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmxkYXBfdXNlcl9zZWFyY2hfaWRfYXR0ciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX2lkX2F0dHIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBsZGFwX3VzZXJfc2VhcmNoX25hbWVfYXR0cj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX25hbWVfYXR0ciA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX25hbWVfYXR0ciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGxkYXBfdXNlcl9zZWFyY2hfdXNlcm5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubGRhcF91c2VyX3NlYXJjaF91c2VybmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5sZGFwX3VzZXJfc2VhcmNoX3VzZXJuYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0xkYXBUZXN0UmVzdWx0KGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5MZGFwVGVzdFJlc3VsdCB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gY29kZT86IG51bWJlclxuICAgICggdHlwZW9mIGFyZy5jb2RlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvZGUgPT09ICdudW1iZXInICkgJiZcbiAgICAvLyBkZXNjPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmRlc2MgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuZGVzYyA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNOb2RlVHlwZShhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuTm9kZVR5cGUge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGNwdT86IG51bWJlclxuICAgICggdHlwZW9mIGFyZy5jcHUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY3B1ID09PSAnbnVtYmVyJyApICYmXG4gICAgLy8gZGlzaz86IG51bWJlclxuICAgICggdHlwZW9mIGFyZy5kaXNrID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmRpc2sgPT09ICdudW1iZXInICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHJhbT86IG51bWJlclxuICAgICggdHlwZW9mIGFyZy5yYW0gPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcucmFtID09PSAnbnVtYmVyJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc09TSW5mbyhhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuT1NJbmZvIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBhcmNoPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmFyY2ggPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuYXJjaCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5uYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gdmVyc2lvbj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy52ZXJzaW9uID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnZlcnNpb24gPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzUHJvdmlkZXJJbmZvKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5Qcm92aWRlckluZm8ge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIHByb3ZpZGVyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnByb3ZpZGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnByb3ZpZGVyID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gdGtyVmVyc2lvbj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy50a3JWZXJzaW9uID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnRrclZlcnNpb24gPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzVEtHTmV0d29yayhhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuVEtHTmV0d29yayB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gY2x1c3RlckROU05hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2x1c3RlckROU05hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2x1c3RlckROU05hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjbHVzdGVyTm9kZUNJRFI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY2x1c3Rlck5vZGVDSURSID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNsdXN0ZXJOb2RlQ0lEUiA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNsdXN0ZXJQb2RDSURSPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNsdXN0ZXJQb2RDSURSID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNsdXN0ZXJQb2RDSURSID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY2x1c3RlclNlcnZpY2VDSURSPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNsdXN0ZXJTZXJ2aWNlQ0lEUiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5jbHVzdGVyU2VydmljZUNJRFIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjbmlUeXBlPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNuaVR5cGUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY25pVHlwZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGh0dHBQcm94eUNvbmZpZ3VyYXRpb24/OiBIVFRQUHJveHlDb25maWd1cmF0aW9uXG4gICAgKCB0eXBlb2YgYXJnLmh0dHBQcm94eUNvbmZpZ3VyYXRpb24gPT09ICd1bmRlZmluZWQnIHx8IGlzSFRUUFByb3h5Q29uZmlndXJhdGlvbihhcmcuaHR0cFByb3h5Q29uZmlndXJhdGlvbikgKSAmJlxuICAgIC8vIG5ldHdvcmtOYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5ldHdvcmtOYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm5ldHdvcmtOYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZwYyhhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuVnBjIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBjaWRyPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNpZHIgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2lkciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGlkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmlkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmlkID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZTcGhlcmVBdmFpbGFiaWxpdHlab25lKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5WU3BoZXJlQXZhaWxhYmlsaXR5Wm9uZSB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbW9pZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5tb2lkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm1vaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNWU3BoZXJlQ3JlZGVudGlhbHMoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLlZTcGhlcmVDcmVkZW50aWFscyB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gaG9zdD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5ob3N0ID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmhvc3QgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBpbnNlY3VyZT86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuaW5zZWN1cmUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuaW5zZWN1cmUgPT09ICdib29sZWFuJyApICYmXG4gICAgLy8gcGFzc3dvcmQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucGFzc3dvcmQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcucGFzc3dvcmQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB0aHVtYnByaW50Pzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnRodW1icHJpbnQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcudGh1bWJwcmludCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHVzZXJuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnVzZXJuYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnVzZXJuYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZTcGhlcmVEYXRhY2VudGVyKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5WU3BoZXJlRGF0YWNlbnRlciB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbW9pZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5tb2lkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm1vaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNWU3BoZXJlRGF0YXN0b3JlKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5WU3BoZXJlRGF0YXN0b3JlIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBtb2lkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm1vaWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubW9pZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5uYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZTcGhlcmVGb2xkZXIoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLlZTcGhlcmVGb2xkZXIge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIG1vaWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubW9pZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5tb2lkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5uYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm5hbWUgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzVnNwaGVyZUluZm8oYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLlZzcGhlcmVJbmZvIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBoYXNQYWNpZmljPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmhhc1BhY2lmaWMgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuaGFzUGFjaWZpYyA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHZlcnNpb24/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcudmVyc2lvbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy52ZXJzaW9uID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZTcGhlcmVNYW5hZ2VtZW50T2JqZWN0KGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5WU3BoZXJlTWFuYWdlbWVudE9iamVjdCB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbW9pZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5tb2lkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm1vaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIHBhcmVudE1vaWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcucGFyZW50TW9pZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5wYXJlbnRNb2lkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gcGF0aD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5wYXRoID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnBhdGggPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyByZXNvdXJjZVR5cGU/OiAnZGF0YWNlbnRlcicgfCAnY2x1c3RlcicgfCAnaG9zdGdyb3VwJyB8ICdmb2xkZXInIHwgJ3Jlc3Bvb2wnIHwgJ3ZtJyB8ICdkYXRhc3RvcmUnIHwgJ2hvc3QnIHwgJ25ldHdvcmsnXG4gICAgKCB0eXBlb2YgYXJnLnJlc291cmNlVHlwZSA9PT0gJ3VuZGVmaW5lZCcgfHwgWydkYXRhY2VudGVyJywgJ2NsdXN0ZXInLCAnaG9zdGdyb3VwJywgJ2ZvbGRlcicsICdyZXNwb29sJywgJ3ZtJywgJ2RhdGFzdG9yZScsICdob3N0JywgJ25ldHdvcmsnXS5pbmNsdWRlcyhhcmcucmVzb3VyY2VUeXBlKSApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZTcGhlcmVOZXR3b3JrKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5WU3BoZXJlTmV0d29yayB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gZGlzcGxheU5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuZGlzcGxheU5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuZGlzcGxheU5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBtb2lkPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm1vaWQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubW9pZCA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG5hbWU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubmFtZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5uYW1lID09PSAnc3RyaW5nJyApICYmXG5cbiAgdHJ1ZVxuICApO1xuICB9XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1ZTcGhlcmVSZWdpb24oYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLlZTcGhlcmVSZWdpb24ge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIG1vaWQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcubW9pZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5tb2lkID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbmFtZT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5uYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm5hbWUgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB6b25lcz86IFZTcGhlcmVBdmFpbGFiaWxpdHlab25lW11cbiAgICAoIHR5cGVvZiBhcmcuem9uZXMgPT09ICd1bmRlZmluZWQnIHx8IChBcnJheS5pc0FycmF5KGFyZy56b25lcykgJiYgYXJnLnpvbmVzLmV2ZXJ5KChpdGVtOiB1bmtub3duKSA9PiBpc1ZTcGhlcmVBdmFpbGFiaWxpdHlab25lKGl0ZW0pKSkgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNWc3BoZXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zKGFyZzogYW55KTogYXJnIGlzIG1vZGVscy5Wc3BoZXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zIHtcbiAgcmV0dXJuIChcbiAgYXJnICE9IG51bGwgJiZcbiAgdHlwZW9mIGFyZyA9PT0gJ29iamVjdCcgJiZcbiAgICAvLyBhbm5vdGF0aW9ucz86IHsgW2tleTogc3RyaW5nXTogc3RyaW5nIH1cbiAgICAoIHR5cGVvZiBhcmcuYW5ub3RhdGlvbnMgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuYW5ub3RhdGlvbnMgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBhdmlDb25maWc/OiBBdmlDb25maWdcbiAgICAoIHR5cGVvZiBhcmcuYXZpQ29uZmlnID09PSAndW5kZWZpbmVkJyB8fCBpc0F2aUNvbmZpZyhhcmcuYXZpQ29uZmlnKSApICYmXG4gICAgLy8gY2VpcE9wdEluPzogYm9vbGVhblxuICAgICggdHlwZW9mIGFyZy5jZWlwT3B0SW4gPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY2VpcE9wdEluID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIGNsdXN0ZXJOYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNsdXN0ZXJOYW1lID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNsdXN0ZXJOYW1lID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gY29udHJvbFBsYW5lRW5kcG9pbnQ/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY29udHJvbFBsYW5lRW5kcG9pbnQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcuY29udHJvbFBsYW5lRW5kcG9pbnQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBjb250cm9sUGxhbmVGbGF2b3I/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuY29udHJvbFBsYW5lRmxhdm9yID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZUZsYXZvciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGNvbnRyb2xQbGFuZU5vZGVUeXBlPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZU5vZGVUeXBlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmNvbnRyb2xQbGFuZU5vZGVUeXBlID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gZGF0YWNlbnRlcj86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5kYXRhY2VudGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmRhdGFjZW50ZXIgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBkYXRhc3RvcmU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuZGF0YXN0b3JlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmRhdGFzdG9yZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGVuYWJsZUF1ZGl0TG9nZ2luZz86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcuZW5hYmxlQXVkaXRMb2dnaW5nID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmVuYWJsZUF1ZGl0TG9nZ2luZyA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyBmb2xkZXI/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuZm9sZGVyID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmZvbGRlciA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGlkZW50aXR5TWFuYWdlbWVudD86IElkZW50aXR5TWFuYWdlbWVudENvbmZpZ1xuICAgICggdHlwZW9mIGFyZy5pZGVudGl0eU1hbmFnZW1lbnQgPT09ICd1bmRlZmluZWQnIHx8IGlzSWRlbnRpdHlNYW5hZ2VtZW50Q29uZmlnKGFyZy5pZGVudGl0eU1hbmFnZW1lbnQpICkgJiZcbiAgICAvLyBpcEZhbWlseT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5pcEZhbWlseSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5pcEZhbWlseSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIGt1YmVybmV0ZXNWZXJzaW9uPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLmt1YmVybmV0ZXNWZXJzaW9uID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLmt1YmVybmV0ZXNWZXJzaW9uID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbGFiZWxzPzogeyBba2V5OiBzdHJpbmddOiBzdHJpbmcgfVxuICAgICggdHlwZW9mIGFyZy5sYWJlbHMgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubGFiZWxzID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZD86IGJvb2xlYW5cbiAgICAoIHR5cGVvZiBhcmcubWFjaGluZUhlYWx0aENoZWNrRW5hYmxlZCA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5tYWNoaW5lSGVhbHRoQ2hlY2tFbmFibGVkID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIG5ldHdvcmtpbmc/OiBUS0dOZXR3b3JrXG4gICAgKCB0eXBlb2YgYXJnLm5ldHdvcmtpbmcgPT09ICd1bmRlZmluZWQnIHx8IGlzVEtHTmV0d29yayhhcmcubmV0d29ya2luZykgKSAmJlxuICAgIC8vIG51bU9mV29ya2VyTm9kZT86IG51bWJlclxuICAgICggdHlwZW9mIGFyZy5udW1PZldvcmtlck5vZGUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubnVtT2ZXb3JrZXJOb2RlID09PSAnbnVtYmVyJyApICYmXG4gICAgLy8gb3M/OiBWU3BoZXJlVmlydHVhbE1hY2hpbmVcbiAgICAoIHR5cGVvZiBhcmcub3MgPT09ICd1bmRlZmluZWQnIHx8IGlzVlNwaGVyZVZpcnR1YWxNYWNoaW5lKGFyZy5vcykgKSAmJlxuICAgIC8vIHJlc291cmNlUG9vbD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5yZXNvdXJjZVBvb2wgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcucmVzb3VyY2VQb29sID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gc3NoX2tleT86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5zc2hfa2V5ID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLnNzaF9rZXkgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyB2c3BoZXJlQ3JlZGVudGlhbHM/OiBWU3BoZXJlQ3JlZGVudGlhbHNcbiAgICAoIHR5cGVvZiBhcmcudnNwaGVyZUNyZWRlbnRpYWxzID09PSAndW5kZWZpbmVkJyB8fCBpc1ZTcGhlcmVDcmVkZW50aWFscyhhcmcudnNwaGVyZUNyZWRlbnRpYWxzKSApICYmXG4gICAgLy8gd29ya2VyTm9kZVR5cGU/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcud29ya2VyTm9kZVR5cGUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcud29ya2VyTm9kZVR5cGUgPT09ICdzdHJpbmcnICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzVlNwaGVyZVJlc291cmNlUG9vbChhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuVlNwaGVyZVJlc291cmNlUG9vbCB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gbW9pZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5tb2lkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm1vaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNWU3BoZXJlVGh1bWJwcmludChhcmc6IGFueSk6IGFyZyBpcyBtb2RlbHMuVlNwaGVyZVRodW1icHJpbnQge1xuICByZXR1cm4gKFxuICBhcmcgIT0gbnVsbCAmJlxuICB0eXBlb2YgYXJnID09PSAnb2JqZWN0JyAmJlxuICAgIC8vIGluc2VjdXJlPzogYm9vbGVhblxuICAgICggdHlwZW9mIGFyZy5pbnNlY3VyZSA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5pbnNlY3VyZSA9PT0gJ2Jvb2xlYW4nICkgJiZcbiAgICAvLyB0aHVtYnByaW50Pzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLnRodW1icHJpbnQgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcudGh1bWJwcmludCA9PT0gJ3N0cmluZycgKSAmJlxuXG4gIHRydWVcbiAgKTtcbiAgfVxuXG5leHBvcnQgZnVuY3Rpb24gaXNWU3BoZXJlVmlydHVhbE1hY2hpbmUoYXJnOiBhbnkpOiBhcmcgaXMgbW9kZWxzLlZTcGhlcmVWaXJ0dWFsTWFjaGluZSB7XG4gIHJldHVybiAoXG4gIGFyZyAhPSBudWxsICYmXG4gIHR5cGVvZiBhcmcgPT09ICdvYmplY3QnICYmXG4gICAgLy8gaXNUZW1wbGF0ZTogYm9vbGVhblxuICAgICggdHlwZW9mIGFyZy5pc1RlbXBsYXRlID09PSAnYm9vbGVhbicgKSAmJlxuICAgIC8vIGs4c1ZlcnNpb24/OiBzdHJpbmdcbiAgICAoIHR5cGVvZiBhcmcuazhzVmVyc2lvbiA9PT0gJ3VuZGVmaW5lZCcgfHwgdHlwZW9mIGFyZy5rOHNWZXJzaW9uID09PSAnc3RyaW5nJyApICYmXG4gICAgLy8gbW9pZD86IHN0cmluZ1xuICAgICggdHlwZW9mIGFyZy5tb2lkID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgYXJnLm1vaWQgPT09ICdzdHJpbmcnICkgJiZcbiAgICAvLyBuYW1lPzogc3RyaW5nXG4gICAgKCB0eXBlb2YgYXJnLm5hbWUgPT09ICd1bmRlZmluZWQnIHx8IHR5cGVvZiBhcmcubmFtZSA9PT0gJ3N0cmluZycgKSAmJlxuICAgIC8vIG9zSW5mbz86IE9TSW5mb1xuICAgICggdHlwZW9mIGFyZy5vc0luZm8gPT09ICd1bmRlZmluZWQnIHx8IGlzT1NJbmZvKGFyZy5vc0luZm8pICkgJiZcblxuICB0cnVlXG4gICk7XG4gIH1cblxuXG4iXX0=