import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export enum MetadataField {
    CLUSTER_LABELS = 'clusterLabels',
    CLUSTER_DESCRIPTION = 'clusterDescription',
    CLUSTER_LOCATION = 'clusterLocation'
}

export const MetadataStepMapping: StepMapping = {
    fieldMappings: [
        { name: MetadataField.CLUSTER_LOCATION, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION], label: 'LOCATION (OPTIONAL)' },
        { name: MetadataField.CLUSTER_DESCRIPTION, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION], label: 'DESCRIPTION (OPTIONAL)' },
        { name: MetadataField.CLUSTER_LABELS, hasNoDomControl: true, isMap: true, label: 'LABELS (OPTIONAL)' },
    ]
}
// About MetadataStep:
// The clusterLabels field does not actually exist in the DOM; the values are held in the step component.
// We use hasNoDomControl because the display value needs to be "manually" generated.
// Note that there are DOM fields that hold various pieces of the cluster labels, but we ignore them in favor of a single "field"
// that contains the entire map.
