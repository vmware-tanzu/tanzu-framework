import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';

export const VsphereNodeSettingStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: VsphereField.NODESETTING_CLUSTER_NAME, validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
        { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, required: true, validators: [SimpleValidator.IS_VALID_FQDN_OR_IP] },
        { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER, required: true },
        { name: VsphereField.NODESETTING_CONTROL_PLANE_SETTING, required: true },
        { name: VsphereField.NODESETTING_INSTANCE_TYPE_DEV, required: true },
        { name: VsphereField.NODESETTING_INSTANCE_TYPE_PROD, required: true },
        { name: VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED, required: true },
    ]
}
export const VsphereNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        ...VsphereNodeSettingStandaloneStepMapping.fieldMappings,
        { name: VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, required: true }
    ]
}
