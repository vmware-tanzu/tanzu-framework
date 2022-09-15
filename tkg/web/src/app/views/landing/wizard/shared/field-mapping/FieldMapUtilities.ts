// Angular imports
import { Injectable } from '@angular/core';
import { ValidatorFn, Validators } from '@angular/forms';
// App imports
import AppServices from '../../../../../shared/service/appServices';
import { Cloneable } from '../utils/cloneable';
import { FieldMapping, StepMapping } from './FieldMapping';
import { managementClusterPlugin } from '../constants/wizard.constants';
import { ValidationService } from '../validation/validation.service';

@Injectable()
export class FieldMapUtilities {
    constructor(private validationService: ValidationService) {}

    cloneStepMapping(stepMapping: StepMapping): StepMapping {
        return Cloneable.deepCopy<StepMapping>(stepMapping);
    }

    // inserts the given field mappings into stepMapping.fieldMappings after the given field; returns false if field not found
    // However, if field is empty (null, undefined or blank) mappings are inserted at beginning
    insertFieldMappingsAfter(stepMapping: StepMapping, field: string, mappings: FieldMapping[]): boolean {
        let index = -1;

        if (field && field.length) {
            index = stepMapping.fieldMappings.findIndex(fieldMapping => fieldMapping.name === field);
            if (index < 0) {
                console.error('insertFieldMappingsAfter() is unable to find field ' + field + ' in provided stepMapping object');
                return false;
            }
        }
        // Note: we increment index before inserting because we're inserting AFTER
        mappings.forEach(newFieldMapping => stepMapping.fieldMappings.splice(++index, 0, newFieldMapping));
    }

    getActiveFieldMappings(stepMapping: StepMapping): FieldMapping[] {
        return stepMapping.fieldMappings.filter(fieldMapping => !fieldMapping.deactivated);
    }

    // getLabeledFieldsWithStoredData() is called to determine what fields to display on confirmation page
    getLabeledFieldsWithStoredData(wizard, step: string, stepMapping: StepMapping): string[] {
        const result = this.getActiveFieldMappings(stepMapping)
            .filter(fieldMapping => fieldMapping.label)
            .filter(fieldMapping => AppServices.userDataService.hasStoredData({wizard, step, field: fieldMapping.name}))
            .reduce<string[]>((accumulator, fieldMapping) => {
                accumulator.push(fieldMapping.name); return accumulator;
            }, []);
        return result;
    }

    // returns a Map of fieldName => labels; the map is offered to HTML pages so they can use the labels, and for confirmation page
    getFieldLabelMap(stepMapping: StepMapping): Map<string, string> {
        return this.getActiveFieldMappings(stepMapping)
            .filter(fieldMapping => fieldMapping.label)
            .reduce<Map<string, string>>((accumulator, fieldMapping) => {
                accumulator[fieldMapping.name] = fieldMapping.label;
                return accumulator;
            }, new Map<string, string>());
    }

    getFieldMapping(field: string, stepMapping: StepMapping): FieldMapping {
        if (!stepMapping.fieldMappings) {
            console.error('trying to getFieldMapping for field ' + field + ', but stepMapping has no fieldMappings');
            return null;
        }
        return stepMapping.fieldMappings.find(fieldMapping => fieldMapping.name === field);
    }

    getFieldMappingsToRestore(stepMapping: StepMapping): FieldMapping[] {
        return this.getActiveFieldMappings(stepMapping).filter(fieldMapping => this.shouldRestoreWithStoredValue(fieldMapping));
    }

    getPrimaryTriggerMappingsToRestore(stepMapping: StepMapping): FieldMapping[] {
        return this.getActiveFieldMappings(stepMapping).filter(fieldMapping => fieldMapping.primaryTrigger);
    }

    // The control's initial value should be: the savedValue if there is one (and the mapping said to use it), or
    // the default value (if the mapping provided one), or the blank value (based on whether the field is boolean).
    // Note that if the field is backed by an object (rather than a simple string), a retriever should be passed that can retrieve
    // the object based on the saved value (which is presumably a unique identifier of the object)
    getInitialValue(wizard, step: string, fieldMapping: FieldMapping) {
        const blankValue = fieldMapping.isBoolean ? false : '';
        if (this.shouldInitializeWithStoredValue(fieldMapping)) {
            const storedValue = AppServices.userDataService.retrieveStoredValue(wizard, step, fieldMapping);
            // note: a storedValue of FALSE should be returned, so be careful with the following "if" test
            if (storedValue !== null && storedValue !== undefined) {
                return storedValue;
            }
        }
        return fieldMapping.defaultValue ? fieldMapping.defaultValue : blankValue;
    }

    shouldAutoSave(fieldMapping: FieldMapping) {
        return !fieldMapping.deactivated && !fieldMapping.doNotAutoSave && !fieldMapping.displayOnly && !fieldMapping.neverStore &&
            (!fieldMapping.featureFlag || this.isFeatureEnabled(fieldMapping.featureFlag))
    }

    // The cases where we should NOT initialize using the stored value:
    // (1) The mapping indicates this is the first trigger field (so it needs to be set after listeners are established), or
    // (2) The mapping explicitly says not to auto-restore (usually because the value is set in an onChange handler for another field), or
    // (3) The mapping indicates the field value is never stored, or
    // (4) The mapping indicates the field depends on back-end data which may not have arrived at initialization time. The listener of
    // the backend-data-arrived event is then responsible for setting the field's value
    // See FieldMapping.ts for a lengthier explanation of the meaning (and expected usage) of these attributes
    private shouldInitializeWithStoredValue(fieldMapping: FieldMapping): boolean {
        return !fieldMapping.deactivated && !fieldMapping.primaryTrigger && !fieldMapping.doNotAutoRestore &&
            !fieldMapping.requiresBackendData && !fieldMapping.neverStore;
    }

    // This is used primarily in the case where a file was imported and a step wants to use the data to populate the relevant fields
    // of the step's form. The form has already been initialized, so all the fields should be set from stored data EXCEPT:
    // (1) fields that are marked doNotAutoRestore, since these fields are generally populated by onChange event handlers,
    // (2) fields that never store a value, and
    // (3) the primaryTrigger field. This is because all the OTHER fields should be set BEFORE the primaryTrigger field is set (in case
    // the primaryTrigger field uses any of the other fields' values,
    // (4) the field is not used due to a feature flag
    private shouldRestoreWithStoredValue(fieldMapping: FieldMapping): boolean {
        return !fieldMapping.deactivated && !fieldMapping.doNotAutoRestore && !fieldMapping.neverStore && !fieldMapping.primaryTrigger &&
            !fieldMapping.displayOnly && this.passesFeatureFlagFilter(fieldMapping);
    }

    passesFeatureFlagFilter(fieldMapping: FieldMapping): boolean {
        return !fieldMapping.featureFlag || this.isFeatureEnabled(fieldMapping.featureFlag);
    }

    getValidatorArray(fieldMapping: FieldMapping): ValidatorFn[] {
        const validators = fieldMapping.required ? [Validators.required] : [];
        if (fieldMapping.validators && fieldMapping.validators.length > 0) {
            fieldMapping.validators.forEach( (simpleValidator, index) => {
                const validator = this.validationService.getSimpleValidator(simpleValidator);
                if (!validator) {
                    console.warn('error building field ' + fieldMapping.name + ': unable to find validator '
                        + simpleValidator + ' (#' + index + ') in fieldMapping ' + JSON.stringify(fieldMapping));
                } else {
                    validators.push(validator);
                }
            });
        }
        return validators;
    }

    validateFieldMapping(formName: string, fieldMapping: FieldMapping): boolean {
        if (fieldMapping.deactivated) {
            return true;
        }
        if (fieldMapping.isBoolean && fieldMapping.required) {
            return this.consoleInvalidFieldMapping(formName, fieldMapping.name, 'field cannot be required AND boolean');
        }
        if (fieldMapping.backingObject && !fieldMapping.retriever) {
            return this.consoleInvalidFieldMapping(formName, fieldMapping.name, 'backingObject requires retriever');
        }
        if (fieldMapping.hasNoDomControl && !fieldMapping.retriever) {
            return this.consoleInvalidFieldMapping(formName, fieldMapping.name, 'hasNoDomControl requires retriever');
        }
        if (fieldMapping.hasNoDomControl && !fieldMapping.restorer) {
            return this.consoleInvalidFieldMapping(formName, fieldMapping.name, 'hasNoDomControl requires restorer');
        }
        return true;
    }

    private consoleInvalidFieldMapping(formName, fieldName, message: string): boolean {
        console.error('invalid field mapping for ' + formName + '.' + fieldName + ': ' + message);
        return false;
    }

    private isFeatureEnabled(featureFlag: string): boolean {
        return AppServices.appDataService.isPluginFeatureActivated(managementClusterPlugin, featureFlag);
    }
}
