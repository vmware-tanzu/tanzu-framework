import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureField } from '../azure-wizard.constants';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const AzureNodeSettingStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.NODESETTING_CONTROL_PLANE_SETTING, required: true, primaryTrigger: true },
        { name: AzureField.NODESETTING_INSTANCE_TYPE_DEV, required: true },
        { name: AzureField.NODESETTING_INSTANCE_TYPE_PROD, required: true },
        { name: AzureField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED, isBoolean: true },
        { name: AzureField.NODESETTING_ENABLE_AUDIT_LOGGING, isBoolean: true, label: 'ENABLE AUDIT LOGGING' },
        { name: AzureField.NODESETTING_MANAMGEMENT_CLUSTER_NAME, validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
    ]
};
export const AzureNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        ...AzureNodeSettingStandaloneStepMapping.fieldMappings,
        { name: AzureField.NODESETTING_WORKERTYPE, required: true },
    ]
}
// About AzureNodeSettingStandaloneStep:
// The first thing the user is expected to do is select DEV or PROD as NODESETTING_CONTROL_PLANE_SETTING, which
// will cascade into activating/deactivating other fields based on its value. Therefore, NODESETTING_CONTROL_PLANE_SETTING is set as a
// primaryTrigger and most of the other fields on the form are set doNotAutoRestore because their values depend either on the DEV/PROD
// setting and/or on backend data events.
