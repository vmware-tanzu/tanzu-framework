import {StepMapping} from "../wizard/shared/FieldMapping";
import {SimpleValidator} from "../wizard/shared/constants/validation.constants";

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
    NODESETTING_SSH_KEY_NAME = 'sshKeyName',
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

export const AwsNodeSettingStepMapping: StepMapping = {
    name: AwsStep.NODESETTING,
    form: AwsForm.NODESETTING,
    fieldMappings: [
        { name: AwsField.NODESETTING_AZ_1, required: true },
        { name: AwsField.NODESETTING_AZ_2, required: true },
        { name: AwsField.NODESETTING_AZ_3, required: true },
        { name: AwsField.NODESETTING_BASTION_HOST_ENABLED, defaultValue: 'yes' },
        { name: AwsField.NODESETTING_CLUSTER_NAME, validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
        { name: AwsField.NODESETTING_CONTROL_PLANE_SETTING, required: true },
        { name: AwsField.NODESETTING_CREATE_CLOUD_FORMATION, isBoolean: true, defaultValue: true },
        { name: AwsField.NODESETTING_INSTANCE_TYPE_DEV, required: true },
        { name: AwsField.NODESETTING_INSTANCE_TYPE_PROD, required: true },
        { name: AwsField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED, isBoolean: true, defaultValue: true },
        { name: AwsField.NODESETTING_SSH_KEY_NAME, required: true },
        { name: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1 },
        { name: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2 },
        { name: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3 },
        { name: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1 },
        { name: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2 },
        { name: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3 },
        { name: AwsField.NODESETTING_WORKERTYPE_1 },
        { name: AwsField.NODESETTING_WORKERTYPE_2 },
        { name: AwsField.NODESETTING_WORKERTYPE_3 },
    ],
};
