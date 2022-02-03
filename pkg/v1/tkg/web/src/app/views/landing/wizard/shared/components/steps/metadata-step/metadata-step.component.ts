// Angular imports
import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';

// App imports
import AppServices from '../../../../../../../shared/service/appServices';
import { FormUtils } from '../../../utils/form-utils';
import { MetadataField, MetadataStepMapping } from './metadata-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';
import { ValidationService } from '../../../validation/validation.service';

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

    constructor(private validationService: ValidationService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, MetadataStepMapping, null,
            this.getCustomRestorerMap());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(MetadataStepMapping);
        this.storeDefaultLabels(MetadataStepMapping);
        this.registerStepDescriptionTriggers({
            fields: [MetadataField.CLUSTER_LOCATION],
            clusterTypeDescriptor: true,
        })
        this.registerDefaultFileImportedHandler(this.eventFileImported, MetadataStepMapping, null,
            this.getCustomRestorerMap());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        // initialize label controls
        for (let [key, value] of this.labels) {
            this.addLabel(key, value);
        }
        this.addLabel();
    }

    // returns a map that associates the field 'clusterLabels' with a closure that restores our map of cluster labels
    private getCustomRestorerMap(): Map<string, (data: any) => void> {
        return new Map<string, (data: any) => void>([['clusterLabels', this.setLabels.bind(this)]]);
    }

    private setLabels(data: Map<string, string>)  {
        return this.labels = data;
    }

    private getCustomRetrievalMap(): Map<string, (key: any) => any> {
        return new Map<string, (data: any) => void>([[MetadataField.CLUSTER_LABELS, this.getClusterLabels.bind(this)]]);
    }

    private getClusterLabels(): Map<string, string> {
        return this.labels;
    }

    addLabel(key?: string, value?: string) {
        // SHIMON TODO: breaks test by not adding the this.labels
        this.labelCounter++;
        this.addLabelControls(this.labelCounter, key, value);
    }

    private addLabelControls(i: number, key, value: string) {
        const keyField = this.keyFieldName(this.labelCounter);
        const valueField = this.valueFieldName(this.labelCounter);
        this.keySet.add(keyField);
        FormUtils.addControl(
            this.formGroup,
            keyField,
            new FormControl(key || '', [
                this.validationService.isValidLabelOrAnnotation(),
                this.validationService.isUniqueLabel(
                    this.formGroup,
                    this.keySet,
                    keyField)
            ])
        );

        FormUtils.addControl(
            this.formGroup,
            valueField,
            new FormControl(value || '', [
                this.validationService.isValidLabelOrAnnotation()
            ])
        );
        // Label value depends on Label key. e.g.: if label key is not empty, then label value is required
        this.onChangeWithDependentField(keyField, valueField);
        // Label key depends on Label value. e.g.: if label value is not empty, then label key is required
        this.onChangeWithDependentField(valueField, keyField);
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
            for (let i = 1; i <= this.labelCounter; i++) {
                const keyFieldName = this.keyFieldName(i);
                const keyControl = this.formGroup.get(keyFieldName);
                const valueFieldName = this.valueFieldName(i);
                const valueControl = this.formGroup.get(valueFieldName);
                if (keyControl) {
                    if (this.savedKeySet.has(keyFieldName)) {
                        keyControl.markAsTouched();
                    }
                    keyControl.updateValueAndValidity();
                }
                if (valueControl) {
                    valueControl.updateValueAndValidity();
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
        this.storeUserDataFromMapping(MetadataStepMapping, this.getCustomRetrievalMap());
        this.storeDefaultDisplayOrder(MetadataStepMapping);
    }

    // should be a 1-based index
    keyFieldName(i: number): string {
        return LABEL_KEY_NAME + i;
    }

    // should be a 1-based index
    valueFieldName(i: number): string {
        return LABEL_VALUE_NAME + i;
    }
}
