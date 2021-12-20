import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export const MetadataStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'clusterLabels' },
        { name: 'clusterDescription', validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
        { name: 'clusterLocation', validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
        { name: 'newLabelKey', validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
        { name: 'newLabelValue', validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] }
    ]
}
