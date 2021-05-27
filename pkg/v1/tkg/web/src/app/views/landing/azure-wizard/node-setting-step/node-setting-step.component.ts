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
import { NodeType, awsNodeTypes } from '../../wizard/shared/constants/wizard.constants';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { AzureInstanceType } from 'src/app/swagger/models';

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
        this.nodeTypes = [...awsNodeTypes];
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

        this.formGroup.addControl(
            'workerNodeInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );

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
        });

    }

    toggleValidations() {
        setTimeout(_ => {
            this.displayForm = true;
            this.formGroup.get('controlPlaneSetting').valueChanges.subscribe(data => {
                if (data === 'dev') {
                    this.setDevCardValidations();
                } else if (data === 'prod') {
                    this.setProdCardValidations();
                }

            });
        });
    }

    setDevCardValidations() {
        this.nodeType = 'dev';
        this.formGroup.get('devInstanceType').setValidators([
            Validators.required
        ]);
        this.formGroup.controls['prodInstanceType'].clearValidators();
        this.formGroup.controls['prodInstanceType'].setValue('');
        this.formGroup.get('devInstanceType').updateValueAndValidity();
        this.formGroup.controls['prodInstanceType'].updateValueAndValidity();
    }

    setProdCardValidations() {
        this.nodeType = 'prod';
        this.formGroup.controls['prodInstanceType'].setValidators([
            Validators.required
        ]);
        this.formGroup.get('devInstanceType').clearValidators();
        this.formGroup.controls['devInstanceType'].setValue('');
        this.formGroup.get('devInstanceType').updateValueAndValidity();
        this.formGroup.controls['prodInstanceType'].updateValueAndValidity();
    }

    ngOnInit() {
        super.ngOnInit();
        this.buildForm();
        this.initForm();
        this.toggleValidations();
    }

    setSavedDataAfterLoad() {
        this.cardClick(this.getSavedValue('devInstanceType', '') === '' ? 'prod' : 'dev');
        this.getSavedValue('devInstanceType', '') === '' ? this.setProdCardValidations() : this.setDevCardValidations()
        super.setSavedDataAfterLoad();
    }

    cardClick(envType: string) {
        this.formGroup.controls['controlPlaneSetting'].setValue(envType);
    }

    getEnvType(): string {
        return this.formGroup.controls['controlPlaneSetting'].value;
    }

}
