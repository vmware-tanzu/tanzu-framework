import { AwsField } from '../aws-wizard.constants';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const AwsNodeSettingStepMapping: StepMapping = {
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
