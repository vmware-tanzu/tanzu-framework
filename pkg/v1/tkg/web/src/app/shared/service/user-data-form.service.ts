import { BackingObjectMap, FieldMapping, StepMapping } from '../../views/landing/wizard/shared/field-mapping/FieldMapping';
import { AbstractControl, FormControl, FormGroup } from '@angular/forms';
import AppServices from './appServices';
import { UserDataIdentifier, UserDataService } from './user-data.service';
import { FormUtils } from '../../views/landing/wizard/shared/utils/form-utils';

export class UserDataFormService {
    storeFromMapping(wizard, step: string, stepMapping: StepMapping, formGroup: FormGroup,
                     objectRetrievalMap?: Map<string, (key: any) => any>) {
        stepMapping.fieldMappings.forEach( fieldMapping => {
            if (AppServices.fieldMapUtilities.shouldAutoSave(fieldMapping)) {
                const retriever = this.getFieldBackingObjectRetrieverIfNec(fieldMapping, objectRetrievalMap)
                this.storeFromFieldMapping(wizard, step, fieldMapping, formGroup, retriever);
            }
        });
        AppServices.userDataService.updateWizardTimestamp(wizard);
    }

    private storeFromFieldMapping(wizard, step: string, fieldMapping: FieldMapping, formGroup: FormGroup,
                                  retriever?: (key: any) => any) {
        const identifier: UserDataIdentifier = { wizard, step, field: fieldMapping.name };
        if (fieldMapping.hasNoDomControl) {
            this.storeFieldWithNoDomControl(wizard, step, fieldMapping, retriever);
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

    private storeFieldWithNoDomControl(wizard, step: string, fieldMapping: FieldMapping, retriever: (key: any) => any) {
        if (!retriever) {
            return;
        }
        const value = retriever(null);
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

    buildForm(formGroup: FormGroup, wizard, step: string, stepMapping: StepMapping,
              objectRetrievalMap?: Map<string, (string) => any>, customRestorerMap?: Map<string, (any) => void>) {
        stepMapping.fieldMappings.forEach(fieldMapping => {
            if (this.shouldBuildField(fieldMapping)) {
                const retriever = fieldMapping.backingObject && objectRetrievalMap ? objectRetrievalMap[fieldMapping.name] : null;
                this.buildFormField(formGroup, wizard, step, fieldMapping, retriever);
            } else if (fieldMapping.hasNoDomControl) {
                const restorer = this.getFieldCustomRestorer(fieldMapping, customRestorerMap);
                if (restorer) {
                    const value = AppServices.userDataService.retrieveStoredValue(wizard, step, fieldMapping, null);
                    restorer(value);
                }
            }
        });
    }

    restoreForm(wizard, step: string, formGroup: FormGroup, stepMapping: StepMapping,
                objectRetrievalMap?: Map<string, (string) => any>, customRestorerMap?: Map<string, (any) => void> ) {
        AppServices.fieldMapUtilities.getFieldMappingsToRestore(stepMapping).forEach(fieldMapping => {
            const retriever = this.getFieldBackingObjectRetrieverIfNec(fieldMapping, objectRetrievalMap);
            const restorer = this.getFieldCustomRestorerIfNec(fieldMapping, customRestorerMap);
            const identifier = { wizard, step, field: fieldMapping.name };
            this.restoreField(identifier, fieldMapping, formGroup, {}, retriever, restorer);

            // Re-store the masked field value, so that if there WAS a value for this masked field in local storage,
            // it will be erased
            if (fieldMapping.mask) {
                this.storeMaskField(identifier, formGroup);
            }
        });
        // Note: we set the values on the primary trigger fields AFTER all the "regular" fields are restored because the
        // handler for the trigger field change may make use the values of the other fields
        AppServices.fieldMapUtilities.getPrimaryTriggerMappingsToRestore(stepMapping).forEach(fieldMapping => {
            const retriever = this.getFieldBackingObjectRetrieverIfNec(fieldMapping, objectRetrievalMap);
            const restorer = this.getFieldCustomRestorerIfNec(fieldMapping, customRestorerMap);
            const identifier = { wizard, step, field: fieldMapping.name };
            this.restoreField(identifier, fieldMapping, formGroup, {}, retriever, restorer);
        })
    }

    private buildFormField(formGroup: FormGroup, wizard, step: string, fieldMapping: FieldMapping, retriever?: (string) => any) {
        AppServices.fieldMapUtilities.validateFieldMapping(step, fieldMapping);
        const initialValue = AppServices.fieldMapUtilities.getInitialValue(wizard, step, fieldMapping, retriever);
        const validators = AppServices.fieldMapUtilities.getValidatorArray(fieldMapping);
        FormUtils.addControl(
            formGroup,
            fieldMapping.name,
            new FormControl(initialValue, validators)
        );
    }

    private shouldBuildField(fieldMapping: FieldMapping) {
        return !fieldMapping.displayOnly && !fieldMapping.hasNoDomControl &&
            AppServices.fieldMapUtilities.passesFeatureFlagFilter(fieldMapping);
    }

    private getFieldBackingObjectRetrieverIfNec(fieldMapping: FieldMapping,
                                                objectRetrievalMap?: Map<string, (string) => any>): (string) => any {
        let retriever: (string) => any = null;
        if (fieldMapping.backingObject || fieldMapping.hasNoDomControl) {
            retriever = this.getFieldBackingObjectRetriever(fieldMapping, objectRetrievalMap);
        }
        return retriever;
    }

    private getFieldBackingObjectRetriever(fieldMapping: FieldMapping,
                                                objectRetrievalMap?: Map<string, (string) => any>): (string) => any {
        let retriever: (string) => any = null;
        if (!objectRetrievalMap) {
            console.error('getFieldBackingObjectRetriever() encountered field "' + fieldMapping.name +
                '" which is using a backingObject, but had no objectRetrievalMap to get the backing object');
            return null;
        }
        retriever = objectRetrievalMap.get(fieldMapping.name);
        if (!retriever) {
            console.error('getFieldBackingObjectRetriever() encountered field "' + fieldMapping.name +
                '" which is using a backingObject, but had no retriever in the objectRetrievalMap');
        }
        return retriever;
    }

    private getFieldCustomRestorerIfNec(fieldMapping: FieldMapping, customRestorerMap: Map<string, (any) => void>): (any) => void {
        let restorer: (any) => void = null;

        if (fieldMapping.hasNoDomControl) {
            restorer = this.getFieldCustomRestorer(fieldMapping, customRestorerMap);
        }
        return restorer;
    }

    private getFieldCustomRestorer(fieldMapping: FieldMapping, customRestorerMap: Map<string, (any) => void>): (any) => void {
        if (!customRestorerMap) {
            console.error('getFieldCustomRestorer() encountered field "' + fieldMapping.name +
                '" which is using a custom restorer, but had no customRestorerMap to get the restorer');
            return null;
        }
        const restorer = customRestorerMap.get(fieldMapping.name);
        if (!restorer) {
            console.error('getFieldCustomRestorer() encountered field "' + fieldMapping.name +
                '" which is using a custom restorer, but had no restorer in the customRestorerMap');
        }
        return restorer;
    }

    restoreField(identifier: UserDataIdentifier, fieldMapping: FieldMapping, formGroup: FormGroup,
                 options?: { onlySelf?: boolean; emitEvent?: boolean }, retriever?: (string) => any, restorer?: (any) => void) {
        const storedValue = AppServices.userDataService.retrieveStoredValue(identifier.wizard, identifier.step, fieldMapping, retriever);
        if (!storedValue === undefined || storedValue === null || fieldMapping.displayOnly) {
            return;
        }
        if (restorer) {
            restorer(storedValue);
            return;
        }

        const control = formGroup.get(identifier.field);
        if (!control) {
            console.error('restoreField(): no control: "' + identifier.wizard + '.' + identifier.step + '.' + identifier.field + '"');
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
