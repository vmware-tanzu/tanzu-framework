import {AbstractControl, FormArray, FormControl, FormGroup} from '@angular/forms';
import AppServices from 'src/app/shared/service/appServices';
import {ControlType, FieldMapping} from '../field-mapping/FieldMapping';

export class FormUtils {
    static addControl(formGroup: FormGroup, name: string, control: AbstractControl) {
        // Extends default FormGroup.addControl method to add form controls and
        // set the emitEvent property to false by default. This may be used to bypass
        // Clarity 5's status change logic, which triggers form validation to occur
        // immediately after form controls are created. Setting emitEvent to false
        // avoids the Clarity 'triggerAllFormControlValidation' method from being executed.
        formGroup.addControl(name, control, {emitEvent: false});
    }

    static addDynamicControl(formGroup: FormGroup, initialValue: any, fieldMapping: FieldMapping): void {
        const dynamicControlHandler =
            (dynamicMapping: Record<ControlType, () => void>, defaultCase = ControlType.FormControl) =>
                (controlType: ControlType) =>
                    (dynamicMapping[controlType] || dynamicMapping[defaultCase])();

        const formArrayHandler: () => void = () => {
            const formData: any[] = (initialValue && initialValue !== '' ? initialValue : null) ??
                fieldMapping.defaultValue ?? [
                    fieldMapping.children.reduce((obj, item) => ((obj[item.name] = item.defaultValue), obj), {}),
                ];

            const formArray: any[] = formData.map((obj) => {
                const group: any = {};
                for (const [key, value] of Object.entries(obj)) {
                    const childFieldMapping: FieldMapping = fieldMapping.children.find((child) => child.name === key);
                    const childValidators = AppServices.fieldMapUtilities.getValidatorArray(childFieldMapping);
                    group[key] = new FormControl(value, childValidators);
                }
                return new FormGroup(group);
            });

            formGroup.addControl(fieldMapping.name, new FormArray(formArray), {emitEvent: false});
        };

        const formControlHandler: () => void = () => {
            const validators = AppServices.fieldMapUtilities.getValidatorArray(fieldMapping);
            formGroup.addControl(fieldMapping.name, new FormControl(initialValue, validators), {emitEvent: false});
            setTimeout(() => {
                formGroup.controls[fieldMapping.name].setValue(initialValue);
            });
        };

        const dynamicMappings: Record<ControlType, () => void> = {
            [ControlType.FormArray]: formArrayHandler,
            [ControlType.FormControl]: formControlHandler,
            // TODO: Add FormGroup handler
            [ControlType.FormGroup]: () => {
            },
        };

        dynamicControlHandler(dynamicMappings)(fieldMapping.controlType);
    }
}
