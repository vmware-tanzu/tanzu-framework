import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureField } from '../azure-wizard.constants';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const AzureNodeSettingStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.NODESETTING_CONTROL_PLANE_SETTING, required: true },
        { name: AzureField.NODESETTING_INSTANCE_TYPE_DEV, required: true },
        { name: AzureField.NODESETTING_INSTANCE_TYPE_PROD, required: true },
        { name: AzureField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED },
        { name: AzureField.NODESETTING_MANAMGEMENT_CLUSTER_NAME, validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
    ]
};
export const AzureNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        ...AzureNodeSettingStandaloneStepMapping.fieldMappings,
        { name: AzureField.NODESETTING_WORKERTYPE, required: true },
    ]
}


