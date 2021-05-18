/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl } from '@angular/forms';

/**
 * App imports
 */
import { ValidationService } from '../../../validation/validation.service';
import { StepFormDirective } from '../../../step-form/step-form';
import { Messenger } from '../../../../../../../shared/service/Messenger';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';

@Component({
    selector: 'app-metadata-step',
    templateUrl: './metadata-step.component.html',
    styleUrls: ['./metadata-step.component.scss']
})
export class MetadataStepComponent extends StepFormDirective implements OnInit {
    labels: Map<String, String> = new Map<String, String>();

    constructor(private validationService: ValidationService,
        private wizardFormService: VSphereWizardFormService, private messenger: Messenger) {

        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.formGroup.addControl(
            'clusterLocation',
            new FormControl('', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        this.formGroup.addControl(
            'clusterDescription',
            new FormControl('', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        this.formGroup.addControl(
            'newLabelKey',
            new FormControl('', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        this.formGroup.addControl(
            'newLabelValue',
            new FormControl('', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        this.formGroup.addControl(
            'clusterLabels',
            new FormControl('', [])
        );
    }

    setSavedDataAfterLoad() {
        const savedLabelsString = this.getSavedValue('clusterLabels', '');
        if (savedLabelsString !== '') {
            const savedLabelsArray = savedLabelsString.split(', ')
            savedLabelsArray.map(label => {
                const labelArray = label.split(':');
                this.labels.set(labelArray[0], labelArray[1]);
            });
        }
        super.setSavedDataAfterLoad();
    }

    addLabel(key: string, value: string) {
        if (key === '' || value === '') {
            this.errorNotification = `Key and value for Labels are required.`;
        } else if (!this.labels.has(key)) {
            this.labels.set(key, value);
            this.formGroup.get('clusterLabels').setValue(this.labels);
            this.formGroup.controls['newLabelKey'].setValue('');
            this.formGroup.controls['newLabelValue'].setValue('');
        } else {
            this.errorNotification = `A Label with the same key already exists.`;
        }
    }

    deleteLabel(key: string) {
        this.labels.delete(key);
        this.formGroup.get('clusterLabels').setValue(this.labels);
    }

    /**
     * Get the current value of 'clusterLabels'
     */
    get clusterLabelsValue() {
        let labelsStr: string = '';
        this.labels.forEach((value: string, key: string) => {
            labelsStr += key + ':' + value + ', '
        });
        return labelsStr.slice(0, -2);
    }

    /**
     * @method getDisabled
     * helper method to get if add btn should be disabled
     */
    getDisabled(): boolean {
        return !(this.formGroup.get('newLabelKey').valid &&
            this.formGroup.get('newLabelValue').valid);
    }
}
