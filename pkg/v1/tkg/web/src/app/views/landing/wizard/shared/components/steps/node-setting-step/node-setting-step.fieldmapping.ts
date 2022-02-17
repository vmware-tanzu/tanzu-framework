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

export const NodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        { name: NodeSettingField.CLUSTER_NAME, label: 'CLUSTER NAME', validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
        { name: NodeSettingField.INSTANCE_TYPE_DEV, label: 'INSTANCE TYPE', primaryTrigger: true },
        { name: NodeSettingField.INSTANCE_TYPE_PROD, label: 'INSTANCE TYPE', primaryTrigger: true },
        { name: NodeSettingField.WORKER_NODE_INSTANCE_TYPE, required: true, label: 'WORKER NODE INSTANCE TYPE' },
        { name: NodeSettingField.ENABLE_AUDIT_LOGGING, isBoolean: true, label: 'ACTIVATE AUDIT LOGGING' },
        { name: NodeSettingField.MACHINE_HEALTH_CHECKS_ENABLED, isBoolean: true, label: 'MACHINE HEALTH CHECKS' },
    ]
}
