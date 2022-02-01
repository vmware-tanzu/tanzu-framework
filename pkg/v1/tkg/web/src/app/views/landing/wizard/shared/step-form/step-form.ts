// Angular imports
import { Directive, OnInit } from '@angular/core';
import { AbstractControl, FormGroup, ValidatorFn } from '@angular/forms';
// Third party imports
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';
// App imports
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import AppServices from 'src/app/shared/service/appServices';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { EditionData } from 'src/app/shared/service/branding.service';
import { FormUtility } from '../components/steps/form-utility';
import { IpFamilyEnum } from 'src/app/shared/constants/app.constants';
import { Notification, NotificationTypes } from 'src/app/shared/components/alert-notification/alert-notification.component';
import { StepCompletedPayload, StepDescriptionChangePayload, TanzuEvent, TanzuEventType } from 'src/app/shared/service/Messenger';
import { StepMapping } from '../field-mapping/FieldMapping';
import { StepRegistrantData } from '../wizard-base/wizard-base';
import { UserDataIdentifier } from '../../../../../shared/service/user-data.service';
import { ValidatorEnum } from './../constants/validation.constants';

export interface StepDescriptionTriggers {
    clusterTypeDescriptor?: boolean,
    fields?: string[],
}

const INIT_FIELD_DELAY = 50;            // ms
/**
 * Abstract class that's available for stepper component to extend.
 * It captures the common logic that should happen to most if not all
 * stepper components.
 */
@Directive()
export abstract class StepFormDirective extends BasicSubscriber implements OnInit {
    wizardName: string;
    formName;
    formGroup: FormGroup;

    edition: AppEdition = AppEdition.TCE;
    validatorEnum = ValidatorEnum;
    errorNotification: string = '';
    configFileNotification: Notification;
    private clusterTypeDescription: string = '';
    modeClusterStandalone: boolean;
    ipFamily: IpFamilyEnum = IpFamilyEnum.IPv4;

    clusterTypeDescriptorUsedInDescription: boolean;

    // This map is made available to HTML pages to display labels before fields
    htmlFieldLabels: Map<string, string>;

    private delayedFieldQueue = [];
    protected eventFileImported: TkgEventType;
    protected eventFileImportError: TkgEventType;

    protected abstract storeUserData();

    // This method is expected to be overridden by any step that provides a dynamic description of itself
    // (dynamic meaning depending on user-entered data). It is public to make it available for testing.
    dynamicDescription(): string {
        return null;
    }

    setStepRegistrantData(stepRegistrantData: StepRegistrantData) {
        this.formName = stepRegistrantData.step;
        this.formGroup = stepRegistrantData.formGroup;
        this.wizardName = stepRegistrantData.wizard;
        this.eventFileImported = stepRegistrantData.eventFileImported;
        this.eventFileImportError = stepRegistrantData.eventFileImportError;
    }

    protected registerDefaultFileImportErrorHandler(eventFailure: TanzuEventType) {
        AppServices.messenger.subscribe<string>(eventFailure, data => {
            // Capture the import file error message
            this.configFileNotification = {
                notificationType: NotificationTypes.ERROR,
                message: data.payload
            };
        });
    }

    ngOnInit(): void {
        this.getFormName();
        this.subscribeToStepCompletedEvents();

        // set branding and cluster type on branding change for base wizard components
        AppServices.messenger.subscribe<EditionData>(TanzuEventType.BRANDING_CHANGED, data => {
                this.edition = data.payload.edition;
                this.setClusterTypeDescriptor(data.payload.clusterTypeDescriptor);
            }, this.unsubscribe);
        this.modeClusterStandalone = AppServices.appDataService.isModeClusterStandalone();
    }

    /**
     * Infer form name from the formGroup object.
     */
    getFormName() {
        if (this.formGroup && this.formGroup.parent && this.formGroup.parent.controls) {
            for (const name of Object.keys(this.formGroup.parent.controls)) {
                if (this.formGroup.parent.controls[name] === this.formGroup) {
                    this.formName = name;
                    break;
                }
            }
        }
    }

    // This method could be protected, since it's primarily intended for subclasses,
    // but since it's helpful for tests to be able to use it, we make it public
    getFieldValue(fieldName: string, suppressWarnings?: boolean): any {
        if (!this.formGroup) {
            if (!suppressWarnings) {
                console.error('getFieldValue(' + fieldName + ') called without a formGroup set');
            }
            return;
        }
        if (!this.formGroup.controls[fieldName]) {
            if (!suppressWarnings) {
                console.error('getFieldValue(' + fieldName + ') called but no control by that name');
            }
            return;
        }
        return this.formGroup.controls[fieldName].value;
    }

    /**
     * Safely looks up the saved value of a control in savedMetadata
     * @param fieldName the name of the control in savedMetadata
     * @param defaultValue the default value if there is no saved value
     */
    getSavedValue(fieldName: string, defaultValue: any) {
        const identifier = this.createUserDataIdentifier(fieldName);
        const entry = AppServices.userDataService.retrieve(identifier);
        let result = defaultValue;
        if (entry !== null && entry !== undefined) {
            result = entry.value;
        }

        return result;
    }

    getRawSavedValue(fieldName: string) {
        const identifier = this.createUserDataIdentifier(fieldName);    // new way
        const entry = AppServices.userDataService.retrieve(identifier);
        const value = entry ? entry.value : null;

        return value;
    }

    /**
     * Safely looks up the saved key of a control in savedMetadata; this will only have been set for listboxes
     * that have a different key from the displayed label
     * @param fieldName the name of the control in savedMetadata
     */
    getSavedKey(fieldName: string): string {
        // NOTE: in new way, we don't need this getSavedKey method at all
        const identifier = this.createUserDataIdentifier(fieldName);
        const entry = AppServices.userDataService.retrieve(identifier);
        const savedKey = entry ? entry.value : null;

        return savedKey;
    }

    hasSavedData() {
        return AppServices.userDataService.hasStoredStepData(this.wizardName, this.formName);
    }

    // This method could be protected, since it's primarily intended for subclasses,
    // but since it's helpful for tests to be able to use it, we make it public
    setFieldValue(fieldName: string, value: any): void {
        const control = this.getControl(fieldName);
        if (control === undefined || control === null) {
            console.log('WARNING: setFieldValue() could not find field ' + fieldName + ' to set value to ' + value);
        } else {
            control.setValue(value);
        }
    }

    protected getControl(fieldName: string): AbstractControl {
        const control = this.formGroup.get(fieldName);
        if (control === undefined || control === null) {
            console.log('WARNING: getControl() could not find field ' + fieldName);
        }
        return control;
    }

    /**
     * Init the field with the saved value. If the initialization has to wait
     * until certain conditions to be satisfied, it automatically add this field
     * to a queue and periodically checks the readiness.
     */
    protected initFieldWithSavedData(fieldName: string): void {
        if (this.isFieldReadyForInitWithSavedValue(fieldName)) {
            const control = this.formGroup.get(fieldName);
            const savedKey = this.getSavedKey(fieldName);
            const savedValue = this.getSavedValue(fieldName, control.value);
            // if a key was saved (for a listbox), we use the key when setting the value of the control (ie the listbox)
            const valueForSettingControl = (savedKey) ? savedKey : savedValue;

            control.setValue(valueForSettingControl, {  emitEvent: false });
            let index;
            if (index = this.delayedFieldQueue.indexOf(fieldName) >= 0) {
                this.delayedFieldQueue.splice(index, 1);
            }
        } else {
            if (this.delayedFieldQueue.indexOf(fieldName) < 0) {
                this.delayedFieldQueue.push(fieldName);
            }
        }
    }

    // scrubPasswordField() should be called AFTER the password field was set from saved data
    protected scrubPasswordField(fieldName: string): void {
        // ensure that the actual field value is not a series of asterisks (****), which is the displayValue kept in local storage
        const passwordControl = this.formGroup.get(fieldName);
        if (passwordControl === undefined || passwordControl === null) {
            console.log('WARNING: scrubPasswordField() is unable to find the field ' + fieldName);
        } else if (this.passwordContainsOnlyAsterisks(passwordControl.value)) {
            passwordControl.setValue('', {onlySelf: true, emitEvent: false});
        }
        // if there is a real password in local storage (say, from import)
        // we erase it from local storage (presuming the caller has already used the value to set the field in the form)
        this.clearFieldSavedData(fieldName);
    }

    private passwordContainsOnlyAsterisks(password: string): boolean {
        if (password === undefined || password === '') {
            return false;
        }
        for (let x = 0; x < password.length; x++) {
            if (password.charAt(x) !== '*') {
                return false;
            }
        }
        return true;
    }

    /**
     * Checks if a field is ready to be initialized with saved data.
     * All fields are ready by default.
     * Sub classes should override this method in order to control
     * the process.
     * @param fieldName the field to be initialized
     * @returns true for fields
     */
    protected isFieldReadyForInitWithSavedValue(fieldName: string): boolean {
        return true;
    }

    protected clearFieldSavedData(fieldName: string) {
        AppServices.userDataService.clear(this.createUserDataIdentifier(fieldName));
    }

    protected saveFieldData(fieldName: string, value: string) {
        AppServices.userDataService.store(this.createUserDataIdentifier(fieldName), {
            display: value,
            value
        });
    }

    private findField(fieldName: string, methodName: string): AbstractControl {
        if (!fieldName) {
            console.warn(`${methodName}(): called with empty fieldName`);
            return null;
        }

        const field = this.formGroup.controls[fieldName];
        if (!field) {
            console.warn(`${methodName}(): unable to find field with name ${fieldName}`);
            return null;
        }
        return field;
    }

    disarmField(fieldName: string, clearSavedData?: boolean, options?: {
        onlySelf?: boolean;
        emitEvent?: boolean;
    }) {
        const field = this.findField(fieldName, 'disarmField');
        if (field) {
            field.clearValidators();
            field.setValue('', options);
            field.updateValueAndValidity(options);
            if (clearSavedData) {
                this.clearFieldSavedData(fieldName);
            }
        }
    }

    resurrectField(fieldName: string, validators: ValidatorFn[], value?: string, options?: {
        onlySelf?: boolean;
        emitEvent?: boolean;
    }) {
        const field = this.findField(fieldName, 'resurrectField');
        if (field) {
            field.setValidators(validators);
            field.updateValueAndValidity(options);
            if (value !== undefined) {
                field.setValue(value || null, options);
            }
        }
    }

    resurrectFieldWithSavedValue(fieldName: string, validators: ValidatorFn[], defaultValue?: string, options?: {
        onlySelf?: boolean;
        emitEvent?: boolean;
    }) {
        this.resurrectField(fieldName, validators, this.getSavedValue(fieldName, defaultValue), options);
    }

    showFormError(formControlname) {
        return this.formGroup.controls[formControlname].invalid &&
            (this.formGroup.controls[formControlname].dirty ||
                this.formGroup.controls[formControlname].touched)
    }

    /**
     * Checks if an array is empty
     */
    isEmptyArray(arr: any[]): boolean {
        return !(arr && arr.length > 0);
    }

    // subclasses can ask us to make sure a StepDescriptionChange is triggered on various occasions
    protected registerStepDescriptionTriggers(triggers: StepDescriptionTriggers) {
        if (triggers.fields) {
            this.registerFieldsAffectingStepDescription(triggers.fields);
        }
        this.clusterTypeDescriptorUsedInDescription = triggers.clusterTypeDescriptor;
    }

    private registerFieldsAffectingStepDescription(fields: string[]) {
        fields.forEach(field => {
            this.registerOnValueChange(field, () => {
                this.triggerStepDescriptionChange();
            });
        })
    }

    /**
     * Registers a callback when a field value changes.
     * This method does more than the "onchange" event handler
     * in that onchange only captures changes through the UI, while
     * this method registers handler for all value change event including
     * the one changed programmatically.
     * @param fieldName the field whose value to be monitored
     * @param callback the function to be called when a value changes.
     */
    registerOnValueChange(fieldName: string, callback: (newValue: any) => void) {
        this.getControl(fieldName).valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(newValue => callback(newValue));
    }

    registerOnIpFamilyChange(fieldName: string, ipv4Validators: ValidatorFn[], ipv6Validators: ValidatorFn[], cb?: () => void) {
        AppServices.messenger.subscribe<IpFamilyEnum>(TanzuEventType.VSPHERE_IP_FAMILY_CHANGE, data => {
                if (data.payload === IpFamilyEnum.IPv4) {
                    this.resurrectField(
                        fieldName,
                        ipv4Validators,
                        this.formGroup.get(fieldName).value,
                        {emitEvent: false, onlySelf: true}
                    );
                } else {
                    this.resurrectField(
                        fieldName,
                        ipv6Validators,
                        this.formGroup.get(fieldName).value,
                        {emitEvent: false, onlySelf: true}
                    );
                }
                this.ipFamily = data.payload;
                if (cb) {
                    cb();
                }
            }, this.unsubscribe);
    }

    protected setControlValueSafely(controlName: string, value: any, options?: {
        onlySelf?: boolean,
        emitEvent?: boolean
    }) {
        const control = this.formGroup.get(controlName);
        if (control) {
            control.setValue(value, options);
        }
    }

    protected clearControlValue(controlName: string) {
        this.setControlValueSafely(controlName, '' , { onlySelf: true, emitEvent: false});
    }

    protected setControlWithSavedValue(controlName: string, defaultValue?: any, options?: {
        onlySelf?: boolean,
        emitEvent?: boolean
    }) {
        const defaultToUse = (defaultValue === undefined || defaultValue === null) ? '' : defaultValue;
        this.setControlValueSafely(controlName, this.getSavedValue(controlName, defaultToUse), options);
    }

    // HTML convenience methods
    //
    get clusterTypeDescriptorTitleCase() {
        return FormUtility.titleCase(this.clusterTypeDescriptor);
    }
    //
    // HTML convenience methods

    protected triggerStepDescriptionChange() {
        const descriptionChangePayload: StepDescriptionChangePayload = {
            wizard: this.wizardName,
            step: this.formName,
            description: this.dynamicDescription(),
        }
        AppServices.messenger.publish({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: descriptionChangePayload,
        });
    }

    // NOTE: this method is public to facilitate testing, otherwise it would be private
    public setClusterTypeDescriptor(descriptor: string) {
        if (this.clusterTypeDescription !== descriptor) {
            this.clusterTypeDescription = descriptor;
            if (this.clusterTypeDescriptorUsedInDescription) {
                this.triggerStepDescriptionChange();
            }
        }
    }

    get clusterTypeDescriptor() {
        return this.clusterTypeDescription;
    }

    private subscribeToStepCompletedEvents() {
        AppServices.messenger.subscribe<StepCompletedPayload>(TanzuEventType.STEP_COMPLETED, event => {
            if (event.payload.wizard === this.wizardName && event.payload.step === this.formName) {
                this.onStepCompleted();
            }
        });
    }

    private onStepCompleted() {
        this.storeUserData();
    }

    // convenience methods
    protected createUserDataIdentifier(field: string): UserDataIdentifier {
        return { wizard: this.wizardName, step: this.formName, field};
    }

    protected storeUserDataFromMapping(stepMapping: StepMapping, objectRetrievalMap?: Map<string, (key: any) => any>) {
        AppServices.userDataFormService.storeFromMapping(this.wizardName, this.formName, stepMapping, this.formGroup,
            objectRetrievalMap);
    }

    protected storeDefaultDisplayOrder(stepMapping: StepMapping) {
        this.storeDisplayOrder(this.defaultDisplayOrder(stepMapping));
    }

    protected storeDisplayOrder(displayOrder: string[]) {
        AppServices.userDataService.storeStepDisplayOrder(this.wizardName, this.formName, displayOrder);
    }

    protected storeDefaultLabels(stepMapping: StepMapping) {
        this.storeLabels(this.defaultLabels(stepMapping));
    }

    protected storeLabels(titles: Map<string, string>) {
        AppServices.userDataService.storeStepLabels(this.wizardName, this.formName, titles);
    }

    protected defaultDisplayOrder(stepMapping: StepMapping): string[] {
        return AppServices.fieldMapUtilities.getLabeledFieldsWithStoredData(this.wizardName, this.formName, stepMapping);
    }

    protected defaultLabels(stepMapping: StepMapping): Map<string, string> {
        return AppServices.fieldMapUtilities.getFieldLabelMap(stepMapping);
    }

    // This method is designed to expose the protected unsubscribe field (from our base class) to allow its use in subscribing to pipes
    get unsubscribeOnDestroy(): Subject<void> {
        return this.unsubscribe;
    }

    protected registerDefaultFileImportedHandler(eventSuccess: TanzuEventType, stepMapping: StepMapping,
        objectRetrievalMap?: Map<string, (string) => any>, customRestorerMap?: Map<string, (any) => void>) {
        AppServices.messenger.subscribe<string>(eventSuccess,
            this.defaultFileImportedHandler(stepMapping, objectRetrievalMap, customRestorerMap));
    }

    // This is a convenience method for child classes that want to register a callback based on this behavior PLUS something of their own
    protected defaultFileImportedHandler(stepMapping: StepMapping, objectRetrievalMap?: Map<string, (string) => any>,
                                         customRestorerMap?: Map<string, (any) => void>)
        : (event: TanzuEvent<any>) => void {
        const step = this;
        return data => {
            this.configFileNotification = {
                notificationType: NotificationTypes.SUCCESS,
                message: data.payload
            };
            // The file import saves the data to local storage, so we reinitialize this step's form from there
            AppServices.userDataFormService.restoreForm(step.wizardName, step.formName, step.formGroup, stepMapping, objectRetrievalMap,
                customRestorerMap);
        }
    }
}
