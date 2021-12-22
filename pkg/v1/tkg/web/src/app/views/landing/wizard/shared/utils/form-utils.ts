import { AbstractControl, FormGroup } from "@angular/forms";

export class FormUtils {
    static addControl(formGroup: FormGroup, name: string, control: AbstractControl) {
        // Extends default FormGroup.addControl method to add form controls and
        // set the emitEvent property to false by default. This may be used to bypass
        // Clarity 5's status change logic, which triggers form validation to occur
        // immediately after form controls are created. Setting emitEvent to false
        // avoids the Clarity 'triggerAllFormControlValidation' method from being executed.
        formGroup.addControl(name, control, { emitEvent: false });
    }
}
