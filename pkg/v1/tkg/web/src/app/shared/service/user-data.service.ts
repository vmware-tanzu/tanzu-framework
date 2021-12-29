// Angular imports
import { AbstractControl, FormGroup } from '@angular/forms';
// App imports
import Broker from './broker';
import { FieldMapping, StepMapping } from '../../views/landing/wizard/shared/field-mapping/FieldMapping';
import { managementClusterPlugin } from '../../views/landing/wizard/shared/constants/wizard.constants';
import { PersistentStore } from '../../views/landing/wizard/shared/PersistentStore';

export interface UserDataEntry {
    display: string,   // what the user should see if this is displayed on a page
    value: string,     // actual value
}
export interface UserDataIdentifier {
    wizard: string,         // name of the wizard that the data is from
    step: string,           // name of the step that the data is from
    field: string,          // name of field that the data is from
}
// UserDataWizard should only be used by the confirmation page; all steps should use convenience methods
export interface UserDataWizard {
    wizard: string,
    steps: Map<string, UserDataStep>,
    displayOrder?: string[],
    titles?: Map<string, string>,
    descriptions?: Map<string, string>,
}
// UserDataStep should only be used by the confirmation page; all steps should use convenience methods
export interface UserDataStep {
    displayOrder?: string[],
    fields: Map<string, UserDataEntry>,
    labels?: Map<string, string>,
}

export class UserDataService {
    static readonly MASK = '********';
    store(identifier: UserDataIdentifier, data: UserDataEntry) {
        const wizardEntry = this.ensureUserDataIdentifier(identifier);
        this.setUserDataEntry(wizardEntry, identifier, data);
        this.storeWizardEntry(wizardEntry);
    }
    storeFromMapping(wizard, step: string, stepMapping: StepMapping, formGroup: FormGroup) {
        stepMapping.fieldMappings.forEach( fieldMapping => {
            if (this.shouldAutoSave(fieldMapping)) {
                this.storeFromFieldMapping(wizard, step, fieldMapping, formGroup);
            }
        });
    }

    storeWizardDisplayOrder(wizard: string, displayOrder: string[]) {
        const wizardEntry = this.ensureWizardEntry(wizard);
        wizardEntry.displayOrder = displayOrder;
        this.storeWizardEntry(wizardEntry);
    }

    storeWizardTitles(wizard: string, titles: Map<string, string>) {
        const wizardEntry = this.ensureWizardEntry(wizard);
        wizardEntry.titles = titles;
        this.storeWizardEntry(wizardEntry);
    }

    storeWizardDescriptions(wizard: string, descriptions: Map<string, string>) {
        const wizardEntry = this.ensureWizardEntry(wizard);
        wizardEntry.descriptions = descriptions;
        this.storeWizardEntry(wizardEntry);
    }

    storeStepDisplayOrder(wizard, step: string, displayOrder: string[]) {
        const wizardEntry = this.ensureWizardEntry(wizard);
        this.ensureStepEntry(wizardEntry, step);
        wizardEntry.steps[step].displayOrder = displayOrder;
        this.storeWizardEntry(wizardEntry);
    }

    storeStepLabels(wizard, step: string, labels: Map<string, string>) {
        const wizardEntry = this.ensureWizardEntry(wizard);
        this.ensureStepEntry(wizardEntry, step);
        wizardEntry.steps[step].labels = labels;
        this.storeWizardEntry(wizardEntry);
    }

    retrieveWizardEntry(wizard: string) {
        return this.ensureWizardEntry(wizard);
    }

    // SHIMON: currently issue where "delete" isn't recognized as a method?!
    delete(identifier: UserDataIdentifier) {
        const wizardEntry = this.getWizardEntry(identifier.wizard);
        if (wizardEntry && wizardEntry.steps[identifier.step]) {
            const userDataStep = wizardEntry.steps[identifier.step] as UserDataStep;
            const mapEntries = userDataStep.fields as Map<string, UserDataEntry>;
            mapEntries.delete(identifier.field);
        }
    }

    clear(identifier: UserDataIdentifier) {
        const wizardEntry = this.getWizardEntry(identifier.wizard);
        if (wizardEntry && wizardEntry.steps[identifier.step]) {
            this.setUserDataEntry(wizardEntry, identifier, null);
        }
    }

    // This internal convenience method is meant to isolate the access to the internal structure
    private setUserDataEntry(wizardEntry: UserDataWizard, identifier: UserDataIdentifier, data: UserDataEntry) {
        wizardEntry.steps[identifier.step].fields[identifier.field] = data;
    }

    private storeFromFieldMapping(wizard, step: string, fieldMapping: FieldMapping, formGroup: FormGroup) {
        const identifier: UserDataIdentifier = { wizard, step, field: fieldMapping.name };
        if (fieldMapping.isBoolean) {
            this.storeBooleanField(identifier, formGroup);
        } else if (fieldMapping.mask) {
            this.storeMaskField(identifier, formGroup);
        } else if (fieldMapping.isMap) {
            this.storeMapField(identifier, formGroup);
        } else {
            this.storeInputField(identifier, formGroup);
        }
    }

    private shouldAutoSave(fieldMapping: FieldMapping) {
        return !fieldMapping.doNotAutoSave && (!fieldMapping.featureFlag || this.isFeatureEnabled(fieldMapping.featureFlag))
    }

    private isFeatureEnabled(featureFlag: string): boolean {
        return Broker.appDataService.isPluginFeatureActivated(managementClusterPlugin, featureFlag);
    }

    retrieve(identifier: UserDataIdentifier): UserDataEntry {
        const wizardEntry: UserDataWizard = this.getWizardEntry(identifier.wizard);
        if (!wizardEntry || !wizardEntry.steps || !wizardEntry.steps[identifier.step] || !wizardEntry.steps[identifier.step].fields) {
            return null;
        }
        return wizardEntry.steps[identifier.step].fields[identifier.field];
    }

    // convenience methods
    storeInputField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        this.store(identifier, { display: control.value, value: control.value });
        return true;
    }
    hasStoredData(identifier: UserDataIdentifier): boolean {
        const userDataEntry = this.retrieve(identifier);
        // NOTE: we want a value of 'false' to return TRUE (that there IS a value)
        return userDataEntry && userDataEntry.value !== null && userDataEntry.value !== undefined && userDataEntry.value !== '';
    }
    hasStoredStepData(wizard, step: string) {
        const wizardEntry = this.retrieveWizardEntry(wizard);
        if (!wizardEntry) {
            return false;
        }
        const stepEntry = wizardEntry.steps[step];
        return stepEntry !== undefined && stepEntry !== null;
    }

    private storeMaskField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        this.store(identifier, { display: control.value ? UserDataService.MASK : '', value: '' });
        return true;
    }

    private storeMapField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        const display = this.mapToString(control.value);
        this.store(identifier, { display, value: control.value });
        return true;
    }

    private mapToString(map: Map<string, string>): string {
        let labelsStr: string = '';
        map.forEach((value: string, key: string) => {
            labelsStr += key + ':' + value + ', '
        });
        return labelsStr.slice(0, -2);  // chop off the last ', '
    }

    // saveListboxField expects to encounter an OBJECT backing the listbox and will use fieldDisplay of that object for the display
    // and fieldValue for the value. If instead the caller has a simple listbox with strings backing it, call saveInputField instead
    storeListboxObjectField(identifier: UserDataIdentifier, formGroup: FormGroup, fieldDisplay, fieldValue: string): boolean {
        const selectedObj = this.getFormObject(identifier, formGroup);
        if (!selectedObj) {
            return false;
        }
        const display = selectedObj[fieldDisplay];
        const value = selectedObj[fieldValue];
        this.store(identifier, { display, value });
        return true;
    }

    storeBooleanField(identifier: UserDataIdentifier, formGroup: FormGroup): boolean {
        const control = this.getFormControl(identifier, formGroup);
        if (!control) {
            return false;
        }
        this.store(identifier, { display: control.value ? 'yes' : 'no', value: control.value });
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

    private getWizardEntry(wizard: string): UserDataWizard {
        return PersistentStore.getItem(this.keyWizard(wizard));
    }

    private ensureWizardEntry(wizard: string): UserDataWizard {
        // get the wizard entry, or create a new one
        let wizardEntry = this.getWizardEntry(wizard);
        if (!wizardEntry) {
            wizardEntry = this.createUserDataWizard(wizard);
        }
        return wizardEntry;
    }
    private ensureStepEntry(wizardEntry, step: string): UserDataWizard {
        // if the step entry isn't there, create it
        if (!wizardEntry.steps[step]) {
            this.createUserStepEntry(wizardEntry, step);
        }
        return wizardEntry;
    }

    private ensureUserDataIdentifier(identifier: UserDataIdentifier): UserDataWizard {
        const wizardEntry = this.ensureWizardEntry(identifier.wizard);
        this.ensureStepEntry(wizardEntry, identifier.step);
        return wizardEntry;
    }

    private createUserDataWizard(wizard: string): UserDataWizard {
        return { wizard, steps: new Map<string, UserDataStep>() };
    }

    private createUserStepEntry(wizardEntry: UserDataWizard, step: string) {
        const newStepEntry: UserDataStep = {
            fields: new Map<string, UserDataEntry>(),
        }
        wizardEntry.steps[step] = newStepEntry;
    }

    private keyWizard(wizard: string): string {
        return wizard + 'Storage';
    }

    private storeWizardEntry(wizardEntry: UserDataWizard) {
        PersistentStore.setItem(this.keyWizard(wizardEntry.wizard), wizardEntry);
    }
}
