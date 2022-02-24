import { Component, Input, OnInit } from '@angular/core';
import { FormArray, FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { StepFormDirective } from '../../../step-form/step-form';
import { FormUtils } from '../../../utils/form-utils';
import { ValidationService } from '../../../validation/validation.service';
import { MetadataField } from '../../steps/metadata-step/metadata-step.fieldmapping';

@Component({
    selector: 'app-labels',
    templateUrl: './labels.component.html',
    styleUrls: ['./labels.component.scss']
})
export default class LabelsComponent extends StepFormDirective implements OnInit {

    @Input() formName: string;
    @Input() formGroup: FormGroup;
    @Input() labels: Map<string, string> = new Map<string, string>();
    @Input() loadConnected: boolean = false;

    constructor(private validationService: ValidationService, private formBuilder: FormBuilder) {
        super();
    }
    ngOnInit() {
        this.formGroup.addControl('labelList',
            this.formBuilder.array([]));

        if (this.labels.size === 0) {
            this.addLabel();
        } else {
            this.labels.forEach((value, key) => {
                this.addLabel(key, value);
            });
        }
    }
    addLabel(labelKey?: string, labelValue?: string) {
        const labelGroup = this.formBuilder.group({
            key: new FormControl(labelKey || '', [
                this.validationService.isValidLabelOrAnnotation(),
                this.validationService.isUniqueLabel(this.labelArray)]),
            value: new FormControl(labelValue || '',
                this.validationService.isValidLabelOrAnnotation())
        })
        const keyControl = labelGroup.get('key') as FormControl;
        const valueControl = labelGroup.get('value') as FormControl;
        this.labelArray.push(labelGroup);

        // Label value depends on Label key. e.g.: if label key is not empty, then label value is required
        this.onChangeWithDependentField(keyControl, valueControl);
        // Label key depends on Label value. e.g.: if label value is not empty, then label key is required
        this.onChangeWithDependentField(valueControl, keyControl);
        // this.onChangeWithDependentField(LABEL_VALUE_NAME + this.labelCounter, LABEL_KEY_NAME + this.labelCounter);
        this.validateAllLabels();
    }

    deleteLabel(delIndex: number) {
        this.labelArray.removeAt(delIndex);
    }
    /**
         * @method onChangeWithDependentField
         * make the dependent field is required if the indepdent field is not empty.
         * @param fieldName is a independent field which determines if the dependent field is required.
         * @param dependentFieldName is dependent on the independent field.
         */
    onChangeWithDependentField(fieldControl: FormControl, dependentControl: FormControl) {
        fieldControl.valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(data => {
            if (data !== '') {
                if (!dependentControl.hasValidator(Validators.required)) {
                    dependentControl.addValidators(Validators.required);
                    dependentControl.markAsPending(); // validation will not be triggered until the field is touched.
                    dependentControl.setErrors({ required: true });
                }
            } else {
                dependentControl.removeValidators(Validators.required);
            }
            this.validateAllLabels(); // all the same label keys can show error message.

        })

    }

    validateAllLabels() {

        // The setTimeout wrapper ensures that validation logic will run after a new label field is added.
        setTimeout(_ => {
            for (const label of this.labelArray.controls) {
                const key = label.get('key');
                const val = label.get('value');
                if (key) {
                    key.updateValueAndValidity();
                }
                if (val) {
                    val.updateValueAndValidity();
                }
            }
        });
    }

    get labelArray() {
        return this.formGroup.get("labelList") as FormArray;
    }

    /**
     * Get the current value of MetadataField.CLUSTER_LABELS
     */
    get clusterLabelsValue() {

        let labelStr = '';
        const labelList = this.formGroup.value.labelList;

        labelList?.forEach(label => {
            const key = label.key;
            if (key) {
                labelStr += key + ', '
            }
        });
        return labelStr.slice(0, -2);
    }

    get labelTypeValue() {
        return this.formName === 'loadBalancerForm' ? 'Workload' : this.clusterTypeDescriptorTitleCase;
    }
}
