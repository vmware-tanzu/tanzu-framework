// Angular imports
import { Component, OnInit } from '@angular/core';

// App imports
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { MetadataField, MetadataStepMapping } from './metadata-step.fieldmapping';
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
            fields: [MetadataField.CLUSTER_LOCATION],
            clusterTypeDescriptor: true,
        })

        this.initFormWithSavedData();
        }

    initFormWithSavedData() {
        const savedLabelsString = this.getSavedValue(MetadataField.CLUSTER_LABELS, '');
        if (savedLabelsString !== '') {
            const savedLabelsArray = savedLabelsString.split(', ')
            savedLabelsArray.map(label => {
                const labelArray = label.split(':');
                this.labels.set(labelArray[0], labelArray[1]);
                this.savedKeySet.add(LABEL_KEY_NAME + this.labelCounter);
            });
        }
        super.initFormWithSavedData();
    }

    dynamicDescription(): string {
        const clusterLocation = this.getFieldValue(MetadataField.CLUSTER_LOCATION, true);
        return clusterLocation ? 'Location: ' + clusterLocation : 'Specify metadata for the ' + this.clusterTypeDescriptor + ' cluster';
    }
}
