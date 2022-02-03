// Angular imports
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';

// App imports
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { MetadataStepMapping } from './metadata-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { ValidationService } from '../../../validation/validation.service';
import { FormUtils } from '../../../utils/form-utils';

const LABEL_KEY_NAME = 'newLabelKey';
const LABEL_VALUE_NAME = 'newLabelValue';

@Component({
    selector: 'app-metadata-step',
    templateUrl: './metadata-step.component.html',
    styleUrls: ['./metadata-step.component.scss']
})
export class MetadataStepComponent extends StepFormDirective implements OnInit {
    labels: Map<string, string> = new Map<string, string>();
    keySet: Set<string> = new Set();
    savedKeySet: Set<string> = new Set();
    labelCounter: number = 0;

    constructor(private validationService: ValidationService,
                private fieldMapUtilities: FieldMapUtilities) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, MetadataStepMapping);
        this.registerStepDescriptionTriggers({
            fields: ['clusterLocation'],
            clusterTypeDescriptor: true,
        })

        this.initFormWithSavedData();
        if (this.labels.size === 0) {
            this.addLabel();
        }
    }

    initFormWithSavedData() {
        const savedLabelsString = this.getSavedValue('clusterLabels', '');
        if (savedLabelsString !== '') {
            const savedLabelsArray = savedLabelsString.split(', ')
            savedLabelsArray.map(label => {
                const labelArray = label.split(':');
                this.addLabel(labelArray[0], labelArray[1]);
                this.savedKeySet.add(LABEL_KEY_NAME + this.labelCounter);
            });
        }
        super.initFormWithSavedData();
    }

    addLabel(key?: string, value?: string) {
        this.labelCounter++;
        this.labels.set(LABEL_KEY_NAME + this.labelCounter, LABEL_VALUE_NAME + this.labelCounter);
        this.keySet.add(LABEL_KEY_NAME + this.labelCounter);
        FormUtils.addControl(
            this.formGroup,
            LABEL_KEY_NAME + this.labelCounter,
            new FormControl(key || '', [
                this.validationService.isValidLabelOrAnnotation(),
                this.validationService.isUniqueLabel(
                    this.formGroup,
                    this.keySet,
                    LABEL_KEY_NAME + this.labelCounter)
            ])
        );

        FormUtils.addControl(
            this.formGroup,
            LABEL_VALUE_NAME + this.labelCounter,
            new FormControl(value || '', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        // Label value depends on Label key. e.g.: if label key is not empty, then label value is required
        this.onChangeWithDependentField(LABEL_KEY_NAME + this.labelCounter, LABEL_VALUE_NAME + this.labelCounter);
        // Label key depends on Label value. e.g.: if label value is not empty, then label key is required
        this.onChangeWithDependentField(LABEL_VALUE_NAME + this.labelCounter, LABEL_KEY_NAME + this.labelCounter);
        this.validateAllLabels();
    }

    /**
     * @method onChangeWithDependentField
     * make the dependent field is required if the indepdent field is not empty.
     * @param fieldName is a independent field which determines if the dependent field is required.
     * @param dependentFieldName is dependent on the independent field.
     */
    onChangeWithDependentField(fieldName: string, dependentFieldName: string) {
        const control = this.formGroup.get(dependentFieldName);
        this.registerOnValueChange(fieldName, (data) => {
            if (data !== '') {
                if (!control.hasValidator(Validators.required)) {
                    control.addValidators(Validators.required);
                    control.markAsPending(); // validation will not be triggered until the field is touched.
                    control.setErrors({required: true});
                }
            } else {
                control.removeValidators(Validators.required);
            }
            this.validateAllLabels(); // all the same label keys can show error message.
        });
    }

    validateAllLabels () {

        // The setTimeout wrapper ensures that validation logic will run after a new label field is added.
        setTimeout(_ => {
            for (const [labelKey, labelVal] of this.labels) {
                const key = this.formGroup.get(labelKey);
                const val = this.formGroup.get(labelVal);
                if (key) {
                    if (this.savedKeySet.has(labelKey)) {
                        key.markAsTouched();
                    }
                    key.updateValueAndValidity();
                }
                if (val) {
                    val.updateValueAndValidity();
                }
            }
        });
    }

    deleteLabel(key: string) {
        this.formGroup.removeControl(key);
        this.formGroup.removeControl(this.labels.get(key));
        this.labels.delete(key);
        this.keySet.delete(key);
        this.formGroup.get('clusterLabels').setValue(this.labels);
    }

    /**
     * Get the current value of 'clusterLabels'
     */
    get clusterLabelsValue() {
        let labelStr = '';
        for (const [labelKey, labelVal] of this.labels) {
            const key = this.formGroup.get(labelKey).value;
            const val = this.formGroup.get(labelVal).value;
            if (key && val) {
                labelStr += key + ':' + val + ', '
            }
        }
        return labelStr.slice(0, -2);
    }

    dynamicDescription(): string {
        const clusterLocation = this.getFieldValue('clusterLocation', true);
        return clusterLocation ? 'Location: ' + clusterLocation : 'Specify metadata for the ' + this.clusterTypeDescriptor + ' cluster';
    }
}
