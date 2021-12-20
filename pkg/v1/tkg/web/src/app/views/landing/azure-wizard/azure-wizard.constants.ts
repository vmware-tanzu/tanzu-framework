export enum ResourceGroupOption {
    EXISTING = 'existing',
    CUSTOM = 'custom',
}
export enum AzureForm {
    PROVIDER = 'azureProviderForm',
    NODESETTING = 'azureNodeSettingForm',
    METADATA = 'metadataForm',
    NETWORK = 'networkForm',
    CEIP = 'ceipOptInForm',
    IDENTITY = 'identityForm',
    OSIMAGE = 'osImageForm',
    VNET = 'vnetForm'
}
export enum AzureField {
    NODESETTING_CONTROL_PLANE_SETTING = 'controlPlaneSetting',
    NODESETTING_INSTANCE_TYPE_DEV = 'devInstanceType',
    NODESETTING_INSTANCE_TYPE_PROD = 'prodInstanceType',
    NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED = 'machineHealthChecksEnabled',
    NODESETTING_MANAMGEMENT_CLUSTER_NAME = 'managementClusterName',
    NODESETTING_WORKERTYPE = 'workerNodeInstanceType',

/*
    NOTE: these enum values are used by backend endpoints, so do not change them:
    PROVIDER_AZURECLOUD
    PROVIDER_CLIENT,
    PROVIDER_CLIENTSECRET,
    PROVIDER_SUBSCRIPTION,
    PROVIDER_TENANT,
*/
    PROVIDER_AZURECLOUD = 'azureCloud',
    PROVIDER_CLIENT = 'clientId',
    PROVIDER_CLIENTSECRET = 'clientSecret',
    PROVIDER_REGION = 'region',
    PROVIDER_RESOURCEGROUPCUSTOM = 'resourceGroupCustom',
    PROVIDER_RESOURCEGROUPEXISTING = 'resourceGroupExisting',
    PROVIDER_RESOURCEGROUPOPTION = 'resourceGroupOption',
    PROVIDER_SSHPUBLICKEY = 'sshPublicKey',
    PROVIDER_SUBSCRIPTION = 'subscriptionId',
    PROVIDER_TENANT = 'tenantId',

    VNET_CUSTOM_NAME = 'vnetNameCustom',
    VNET_CUSTOM_CIDR = 'vnetCidrBlock',
    VNET_EXISTING_NAME = 'vnetNameExisting',
    VNET_EXISTING_OR_CUSTOM = 'vnetOption',
    VNET_PRIVATE_CLUSTER = 'privateAzureCluster',
    VNET_PRIVATE_IP = 'privateIP',
    VNET_RESOURCE_GROUP = 'vnetResourceGroup',
    // subnet fields:
    VNET_CONTROLPLANE_NEWSUBNET_CIDR = 'controlPlaneSubnetCidrNew',
    VNET_CONTROLPLANE_NEWSUBNET_NAME = 'controlPlaneSubnetNew',
    VNET_CONTROLPLANE_SUBNET_CIDR = 'controlPlaneSubnetCidr',
    VNET_CONTROLPLANE_SUBNET_NAME = 'controlPlaneSubnet',
    VNET_WORKER_SUBNET_NAME = 'workerNodeSubnet',
    VNET_WORKER_NEWSUBNET_CIDR = 'workerNodeSubnetCidrNew',
    VNET_WORKER_NEWSUBNET_NAME = 'workerNodeSubnetNew',
}
