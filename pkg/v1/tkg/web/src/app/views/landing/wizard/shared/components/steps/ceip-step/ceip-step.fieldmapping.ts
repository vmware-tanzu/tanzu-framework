import { StepMapping } from '../../../field-mapping/FieldMapping';

export enum CeipField {
    OPTIN = 'ceipOptIn',
}

export const CeipStepMapping: StepMapping = {
    fieldMappings: [
        { name: CeipField.OPTIN, isBoolean: true, defaultValue: true }
    ]
}
