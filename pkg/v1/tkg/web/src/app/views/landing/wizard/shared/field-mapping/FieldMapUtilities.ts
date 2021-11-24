import { FieldMapping, StepMapping } from './FieldMapping';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { ValidationService } from '../validation/validation.service';

export class FieldMapUtilities {
    constructor(private validationService: ValidationService) {
    }

    buildForm(stepMapping: StepMapping, formGroup: FormGroup) {
        stepMapping.fieldMappings.forEach(fieldMapping => {
            let validators = fieldMapping.required ? [Validators.required] : [];
            if (fieldMapping.validators && fieldMapping.validators.length > 0) {
                fieldMapping.validators.forEach( (simpleValidator, index) => {
                    const validator = this.validationService.getSimpleValidator(simpleValidator);
                    if (!validator) {
                        console.warn('unable to find validator #' + index + ' in fieldMapping ' + JSON.stringify(fieldMapping));
                    } else {
                        validators.push(validator);
                    }
                })
            }
            const blankValue = fieldMapping.isBoolean ? false : '';
            const initialValue = fieldMapping.defaultValue ? fieldMapping.defaultValue : blankValue;
            // NOTE: when building the form, we use either a blank value or the given the default value, but NOT
            // the saved value (in local storage). That way when (later) initializing with saved values,
            // there will be an onChange event, which will trigger the right event handler to react to the saved value.
            formGroup.addControl(
                fieldMapping.name,
                new FormControl(initialValue, validators)
            );
        });

    }
}
