import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export enum NodeSettingField {
    CLUSTER_NAME = 'clusterName',
    INSTANCE_TYPE_DEV = 'devInstanceType',
    INSTANCE_TYPE_PROD = 'prodInstanceType',
    MACHINE_HEALTH_CHECKS_ENABLED = 'machineHealthChecksEnabled',
    ENABLE_AUDIT_LOGGING = 'enableAuditLogging',
    WORKER_NODE_INSTANCE_TYPE = 'workerNodeInstanceType'
}

export const NodeSettingStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: NodeSettingField.CLUSTER_NAME, label: 'CLUSTER NAME', validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
        { name: NodeSettingField.ENABLE_AUDIT_LOGGING, isBoolean: true, label: 'ENABLE AUDIT LOGGING' },
        { name: NodeSettingField.INSTANCE_TYPE_DEV, label: 'INSTANCE TYPE' },
        { name: NodeSettingField.INSTANCE_TYPE_PROD, label: 'INSTANCE TYPE' },
        { name: NodeSettingField.MACHINE_HEALTH_CHECKS_ENABLED, isBoolean: true, label: 'MACHINE HEALTH CHECKS' },
    ]
};
export const NodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        ...NodeSettingStandaloneStepMapping.fieldMappings,
        { name: NodeSettingField.WORKER_NODE_INSTANCE_TYPE, required: true, label: 'WORKER NODE INSTANCE TYPE' },
    ]
}
