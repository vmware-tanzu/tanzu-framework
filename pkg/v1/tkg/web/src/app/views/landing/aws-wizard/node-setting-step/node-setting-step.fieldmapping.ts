import { AwsField } from '../aws-wizard.constants';
import { NodeSettingField } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.fieldmapping';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const AwsNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        { name: AwsField.NODESETTING_AZ_1, required: true, label: 'AVAILABILITY ZONE 1', requiresBackendData: true },
        { name: AwsField.NODESETTING_AZ_2, required: true, label: 'AVAILABILITY ZONE 2', requiresBackendData: true },
        { name: AwsField.NODESETTING_AZ_3, required: true, label: 'AVAILABILITY ZONE 3', requiresBackendData: true },
        { name: AwsField.NODESETTING_BASTION_HOST_ENABLED, isBoolean: true, defaultValue: true, label: 'ENABLE BASTION HOST' },
        { name: AwsField.NODESETTING_CREATE_CLOUD_FORMATION, isBoolean: true, defaultValue: true,
            label: 'AUTOMATE CREATION OF AWS CLOUDFORMATION STACK' },
        { name: AwsField.NODESETTING_SSH_KEY_NAME, required: true, label: 'EC2 KEY PAIR' },
        { name: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1, label: 'VPC PUBLIC SUBNET 1', requiresBackendData: true },
        { name: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2, label: 'VPC PUBLIC SUBNET 2', requiresBackendData: true },
        { name: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3, label: 'VPC PUBLIC SUBNET 3', requiresBackendData: true },
        { name: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1, label: 'VPC PRIVATE SUBNET 1', requiresBackendData: true },
        { name: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2, label: 'VPC PRIVATE SUBNET 2', requiresBackendData: true },
        { name: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3, label: 'VPC PRIVATE SUBNET 3', requiresBackendData: true },
        { name: AwsField.NODESETTING_WORKERTYPE_2, label: 'AZ2 WORKER NODE INSTANCE TYPE', requiresBackendData: true },
        { name: AwsField.NODESETTING_WORKERTYPE_3, label: 'AZ3 WORKER NODE INSTANCE TYPE', requiresBackendData: true },
    ],
};

export const AwsFieldDisplayOrder: string[] = [
    NodeSettingField.CLUSTER_NAME,
    NodeSettingField.INSTANCE_TYPE_DEV,
    NodeSettingField.INSTANCE_TYPE_PROD,
    AwsField.NODESETTING_BASTION_HOST_ENABLED,
    AwsField.NODESETTING_CREATE_CLOUD_FORMATION,
    AwsField.NODESETTING_SSH_KEY_NAME,
    AwsField.NODESETTING_AZ_1,
    AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1,
    AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
    NodeSettingField.WORKER_NODE_INSTANCE_TYPE,
    AwsField.NODESETTING_AZ_2,
    AwsField.NODESETTING_VPC_PUBLIC_SUBNET_2,
    AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
    AwsField.NODESETTING_WORKERTYPE_2,
    AwsField.NODESETTING_AZ_3,
    AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3,
    AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
    AwsField.NODESETTING_WORKERTYPE_3,
    NodeSettingField.ENABLE_AUDIT_LOGGING,
    NodeSettingField.MACHINE_HEALTH_CHECKS_ENABLED,
]
// About AwsNodeSettingStep:
// This is a complex form. The first thing the user is expected to do is select DEV or PROD as NODESETTING_CONTROL_PLANE_SETTING, which
// will cascade into activating/deactivating other fields based on its value. Therefore, NODESETTING_CONTROL_PLANE_SETTING is set as a
// primaryTrigger and most of the other fields on the form are set doNotAutoRestore because their values depend either on the DEV/PROD
// setting and/or on backend data events.
