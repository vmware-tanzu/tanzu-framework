import { BackingObjectMap, FieldMapping, StepMapping } from '../../views/landing/wizard/shared/field-mapping/FieldMapping';
import { AbstractControl, FormControl, FormGroup } from '@angular/forms';
import AppServices from './appServices';
import { UserDataIdentifier, UserDataService } from './user-data.service';
import { FormUtils } from '../../views/landing/wizard/shared/utils/form-utils';

export class UserDataFormService {
    storeFromMapping(wizard, step: string, stepMapping: StepMapping, formGroup: FormGroup) {
        stepMapping.fieldMappings.forEach( fieldMapping => {
            if (AppServices.fieldMapUtilities.shouldAutoSave(fieldMapping)) {
                this.storeFromFieldMapping(wizard, step, fieldMapping, formGroup);
            }
        });
        AppServices.userDataService.updateWizardTimestamp(wizard);
    }

    private storeFromFieldMapping(wizard, step: string, fieldMapping: FieldMapping, formGroup: FormGroup) {
        const identifier: UserDataIdentifier = { wizard, step, field: fieldMapping.name };
        if (fieldMapping.hasNoDomControl) {
            this.storeFieldWithNoDomControl(wizard, step, fieldMapping);
        } else if (fieldMapping.isBoolean) {
            this.storeBooleanField(identifier, formGroup);
        } else if (fieldMapping.mask) {
            this.storeMaskField(identifier, formGroup);
        } else if (fieldMapping.isMap) {
            this.storeMapField(identifier, formGroup);
        } else if (fieldMapping.backingObject) {
            this.storeBackingObjectField(identifier, formGroup, fieldMapping.backingObject)
        }  else {
            this.storeInputField(identifier, formGroup);
        }
    }

    private storeFieldWithNoDomControl(wizard, step: string, fieldMapping: FieldMapping) {
        if (!fieldMapping.retriever) {
            console.error('field ' + fieldMapping.name + ' has no DOM control, but no retriever provided');
            return;
        }
        const value = fieldMapping.retriever(null);
        const identifier = { wizard, step, field: fieldMapping.name };
        if (fieldMapping.isBoolean) {
            AppServices.userDataService.storeBoolean(identifier, value);
        } else if (fieldMapping.isMap) {
            AppServices.userDataService.storeMap(identifier, value);
        } else {
            AppServices.userDataService.store(identifier, {display: value, value});
        }
    }

    // convenience methods
    storeInputField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        AppServices.userDataService.store(identifier, { display: control.value, value: control.value });
        return true;
    }

    storeBooleanField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        AppServices.userDataService.storeBoolean(identifier, control.value);
        return true;
    }

    buildForm(formGroup: FormGroup, wizard, step: string, stepMapping: StepMapping) {
        stepMapping.fieldMappings.forEach(fieldMapping => {
            if (this.shouldBuildField(fieldMapping)) {
                this.buildFormField(formGroup, wizard, step, fieldMapping);
            } else if (fieldMapping.hasNoDomControl) {
                if (fieldMapping.restorer) {
                    const value = AppServices.userDataService.retrieveStoredValue(wizard, step, fieldMapping);
                    fieldMapping.restorer(value);
                } else {
                    console.log('field ' + fieldMapping.name + ' has no DOM control, but no restorer was provided');
                }
            }
        });
    }

    restoreForm(wizard, step: string, formGroup: FormGroup, stepMapping: StepMapping) {
        AppServices.fieldMapUtilities.getFieldMappingsToRestore(stepMapping).forEach(fieldMapping => {
            const identifier = { wizard, step, field: fieldMapping.name };
            this.restoreField(identifier, fieldMapping, formGroup);

            // Re-store the masked field value, so that if there WAS a value for this masked field in local storage,
            // it will be erased
            if (fieldMapping.mask) {
                this.storeMaskField(identifier, formGroup);
            }
        });
        // Note: we set the values on the primary trigger fields AFTER all the "regular" fields are restored because the
        // handler for the trigger field change may make use the values of the other fields
        AppServices.fieldMapUtilities.getPrimaryTriggerMappingsToRestore(stepMapping).forEach(fieldMapping => {
            const identifier = { wizard, step, field: fieldMapping.name };
            this.restoreField(identifier, fieldMapping, formGroup);
        })
    }

    private buildFormField(formGroup: FormGroup, wizard, step: string, fieldMapping: FieldMapping) {
        AppServices.fieldMapUtilities.validateFieldMapping(step, fieldMapping);
        const initialValue = AppServices.fieldMapUtilities.getInitialValue(wizard, step, fieldMapping);
        const validators = AppServices.fieldMapUtilities.getValidatorArray(fieldMapping);
        FormUtils.addControl(
            formGroup,
            fieldMapping.name,
            new FormControl(initialValue, validators)
        );
        // TODO: figure out why we cannot seem to set the initialValue using the above code: new FormControl(initialValue, validators),
        // but putting it into a setTimeout closure seems to "fix" the problem
        setTimeout(() => {
            formGroup.controls[fieldMapping.name].setValue(initialValue);
        });
    }

    private shouldBuildField(fieldMapping: FieldMapping) {
        return !fieldMapping.displayOnly && !fieldMapping.hasNoDomControl &&
            AppServices.fieldMapUtilities.passesFeatureFlagFilter(fieldMapping);
    }

    restoreField(identifier: UserDataIdentifier, fieldMapping: FieldMapping, formGroup: FormGroup,
                 options?: { onlySelf?: boolean; emitEvent?: boolean }) {
        const storedValue = AppServices.userDataService.retrieveStoredValue(identifier.wizard, identifier.step, fieldMapping);
        if (storedValue === undefined || storedValue === null || fieldMapping.displayOnly) {
            return;
        }
        if (fieldMapping.restorer) {
            fieldMapping.restorer(storedValue);
            return;
        }

        const control = formGroup.get(identifier.field);
        if (!control) {
            console.error('restoreField(): no DOM control: "' + identifier.wizard + '.' + identifier.step + '.' + identifier.field + '"');
            return;
        }
        control.setValue(storedValue, options);
    }

    // storeBackingObject expects to encounter an OBJECT backing the listbox and will use fieldDisplay of that object for the display
    // and fieldValue for the value. If instead the caller has a simple listbox with strings backing it, call saveInputField instead
    private storeBackingObjectField(identifier: UserDataIdentifier, formGroup: FormGroup, backingObjectMap: BackingObjectMap): boolean {
        const selectedObj = this.getFormObject(identifier, formGroup);
        return AppServices.userDataService.storeBackingObject(identifier, selectedObj, backingObjectMap);
    }

    private storeMaskField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        AppServices.userDataService.store(identifier, { display: control.value ? UserDataService.MASK : '', value: '' });
        return true;
    }

    private storeMapField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        AppServices.userDataService.storeMap(identifier, control.value);
        return true;
    }

    private getFormObject(identifier: UserDataIdentifier, formGroup: FormGroup) {
        return formGroup.value[identifier.field];
    }

    private getFormControl(identifier: UserDataIdentifier, formGroup: FormGroup): AbstractControl {
        const control = formGroup.controls[identifier.field];
        if (!control) {
            console.error('UserDataService.saveSimpleFormField was passed a form group that did not have field ' + identifier.field +
                '. identifier=' + JSON.stringify(identifier));
        }
        return control;
    }
}
