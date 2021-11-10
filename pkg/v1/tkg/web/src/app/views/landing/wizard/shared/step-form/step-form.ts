import {Directive, Input, OnInit} from '@angular/core';
import {FormGroup, ValidatorFn} from '@angular/forms';

import {ValidatorEnum} from './../constants/validation.constants';
import {BasicSubscriber} from 'src/app/shared/abstracts/basic-subscriber';
import {FormMetaData, FormMetaDataStore} from '../FormMetaDataStore';
import {TkgEvent, TkgEventType} from 'src/app/shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';

import {distinctUntilChanged, takeUntil} from 'rxjs/operators';
import {AppEdition} from 'src/app/shared/constants/branding.constants';
import {EditionData} from 'src/app/shared/service/branding.service';
import {IpFamilyEnum} from 'src/app/shared/constants/app.constants';

const INIT_FIELD_DELAY = 50;            // ms
/**
 * Abstract class that's available for stepper component to extend.
 * It captures the common logic that should happen to most if not all
 * stepper components.
 */
@Directive()
export abstract class StepFormDirective extends BasicSubscriber implements OnInit {

    @Input() formName;
    @Input() formGroup: FormGroup;
    @Input() savedMetadata: { [fieldName: string]: FormMetaData };

    edition: AppEdition = AppEdition.TCE;
    validatorEnum = ValidatorEnum;
    errorNotification: string;
    errorImportFile: string;
    clusterTypeDescriptor: string;
    modeClusterStandalone: boolean;
    ipFamily: IpFamilyEnum = IpFamilyEnum.IPv4;

    private delayedFieldQueue = [];

    ngOnInit(): void {
        this.getFormName();
        this.savedMetadata = FormMetaDataStore.getMetaData(this.formName);
        FormMetaDataStore.updateFormList(this.formName);

        console.log('ngOnInit for form: ' + this.formName);

        // set branding and cluster type on branding change for base wizard components
        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                const content: EditionData = data.payload;
                this.edition = content.edition;
                this.clusterTypeDescriptor = data.payload.clusterTypeDescriptor;
            });
        this.modeClusterStandalone = Broker.appDataService.isModeClusterStandalone();

        Broker.messenger.getSubject(TkgEventType.CONFIG_FILE_IMPORT_ERROR)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                // Capture the import file error message
                this.errorImportFile = data.payload;
            });
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

    /**
     * Safely looks up the saved value of a control in savedMetadata
     * @param fieldName the name of the control in savedMetadata
     * @param defaultValue the default value if there is no saved value
     */
    getSavedValue(fieldName: string, defaultValue: any) {
        const value = this.getRawSavedValue(fieldName);
        const result = (value) ? value : defaultValue;

        const shouldReturnBooleanValue = typeof defaultValue === "boolean";
        if (shouldReturnBooleanValue) {
            const booleanResult = result === "yes" || result === true;
            return booleanResult;
        }
        return result;
    }

    getRawSavedValue(fieldName: string) {
        const savedValue = this.savedMetadata && this.savedMetadata[fieldName] && this.savedMetadata[fieldName].displayValue;
        const savedKey = this.savedMetadata && this.savedMetadata[fieldName] && this.savedMetadata[fieldName].key;
        const result = (savedKey) ? savedKey : savedValue;
        return result;
    }

    /**
     * Safely looks up the saved key of a control in savedMetadata; this will only have been set for listboxes
     * that have a different key from the displayed label
     * @param fieldName the name of the control in savedMetadata
     */
    getSavedKey(fieldName: string): string {
        const savedKey = this.savedMetadata && this.savedMetadata[fieldName] && this.savedMetadata[fieldName].key;
        return savedKey;
    }

    hasSavedData() {
        return this.savedMetadata != null
    }

    protected getFieldValue(fieldName: string): any {
        const control = this.formGroup.get(fieldName);
        if (control === undefined || control === null) {
            console.log('WARNING: getFieldValue() could not find field ' + fieldName );
            return '';
        }
        return control.value;
    }

    protected setFieldValue(fieldName: string, value: any): void {
        const control = this.formGroup.get(fieldName);
        if (control === undefined || control === null) {
            console.log('WARNING: setFieldValue() could not find field ' + fieldName + ' to set value to ' + value);
        } else {
            control.setValue(value);
        }
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

            control.setValue(valueForSettingControl);
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
            passwordControl.setValue('');
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
     * Start the process of initializing the fields as soon as they become ready.
     */
    protected startProcessDelayedFieldInit() {
        if (this.delayedFieldQueue && this.delayedFieldQueue.length > 0) {
            this.delayedFieldQueue.forEach(fieldName => this.initFieldWithSavedData(fieldName));
        }
        if (this.delayedFieldQueue && this.delayedFieldQueue.length > 0) {
            setTimeout(this.startProcessDelayedFieldInit.bind(this), INIT_FIELD_DELAY);
        }
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

    /**
     * Inits form fields with saved data if any;
     */
    initFormWithSavedData() {
        console.log('step-form.initFormWithSavedData for form: ' + this.formName);
        if (this.hasSavedData()) {
            for (const [controlName, control] of Object.entries(this.formGroup.controls)) {
                this.initFieldWithSavedData(controlName);
            }
        }
        this.startProcessDelayedFieldInit();
    }

    protected clearFieldSavedData(fieldName: string) {
        FormMetaDataStore.deleteMetaDataEntry(this.formName, fieldName);
    }

    // TODO: this method saves the value as both the display value and the key, but that's sloppy
    protected saveFieldData(fieldName: string, value: string) {
        FormMetaDataStore.saveMetaDataEntry(this.formName, fieldName, {
            label: '',
            displayValue: value,
            key: value
        })
    }

    disarmField(fieldName: string, clearSavedData: boolean) {
        if (fieldName && this.formGroup.controls[fieldName]) {
            this.formGroup.controls[fieldName].clearValidators();
            this.formGroup.controls[fieldName].setValue('');
            this.formGroup.controls[fieldName].updateValueAndValidity();
            if (clearSavedData) {
                this.clearFieldSavedData(fieldName);
            }
        } else {
            console.warn(`disarmField(): Unable to find field with name ${fieldName}`);
        }
    }

    resurrectField(fieldName: string, validators: ValidatorFn[], value?: string) {
        if (fieldName && this.formGroup.controls[fieldName]) {
            this.formGroup.controls[fieldName].setValidators(validators);
            this.formGroup.controls[fieldName].updateValueAndValidity();
            if (fieldName === 'vnetNameExisting') {
                console.log('!   resurrectField() is setting vnetNameExisting to ' + value);
            }
            this.formGroup.controls[fieldName].setValue(value || null);
        } else {
            console.warn(`resurrectField(): Unable to find field with name ${fieldName}`);
        }
    }

    resurrectFieldWithSavedValue(fieldName: string, validators: ValidatorFn[], defaultValue?: string) {
        this.resurrectField(fieldName, validators, this.getSavedValue(fieldName, defaultValue));
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

    /**
     * Registers a callback when a field value changes.
     * This method does more than the "onchange" event handler
     * in that onchange only captures changes through the UI, while
     * this method registers handler for all value change event including
     * the one changed programatically.
     * @param fieldName the field whose value to be monitored
     * @param callback the function to be called when a value changes.
     */
    registerOnValueChange(fieldName: string, callback: (newValue: any) => void) {
        this.formGroup.get(fieldName).valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(newValue => callback(newValue));
    }

    registerOnIpFamilyChange(fieldName: string, ipv4Validators: ValidatorFn[], ipv6Validators: ValidatorFn[], cb?: () => void) {
        Broker.messenger.getSubject(TkgEventType.IP_FAMILY_CHANGE)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                if (data.payload === IpFamilyEnum.IPv4) {
                    this.resurrectField(
                        fieldName,
                        ipv4Validators,
                        this.formGroup.get(fieldName).value
                    );
                } else {
                    this.resurrectField(
                        fieldName,
                        ipv6Validators,
                        this.formGroup.get(fieldName).value
                    );
                }
                this.ipFamily = data.payload;
                if (cb) {
                    cb();
                }
            });
    }

    protected setControlValueSafely(controlName: string, value: any) {
        const control = this.formGroup.get(controlName);
        if (control) {
            control.setValue(value);
        }
    }

    protected setControlWithSavedValue(controlName: string, defaultValue?: any) {
        this.setControlValueSafely(controlName, this.getSavedValue(controlName, (defaultValue) ? defaultValue : ''));
    }
}
