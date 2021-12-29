import { StepMapping } from '../../../field-mapping/FieldMapping';

export enum OsImageField {
    IMAGE = 'osImage',
}

export const OsImageStepMapping: StepMapping = {
    fieldMappings: [
        { name: OsImageField.IMAGE, required: true, label: 'OS IMAGE', requiresBackendData: true }
    ]
}
// About OsImageStep:
// The osImage is always selected from a backing array. Therefore, we don't want the field set to the stored value until the backend
// has populated the osImage listbox. We therefore set requiresBackendData to TRUE, and rely on the event handler of the osImage data
// arriving to set the field value from stored data.
