import {SimpleValidator} from "./constants/validation.constants";

export interface FieldMapping {
    name: string,
    validators?: SimpleValidator[],
    defaultValue?: any,
    isBoolean?: boolean,
    required?: boolean,
    featureFlag?: string,
}
export interface StepMapping {
    name: string,
    form: string,
    fieldMappings: FieldMapping[],
}
