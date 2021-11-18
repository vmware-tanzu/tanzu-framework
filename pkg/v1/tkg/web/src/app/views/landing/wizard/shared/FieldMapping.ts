export interface FieldMapping {
    name: string,
    validators?: any[],
    defaultValue?: any,
    isBoolean?: boolean,
    required?: boolean,
}
 export interface StepMapping {
     name: string,
     form: string,
     fieldMappings: FieldMapping[],
 }
