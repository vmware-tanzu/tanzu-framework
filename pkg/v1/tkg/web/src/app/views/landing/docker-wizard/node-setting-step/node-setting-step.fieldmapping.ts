import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const TkgDockerNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'clusterName', validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
    ]
}
export const DockerNodeSettingStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'clusterName', required: true, validators: [SimpleValidator.IS_VALID_CLUSTER_NAME] },
    ]
}
