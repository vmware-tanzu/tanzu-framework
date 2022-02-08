import { StepMapping } from '../../../field-mapping/FieldMapping';

export enum OsImageField {
    IMAGE = 'osImage',
}

export const OsImageStepMapping: StepMapping = {
    fieldMappings: [
        { name: OsImageField.IMAGE, required: true }
    ]
}
