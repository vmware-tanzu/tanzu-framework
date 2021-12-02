export enum AwsStep {
    PROVIDER = 'provider',
    VPC = 'vpc',
    NODESETTING = 'nodeSetting',
    NETWORK = 'network',
    METADATA = 'metadata',
    IDENTITY = 'identity',
    OSIMAGE = 'osImage'
}
export enum AwsForm {
    PROVIDER = 'awsProviderForm',
    VPC = 'vpcForm',
    NODESETTING = 'awsNodeSettingForm',
    NETWORK = 'networkForm',
    METADATA = 'metadataForm',
    IDENTITY = 'identityForm',
    OSIMAGE = 'osImageForm'
}
export enum AwsField {
    NODESETTING_AZ_1 = 'awsNodeAz1',
    NODESETTING_AZ_2 = 'awsNodeAz2',
    NODESETTING_AZ_3 = 'awsNodeAz3',
    NODESETTING_BASTION_HOST_ENABLED = 'bastionHostEnabled',
    NODESETTING_CLUSTER_NAME = 'clusterName',
    NODESETTING_CONTROL_PLANE_SETTING = 'controlPlaneSetting',
    NODESETTING_CREATE_CLOUD_FORMATION = 'createCloudFormation',
    NODESETTING_INSTANCE_TYPE_DEV = 'devInstanceType',
    NODESETTING_INSTANCE_TYPE_PROD = 'prodInstanceType',
    NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED = 'machineHealthChecksEnabled',
    NODESETTING_VPC_PUBLIC_SUBNET_1 = 'vpcPublicSubnet1',
    NODESETTING_VPC_PUBLIC_SUBNET_2 = 'vpcPublicSubnet2',
    NODESETTING_VPC_PUBLIC_SUBNET_3 = 'vpcPublicSubnet3',
    NODESETTING_VPC_PRIVATE_SUBNET_1 = 'vpcPrivateSubnet1',
    NODESETTING_VPC_PRIVATE_SUBNET_2 = 'vpcPrivateSubnet2',
    NODESETTING_VPC_PRIVATE_SUBNET_3 = 'vpcPrivateSubnet3',
    NODESETTING_WORKERTYPE_1 = 'workerNodeInstanceType1',
    NODESETTING_WORKERTYPE_2 = 'workerNodeInstanceType2',
    NODESETTING_WORKERTYPE_3 = 'workerNodeInstanceType3',

    PROVIDER_ACCESS_KEY = 'accessKeyID',
    PROVIDER_AUTH_TYPE = 'authType',
    PROVIDER_PROFILE_NAME = 'profileName',
    PROVIDER_REGION = 'region',
    PROVIDER_SECRET_ACCESS_KEY = 'secretAccessKey',
    PROVIDER_SESSION_TOKEN = 'sessionToken',

    VPC_EXISTING_ID = 'existingVpcId',
    VPC_EXISTING_CIDR = 'existingVpcCidr',
    VPC_NEW_CIDR = 'vpc',
    VPC_NON_INTERNET_FACING = 'nonInternetFacingVPC',
    VPC_TYPE = 'vpcType'
}
