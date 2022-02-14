import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';

export const VsphereNodeSettingFieldMappings = [
    { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER, required: true, label: 'CONTROL PLANE ENDPOINT PROVIDER' },
    { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, required: true, validators: [SimpleValidator.IS_VALID_FQDN_OR_IP],
        label: 'CONTROL PLANE ENDPOINT' },
];
// About VsphereNodeSettingStandaloneStep:
// Extends common NodeSettingStep
