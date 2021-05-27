/**
 * Angular Modules
 */
import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
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

@Component({
    selector: 'app-os-image-step',
    templateUrl: './os-image-step.component.html',
    styleUrls: ['./os-image-step.component.scss']
})
export class SharedOsImageStepComponent extends StepFormDirective implements OnInit {
    @Input() tkrVersion: string;
    @Input() wizardFormService: VSphereWizardFormService|AwsWizardFormService|AzureWizardFormService;
    @Input() type: string;
    @Input() enableNonTemplateAlert: boolean;
    @Input() noImageAlertMessage: string;
    @Input() osImageTooltipContent: string;

    osImages: Array<VSphereVirtualMachine|AwsWizardFormService|AzureWizardFormService>;
    loadingOsTemplate: boolean = false;
    nonTemplateAlert: boolean = false;

    constructor() {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.formGroup.addControl(
            'osImage',
            new FormControl('', [
                Validators.required
            ])
        );

        this.wizardFormService.getErrorStream(TkgEventType[`${this.type}_GET_OS_IMAGES`])
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.wizardFormService.getDataStream(TkgEventType[`${this.type}_GET_OS_IMAGES`])
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((images: Array<VSphereVirtualMachine|AwsWizardFormService|AzureWizardFormService>) => {
                this.osImages = images;
                this.loadingOsTemplate = false;
            });

        /**
         * Whenever data center selection changes, reset the relevant fields
         */
        Broker.messenger.getSubject(TkgEventType.DATACENTER_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.resetFieldsUponDCChange();
            });
    }

    /**
     * Reset relavent fields upon data center selection change.
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
            type: TkgEventType[`${this.type}_GET_OS_IMAGES`]
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
}
