import { FieldMapping, StepMapping } from './FieldMapping';
import { FormControl, FormGroup, ValidatorFn, Validators } from '@angular/forms';
import { ValidationService } from '../validation/validation.service';
import { FormMetaDataStore } from '../FormMetaDataStore';
import { FormUtils } from '../utils/form-utils';
import { Injectable } from '@angular/core';
import Broker from '../../../../../shared/service/broker';
import { managementClusterPlugin } from '../constants/wizard.constants';
import { UserDataIdentifier } from '../../../../../shared/service/user-data.service';

@Injectable()
export class FieldMapUtilities {
    constructor(private validationService: ValidationService) {
    }

    getLabeledFieldsWithStoredData(wizard, step: string, stepMapping: StepMapping): string[] {
        const result = stepMapping.fieldMappings
            .filter(fieldMapping => fieldMapping.label)
            .filter(fieldMapping => Broker.userDataService.hasStoredData({wizard, step, field: fieldMapping.name}))
            .reduce<string[]>((accumulator, fieldMapping) => {
                accumulator.push(fieldMapping.name); return accumulator;
            }, []);
        return result;
    }

    // returns a Map of fieldName => labels
    getFieldLabelMap(stepMapping: StepMapping): Map<string, string> {
        return stepMapping.fieldMappings.filter(fieldMapping => fieldMapping.label)
            .reduce<Map<string, string>>((accumulator, fieldMapping) => {
                accumulator[fieldMapping.name] = fieldMapping.label;
                return accumulator;
            }, new Map<string, string>());
    }

    getFieldMapping(field: string, stepMapping: StepMapping): FieldMapping {
        return stepMapping.fieldMappings.find(mapping => mapping.name === field);
    }

    buildForm(formGroup: FormGroup, formName: string, stepMapping: StepMapping) {
        stepMapping.fieldMappings.forEach(fieldMapping => {
            if (this.passesFeatureFlagFilter(fieldMapping)) {
                this.buildFormField(formGroup, formName, fieldMapping);
            }
        });
    }

    restoreForm(wizard, step: string, formGroup: FormGroup, stepMapping: StepMapping) {
        this.getFieldsToRestore(stepMapping).forEach(field => {
            this.restoreField({ wizard, step, field }, formGroup, { emitEvent: false });
        });
        // Note: only AFTER all the "regular" fields are restored do we set the values on the primary trigger fields, because the
        // handler for the trigger field change may expect the values of the other fields to have been set
        this.getPrimaryTriggers(stepMapping).forEach(field => {
            this.restoreField({ wizard, step, field }, formGroup);
        })
    }

    restoreField(identifier: UserDataIdentifier, formGroup: FormGroup, options?: {
        onlySelf?: boolean,
        emitEvent?: boolean
    }) {
        const storedEntry = Broker.userDataService.retrieve(identifier);
        const control = formGroup.get(identifier.field);
        if (control && storedEntry) {
            control.setValue(storedEntry.value, options);
        }
        if (!control) {
            console.warn('Trying to restore field ' + identifier.field + ', but cannot locate field in formGroup');
        }
    }

    private getFieldsToRestore(stepMapping: StepMapping): string[] {
        return stepMapping.fieldMappings
            .filter(fieldMapping => this.shouldRestoreWithStoredValue(fieldMapping))
            .reduce<string[]>((accumulator, fieldMapping) => {
                accumulator.push(fieldMapping.name);
                return accumulator;
            }, []);
    }

    private getPrimaryTriggers(stepMapping: StepMapping): string[] {
        return stepMapping.fieldMappings
            .filter(fieldMapping => fieldMapping.primaryTrigger)
            .reduce<string[]>((accumulator, fieldMapping) => {
                accumulator.push(fieldMapping.name);
                return accumulator;
            }, []);
    }
    // The control's initial value should be: the savedValue if there is one (and the mapping said to use it), or
    // the default value (if the mapping provided one), or the blank value (based on whether the field is boolean).
    private getInitialValue(formName: string, fieldMapping: FieldMapping) {
        const blankValue = fieldMapping.isBoolean ? false : '';
        let savedValue;
        if (this.shouldInitializeWithStoredValue(fieldMapping)) {
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

    // There are four cases where we should NOT initialize using the stored value:
    // (1) The mapping indicates this is the first trigger field (so it needs to be set after listeners are established), or
    // (2) The mapping explicitly says not to auto-restore (usually because the value is set in an onChange handler for another field), or
    // (3) The mapping indicates the field value is never stored, or
    // (4) The mapping indicates the field depends on back-end data which may not have arrived at initialization time. The listener of
    // the backend-data-arrived event is then responsible for setting the field's value
    // See FieldMapping.ts for a lengthier explanation of the meaning (and expected usage) of these attributes
    private shouldInitializeWithStoredValue(fieldMapping: FieldMapping): boolean {
        return !fieldMapping.primaryTrigger && !fieldMapping.doNotAutoRestore && !fieldMapping.requiresBackendData &&
            !fieldMapping.neverStore;
    }

    // This is used primarily in the case where a file was imported and a step wants to use the data to populate the relevant fields
    // of the step's form. The form has already been initialized, so all the fields should be set from stored data EXCEPT:
    // (1) fields that are marked doNotAutoRestore, since these fields are generally populated by onChange event handlers,
    // (2) fields that never store a value, and
    // (3) the primaryTrigger field. This is because all the OTHER fields should be set BEFORE the primaryTrigger field is set (in case
    // the primaryTrigger field uses any of the other fields' values,
    // (4) the field is not used due to a feature flag
    private shouldRestoreWithStoredValue(fieldMapping: FieldMapping): boolean {
        return !fieldMapping.doNotAutoRestore && !fieldMapping.neverStore && !fieldMapping.primaryTrigger &&
            this.passesFeatureFlagFilter(fieldMapping);
    }

    private passesFeatureFlagFilter(fieldMapping: FieldMapping): boolean {
        return !fieldMapping.featureFlag || this.isFeatureEnabled(fieldMapping.featureFlag);
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

    private buildFormField(formGroup: FormGroup, formName: string, fieldMapping: FieldMapping) {
        this.validateFieldMapping(formName, fieldMapping);
        const initialValue = this.getInitialValue(formName, fieldMapping);
        const validators = this.getValidatorArray(formName, fieldMapping);
        FormUtils.addControl(
            formGroup,
            fieldMapping.name,
            new FormControl(initialValue, validators)
        );
    }

    private isFeatureEnabled(featureFlag: string): boolean {
        return Broker.appDataService.isPluginFeatureActivated(managementClusterPlugin, featureFlag);
    }
}
