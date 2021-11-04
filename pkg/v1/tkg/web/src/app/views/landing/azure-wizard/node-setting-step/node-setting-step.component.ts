/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { takeUntil } from 'rxjs/operators';

/**
 * App imports
 */
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { AzureInstanceType } from 'src/app/swagger/models';
import { AppEdition } from 'src/app/shared/constants/branding.constants';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {

    nodeTypes: AzureInstanceType[] = [];
    nodeType: string;
    currentRegion = "US-WEST";
    displayForm = false;

    constructor(private validationService: ValidationService,
                private azureWizardFormService: AzureWizardFormService) {
        super();
        this.nodeTypes = [];
    }

    buildForm() {
        this.formGroup.addControl(
            'controlPlaneSetting',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'devInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'prodInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'devInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );

        this.formGroup.addControl(
            'managementClusterName',
            new FormControl('', [
                this.validationService.isValidClusterName()
            ])
        );

        if (!this.modeClusterStandalone) {
            this.formGroup.addControl(
                'workerNodeInstanceType',
                new FormControl('', [
                    Validators.required
                ])
            );
        }

        this.formGroup.addControl(
            'machineHealthChecksEnabled',
            new FormControl(true, [])
        );
    }

    initForm() {
        this.azureWizardFormService.getErrorStream(TkgEventType.AZURE_GET_INSTANCE_TYPES)
        .pipe(takeUntil(this.unsubscribe))
        .subscribe(error => {
            this.errorNotification = error;
        });

        this.azureWizardFormService.getDataStream(TkgEventType.AZURE_GET_INSTANCE_TYPES)
        .pipe(takeUntil(this.unsubscribe))
        .subscribe((instanceTypes: AzureInstanceType[]) => {
            this.nodeTypes = instanceTypes.sort();
            if (!this.modeClusterStandalone && this.nodeTypes.length === 1) {
                this.formGroup.get('workerNodeInstanceType').setValue(this.nodeTypes[0].name);
            }
        });

        if (this.edition !== AppEdition.TKG) {
            this.resurrectField('managementClusterName',
                [Validators.required, this.validationService.isValidClusterName()],
                this.formGroup.get('managementClusterName').value);
        }
    }

    toggleValidations() {
        setTimeout(_ => {
            this.displayForm = true;
            const controlPlaneSettingControl = this.formGroup.get('controlPlaneSetting');
            if (controlPlaneSettingControl) {
                controlPlaneSettingControl.valueChanges.subscribe(data => {
                    if (data === 'dev') {
                        this.setDevCardValidations();
                    } else if (data === 'prod') {
                        this.setProdCardValidations();
                    }
                });
            } else {
                console.log('WARNING: azure-wizard.node-setting-step.toggleValidations() unable to find controlPlaneSettingControl!');
            }
        });
    }

    setDevCardValidations() {
        this.nodeType = 'dev';
        const devInstanceTypeControl = this.formGroup.get('devInstanceType');
        if (devInstanceTypeControl) {
            devInstanceTypeControl.setValidators([Validators.required]);
            devInstanceTypeControl.setValue(this.nodeTypes.length === 1 ? this.nodeTypes[0].name : '');
            devInstanceTypeControl.updateValueAndValidity();
        }
        const prodInstanceTypeControl = this.formGroup.controls['prodInstanceType'];
        if (prodInstanceTypeControl) {
            prodInstanceTypeControl.clearValidators();
            prodInstanceTypeControl.setValue('');
            prodInstanceTypeControl.updateValueAndValidity();
        }
    }

    setProdCardValidations() {
        this.nodeType = 'prod';
        const devInstanceTypeControl = this.formGroup.get('devInstanceType');
        if (devInstanceTypeControl) {
            devInstanceTypeControl.setValue('');
            devInstanceTypeControl.updateValueAndValidity();
            devInstanceTypeControl.clearValidators();
        }
        const prodInstanceTypeControl = this.formGroup.controls['prodInstanceType'];
        if (prodInstanceTypeControl) {
            prodInstanceTypeControl.setValidators([Validators.required]);
            prodInstanceTypeControl.setValue(this.nodeTypes.length === 1 ? this.nodeTypes[0].name : '');
            prodInstanceTypeControl.updateValueAndValidity();
        }
    }

    ngOnInit() {
        super.ngOnInit();
        this.buildForm();
        this.initForm();
        this.toggleValidations();
        this.initFormWithSavedData();
    }

    initFormWithSavedData() {
        this.cardClick(this.getSavedValue('devInstanceType', '') === '' ? 'prod' : 'dev');
        this.getSavedValue('devInstanceType', '') === '' ? this.setProdCardValidations() : this.setDevCardValidations()
        super.initFormWithSavedData();
    }

    cardClick(envType: string) {
        this.setControlValueSafely('controlPlaneSetting', envType);
    }

    getEnvType(): string {
        return this.formGroup.controls['controlPlaneSetting'].value;
    }

}
