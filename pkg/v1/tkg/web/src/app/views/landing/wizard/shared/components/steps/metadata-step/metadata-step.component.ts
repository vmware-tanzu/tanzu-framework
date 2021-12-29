// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { MetadataStepMapping } from './metadata-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { ValidationService } from '../../../validation/validation.service';

@Component({
    selector: 'app-metadata-step',
    templateUrl: './metadata-step.component.html',
    styleUrls: ['./metadata-step.component.scss']
})
export class MetadataStepComponent extends StepFormDirective implements OnInit {
    labels: Map<String, String> = new Map<String, String>();

    constructor(private validationService: ValidationService,
                private fieldMapUtilities: FieldMapUtilities) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, MetadataStepMapping);
        this.registerFieldsAffectingStepDescription(['clusterLocation']);

        this.initFormWithSavedData();
    }

    initFormWithSavedData() {
        const savedLabelsString = this.getSavedValue('clusterLabels', '');
        if (savedLabelsString !== '') {
            const savedLabelsArray = savedLabelsString.split(', ')
            savedLabelsArray.map(label => {
                const labelArray = label.split(':');
                this.labels.set(labelArray[0], labelArray[1]);
            });
        }
        super.initFormWithSavedData();
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

    protected clusterTypeDescriptorAffectsStepDescription(): boolean {
        return true;
    }

    dynamicDescription(): string {
        const clusterLocation = this.getFieldValue('clusterLocation', true);
        return clusterLocation ? 'Location: ' + clusterLocation : 'Specify metadata for the ' + this.clusterTypeDescriptor + ' cluster';
    }
}
