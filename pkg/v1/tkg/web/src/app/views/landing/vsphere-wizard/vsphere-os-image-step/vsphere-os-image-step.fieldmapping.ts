import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const VsphereOsImageStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'osImage', required: true, label: 'OS IMAGE', requiresBackendData: true,
            backingObject: { displayField: 'name', valueField: 'moid' } }
    ]
}
// About VsphereOsImageStep:
// The osImage is always selected from a backing array. Therefore, we don't want the field set to the stored value until the backend
// has populated the osImage listbox. We therefore set requiresBackendData to TRUE, and rely on the event handler of the osImage data
// arriving to set the field value from stored data.
