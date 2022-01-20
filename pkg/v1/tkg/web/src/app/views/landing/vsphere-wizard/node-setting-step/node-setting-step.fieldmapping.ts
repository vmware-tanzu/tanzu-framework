import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';

export const VsphereNodeSettingStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: VsphereField.NODESETTING_CLUSTER_NAME, validators: [SimpleValidator.IS_VALID_CLUSTER_NAME], label: 'CLUSTER NAME' },
        { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, required: true, validators: [SimpleValidator.IS_VALID_FQDN_OR_IP],
            label: 'CONTROL PLANE ENDPOINT' },
        { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER, required: true, label: 'CONTROL PLANE ENDPOINT PROVIDER' },
        { name: VsphereField.NODESETTING_CONTROL_PLANE_SETTING, required: true, primaryTrigger: true },
        { name: VsphereField.NODESETTING_INSTANCE_TYPE_DEV, required: true, label: 'DEVELOPMENT INSTANCE TYPE' },
        { name: VsphereField.NODESETTING_INSTANCE_TYPE_PROD, required: true, label: 'PRODUCTION INSTANCE TYPE' },
        { name: VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED, isBoolean: true, label: 'MACHINE HEALTH CHECKS' },
        { name: VsphereField.NODESETTING_ENABLE_AUDIT_LOGGING, isBoolean: true, label: 'ENABLE AUDIT LOGGING' },
    ]
}
export const VsphereNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        ...VsphereNodeSettingStandaloneStepMapping.fieldMappings,
        { name: VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, required: true, label: 'WORKER NODE INSTANCE TYPE',
            requiresBackendData: true }
    ]
}
// About VsphereNodeSettingStandaloneStep:
// The first thing the user is expected to do is select DEV or PROD as NODESETTING_CONTROL_PLANE_SETTING, which
// will cascade into activating/deactivating other fields based on its value. Therefore, NODESETTING_CONTROL_PLANE_SETTING is set as a
// primaryTrigger.
