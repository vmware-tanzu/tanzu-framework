/**
 * Angular Modules
 */
import { Component, OnInit, Input} from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { takeUntil } from 'rxjs/operators';

/**
 * App imports
 */

import { StepFormDirective } from '../../../step-form/step-form';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { VSphereVirtualMachine } from 'src/app/swagger/models/v-sphere-virtual-machine.model';
import { AwsWizardFormService } from 'src/app/shared/service/aws-wizard-form.service';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import Broker from 'src/app/shared/service/broker';
import { Observable } from 'rxjs/internal/Observable';
import { AWSVirtualMachine, AzureVirtualMachine } from 'src/app/swagger/models';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { OsImageStepMapping } from './os-image-step.fieldmapping';
import { StepMapping } from '../../../field-mapping/FieldMapping';

export abstract class SharedOsImageStepComponent extends StepFormDirective {
    wizardFormService: VSphereWizardFormService|AwsWizardFormService|AzureWizardFormService;
    enableNonTemplateAlert: boolean;
    noImageAlertMessage: string;
    osImageTooltipContent: string;
    eventType: TkgEventType;
    // TODO: refactor to use generics instead of VSphereVirtualMachine|AWSVirtualMachine|AzureVirtualMachine
    osImages: Array<VSphereVirtualMachine|AWSVirtualMachine|AzureVirtualMachine>;
    loadingOsTemplate: boolean = false;
    nonTemplateAlert: boolean = false;
    tkrVersion: Observable<string>;

    protected constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super();
        this.tkrVersion = Broker.appDataService.getTkrVersion();
    }

    // This method allows child classes to set the inputs (rather than having them passed as part of an HTML component tag).
    // This allows the step to follow the same pattern as all the other steps, which only take formGroup and formName as inputs.
    protected abstract setProviderInputs();

    private customizeForm() {
        /**
         * Whenever data center selection changes, reset the relevant fields
         */
        Broker.messenger.getSubject(TkgEventType.DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.resetFieldsUponDCChange();
            });

        this.wizardFormService.getErrorStream(this.eventType)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.wizardFormService.getDataStream(this.eventType)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((images: Array<VSphereVirtualMachine|AWSVirtualMachine|AzureVirtualMachine>) => {
                this.osImages = images;
                this.loadingOsTemplate = false;
                if (this.osImages.length === 1) {
                    this.formGroup.get('osImage').setValue(images[0]);
                }
            });
    }

    // onInit() should be called from subclass' ngOnInit()
    protected onInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, OsImageStepMapping);
        this.setProviderInputs();
        this.customizeForm();
        this.initFormWithSavedData();
    }

    /**
     * Reset relevant fields upon data center selection change.
     */
    resetFieldsUponDCChange() {
        const fieldsToReset = ['osImage'];
        fieldsToReset.forEach(f => this.formGroup.get(f).setValue(""));
    }

    /**
     * @method retrieveOsImages
     * helper method to retrieve and preload list of available OS images from connected VC environment.
     * emits list of OS images to wizard-data.service
     */
    retrieveOsImages() {
        this.loadingOsTemplate = true;
        this.nonTemplateAlert = false;
        this.resetFieldsUponDCChange();
        Broker.messenger.publish({
            type: this.eventType
        });
    }

    /**
     * @method onOptionsSelected
     * @param name
     * helper method to determine if osImage.isTemplate is true or false; if false show warning
     */
    onOptionsSelected() {
        this.nonTemplateAlert = !this.formGroup.value["osImage"].isTemplate;
    }

    dynamicDescription(): string {
        if (this.getFieldValue('osImage', true) && this.getFieldValue('osImage').name) {
            return 'OS Image: ' + this.getFieldValue('osImage').name;
        }
        return 'Specify the OS Image';
    }
}
