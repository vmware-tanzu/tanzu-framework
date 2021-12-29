import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export enum MetadataField {
    CLUSTER_LABELS = 'clusterLabels',
    CLUSTER_DESCRIPTION = 'clusterDescription',
    CLUSTER_LOCATION = 'clusterLocation'
}

export const MetadataStepMapping: StepMapping = {
    fieldMappings: [
        { name: MetadataField.CLUSTER_LABELS, doNotAutoSave: true, label: 'LABELS (OPTIONAL)', isMap: true },
        { name: MetadataField.CLUSTER_DESCRIPTION, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION], label: 'DESCRIPTION (OPTIONAL)' },
        { name: MetadataField.CLUSTER_LOCATION, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION], label: 'LOCATION (OPTIONAL)' },
    ]
}
// About MetadataStep:
// The clusterLabels field is a hidden field that holds a map of all the key-value pairs the user enters.
// We do not AutoSave because the display value needs to be "manually" generated (turning the map into a string)
// TODO: use isMap to auto-generate the display, allowing AutoSave to work for clusterLabels
// We never save the newLabel key-value pairs because they should only be used temporarily while the user is creating the values
