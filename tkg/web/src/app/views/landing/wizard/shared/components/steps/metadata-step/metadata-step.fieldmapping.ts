import { ControlType, StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export enum MetadataField {
    CLUSTER_LABELS = 'clusterLabels',
    CLUSTER_DESCRIPTION = 'clusterDescription',
    CLUSTER_LOCATION = 'clusterLocation',
}

export const MetadataStepMapping: StepMapping = {
    fieldMappings: [
        {
            name: MetadataField.CLUSTER_LOCATION,
            validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION],
            label: 'LOCATION (OPTIONAL)'
        },
        {
            name: MetadataField.CLUSTER_DESCRIPTION,
            validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION],
            label: 'DESCRIPTION (OPTIONAL)'
        },
        {
            name: MetadataField.CLUSTER_LABELS,
            label: 'LABELS (OPTIONAL)',
            controlType: ControlType.FormArray,
            displayFunction: labels => labels.filter(label => label.key).map(label => `${label.key} : ${label.value}`).join(', '),
            children: [
                {
                    name: 'key',
                    defaultValue: '',
                    controlType: ControlType.FormControl,
                    validators: [
                        SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION,
                        SimpleValidator.RX_UNIQUE,
                        SimpleValidator.RX_REQUIRED_IF_VALUE
                    ]
                },
                {
                    name: 'value',
                    defaultValue: '',
                    controlType: ControlType.FormControl,
                    validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION, SimpleValidator.RX_REQUIRED_IF_KEY]
                }
            ]
        }
    ]
};
