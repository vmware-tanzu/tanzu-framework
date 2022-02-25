export enum CredentialType {
    ONETIME = 'oneTimeCredentials',
    PROFILE = 'credentialProfile'
}

export enum VpcType {
    EXISTING = 'existing',
    NEW = 'new'
}

export enum AwsForm {
    PROVIDER = 'awsProviderForm',
    VPC = 'vpcForm',
    NODESETTING = 'awsNodeSettingForm',
}

export enum AwsField {
    NODESETTING_AZ_1 = 'awsNodeAz1',
    NODESETTING_AZ_2 = 'awsNodeAz2',
    NODESETTING_AZ_3 = 'awsNodeAz3',
    NODESETTING_BASTION_HOST_ENABLED = 'bastionHostEnabled',
    NODESETTING_CREATE_CLOUD_FORMATION = 'createCloudFormation',
    NODESETTING_SSH_KEY_NAME = 'sshKeyName',
    NODESETTING_VPC_PUBLIC_SUBNET_1 = 'vpcPublicSubnet1',
    NODESETTING_VPC_PUBLIC_SUBNET_2 = 'vpcPublicSubnet2',
    NODESETTING_VPC_PUBLIC_SUBNET_3 = 'vpcPublicSubnet3',
    NODESETTING_VPC_PRIVATE_SUBNET_1 = 'vpcPrivateSubnet1',
    NODESETTING_VPC_PRIVATE_SUBNET_2 = 'vpcPrivateSubnet2',
    NODESETTING_VPC_PRIVATE_SUBNET_3 = 'vpcPrivateSubnet3',
    // NOTE: worker type 1 uses NodeSettingField.WORKER_NODE_INSTANCE_TYPE from the shared component
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
    VPC_PRIVATE_NODE_CIDR = 'privateNodeCidr',
    VPC_PUBLIC_NODE_CIDR = 'publicNodeCidr',
    VPC_TYPE = 'vpcType'
}
