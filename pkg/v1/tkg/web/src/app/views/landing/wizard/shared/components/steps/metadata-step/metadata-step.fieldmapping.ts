import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export enum MetadataField {
    CLUSTER_LABELS = 'clusterLabels',
    CLUSTER_DESCRIPTION = 'clusterDescription',
    CLUSTER_LOCATION = 'clusterLocation'
}

export const MetadataStepMapping: StepMapping = {
    fieldMappings: [
        { name: MetadataField.CLUSTER_LABELS },
        { name: MetadataField.CLUSTER_DESCRIPTION, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
        { name: MetadataField.CLUSTER_LOCATION, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
    ]
}
