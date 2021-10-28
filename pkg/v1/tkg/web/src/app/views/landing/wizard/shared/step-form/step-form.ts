import { Input, OnInit, Directive, Self, Optional } from '@angular/core';
import { FormGroup, ValidatorFn } from '@angular/forms';

import { ValidatorEnum } from './../constants/validation.constants';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { FormMetaDataStore, FormMetaData } from '../FormMetaDataStore';
import { TkgEvent, TkgEventType } from 'src/app/shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';

import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { EditionData } from 'src/app/shared/service/branding.service';
import { IpFamilyEnum } from 'src/app/shared/constants/app.constants';

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
    clusterTypeDescriptor: string;
    modeClusterStandalone: boolean;
    ipFamily: IpFamilyEnum = IpFamilyEnum.IPv4;

    private delayedFieldQueue = [];

    ngOnInit(): void {
        this.getFormName();
        this.savedMetadata = FormMetaDataStore.getMetaData(this.formName)
        FormMetaDataStore.updateFormList(this.formName);
        // waits 10 milliseconds so that other setTimeout calls finish first
        setTimeout(_ => {
            this.setSavedDataAfterLoad();
        }, 10);

        // set branding and cluster type on branding change for base wizard components
        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                const content: EditionData = data.payload;
                this.edition = content.edition;
                this.clusterTypeDescriptor = data.payload.clusterTypeDescriptor;
            });
        this.modeClusterStandalone = Broker.appDataService.isModeClusterStandalone();
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
     * @param key the key of the control in savedMetadata
     * @param defaultValue the default value if there is no saved value
     */
    getSavedValue(key: string, defaultValue: any) {
        const ret = (this.savedMetadata && this.savedMetadata[key]
            && this.savedMetadata[key].displayValue) ? this.savedMetadata[key].displayValue : defaultValue;
        return (typeof defaultValue === "boolean") ? ret === "yes" : ret
    }

    hasSavedData() {
        return this.savedMetadata != null
    }

    /**
     * Init the field with the saved value. If the initialization has to wait
     * until certain conditions to be satisfied, it automatically add this field
     * to a queue and periodically checks the readiness.
     */
    protected initFieldWithSavedValue(fieldName: string): void {
        if (this.isFieldReadyForInitWithSavedValue(fieldName)) {
            const control = this.formGroup.get(fieldName);
            control.setValue(this.getSavedValue(fieldName, control.value));
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

    /**
     * Start the process of initializing the fields as soon as they become ready.
     */
    protected startProcessDelayedFieldInit() {
        if (this.delayedFieldQueue && this.delayedFieldQueue.length > 0) {
            this.delayedFieldQueue.forEach(fieldName => this.initFieldWithSavedValue(fieldName));
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
    setSavedDataAfterLoad() {
        if (this.hasSavedData()) {
            for (const [key, control] of Object.entries(this.formGroup.controls)) {
                if (this.getSavedValue(key, '') !== '') {
                    // control.setValue(this.getSavedValue(key, control.value));
                    this.initFieldWithSavedValue(key);
                }
            }
        }

        this.startProcessDelayedFieldInit();
    }

    clearFieldSavedData(fieldName: string) {
        FormMetaDataStore.deleteMetaDataEntry(this.formName, fieldName);
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
            this.formGroup.controls[fieldName].setValue(value || null);
        } else {
            console.warn(`resurrectField(): Unable to find field with name ${fieldName}`);
        }
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
}
