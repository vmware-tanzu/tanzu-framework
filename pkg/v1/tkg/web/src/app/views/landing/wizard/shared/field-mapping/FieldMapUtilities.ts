import { FieldMapping, StepMapping } from './FieldMapping';
import { FormControl, FormGroup, ValidatorFn, Validators } from '@angular/forms';
import { ValidationService } from '../validation/validation.service';
import { FormMetaDataStore } from '../FormMetaDataStore';
import { FormUtils } from '../utils/form-utils';
import { Injectable } from '@angular/core';

@Injectable()
export class FieldMapUtilities {
    constructor(private validationService: ValidationService) {
    }

    static getFieldMapping(name: string, stepMapping: StepMapping): FieldMapping {
        if (stepMapping && name) {
            return stepMapping.fieldMappings.find((daFieldMapping) => { return daFieldMapping.name === name; });
        }
        console.warn('getFieldMapping could not find an entry for field named ' + name);
        return null;
    }

    // The control's initial value should be: the savedValue if there is one (and the mapping said to use it), or
    // the default value (if the mapping provided one), or the blank value (based on whether the field is boolean).
    // NOTE: usually when building the form, we use either a blank value or the given the default value, but NOT
    // the saved value (in local storage). That way when (later) initializing with saved values,
    // there will be an onChange event, which will trigger the right event handler to react to the saved value. However,
    // controls that don't trigger an onChange event can use the "initWithSavedValue=true" to init the control via the mapping.
    private getInitialValue(formName: string, fieldMapping: FieldMapping) {
        const blankValue = fieldMapping.isBoolean ? false : '';
        let savedValue;
        if (fieldMapping.initWithSavedValue) {
            const metadataEntry = FormMetaDataStore.getMetaDataItem(formName, fieldMapping.name);
            if (metadataEntry) {
                savedValue = metadataEntry.key ? metadataEntry.key : metadataEntry.displayValue;
                if (savedValue) {
                    return savedValue;
                }
            }
        }
        return fieldMapping.defaultValue ? fieldMapping.defaultValue : blankValue;
    }

    private getValidatorArray(formName: string, fieldMapping: FieldMapping): ValidatorFn[] {
        const validators = fieldMapping.required ? [Validators.required] : [];
        if (fieldMapping.validators && fieldMapping.validators.length > 0) {
            fieldMapping.validators.forEach( (simpleValidator, index) => {
                const validator = this.validationService.getSimpleValidator(simpleValidator);
                if (!validator) {
                    console.warn('error building field ' + formName + '.' + fieldMapping.name + ': unable to find validator '
                        + simpleValidator + ' (#' + index + ') in fieldMapping ' + JSON.stringify(fieldMapping));
                } else {
                    validators.push(validator);
                }
            });
        }
        return validators;
    }

    private validateFieldMapping(formName: string, fieldMapping: FieldMapping): boolean {
        let result = true;
        if (fieldMapping.isBoolean && fieldMapping.required) {
            result = false;
            console.error('invalid field mapping for ' + formName + '.' + fieldMapping.name + ': field cannot be required AND boolean');
        }
        return result;
    }

    // Note: the form name is only used to retrieve saved values
    buildForm(formGroup: FormGroup, formName: string, stepMapping: StepMapping) {
        stepMapping.fieldMappings.forEach(fieldMapping => {
            this.validateFieldMapping(formName, fieldMapping);
            const initialValue = this.getInitialValue(formName, fieldMapping);
            const validators = this.getValidatorArray(formName, fieldMapping);
            FormUtils.addControl(
                formGroup,
                fieldMapping.name,
                new FormControl(initialValue, validators)
            );
        });
    }
}
