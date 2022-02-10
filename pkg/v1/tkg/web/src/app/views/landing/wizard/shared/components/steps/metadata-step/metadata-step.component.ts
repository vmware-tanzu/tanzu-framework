// Angular imports
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';

// App imports
import AppServices from '../../../../../../../shared/service/appServices';
import { FormUtils } from '../../../utils/form-utils';
import { MetadataField, MetadataStepMapping } from './metadata-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { ValidationService } from '../../../validation/validation.service';
import { StepMapping } from '../../../field-mapping/FieldMapping';

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
    private stepMapping: StepMapping;

    constructor(private validationService: ValidationService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.registerStepDescriptionTriggers({
            fields: [MetadataField.CLUSTER_LOCATION],
            clusterTypeDescriptor: true,
        })
        this.registerDefaultFileImportedHandler(this.eventFileImported, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        // initialize label controls
        if (this.labels.size === 0) {
            this.addLabel();
        }
    }

    private supplyStepMapping(): StepMapping {
        if (!this.stepMapping) {
            this.stepMapping = this.createStepMapping();
        }
        return this.stepMapping;
    }

    private createStepMapping() {
        const result = MetadataStepMapping;
        const clusterFieldMapping = AppServices.fieldMapUtilities.getFieldMapping(MetadataField.CLUSTER_LABELS, result);
        clusterFieldMapping.retriever = this.getClusterLabels.bind(this);
        clusterFieldMapping.restorer = this.setClusterLabels.bind(this);
        return result;
    }

    // TODO: the 'labels' field now holds a keyField => valueField mapping, so when receiving the data, we build new controls to hold data
    private setClusterLabels(data: Map<string, string>)  {
        this.clearLabels();
        // ADD new ones
        for (const [key, value] of data) {
            this.addLabel(key, value);
        }
        // ensure at least one field
        if (this.labels.size === 0) {
            this.addLabel();
        }
    }

    private clearLabels() {
        // REMOVE existing label fields
        for (const [keyField, valueField] of this.labels) {
            this.formGroup.removeControl(keyField);
            this.formGroup.removeControl(valueField);
        }
        this.labels = new Map<string, string>();
        this.keySet = new Set();
        this.labelCounter = 0;
    }

    // TODO: the 'labels' field holds a keyField => valueField mapping, so when returning the data, we build a new map from field data
    // TODO: public for testing only
    getClusterLabels(): Map<string, string> {
        const result = new Map<string, string>();
        for (const [keyField, valueField] of this.labels) {
            const key = this.formGroup.get(keyField).value;
            const val = this.formGroup.get(valueField).value;
            if (key && val) {
                result.set(key, val);
            }
        }
        return result;
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
    }

    dynamicDescription(): string {
        const clusterLocation = this.getFieldValue(MetadataField.CLUSTER_LOCATION, true);
        return clusterLocation ? 'Location: ' + clusterLocation : 'Specify metadata for the ' + this.clusterTypeDescriptor + ' cluster';
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDefaultDisplayOrder(this.supplyStepMapping());
    }
}
