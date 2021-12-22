/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { FormControl } from '@angular/forms';
import { FormMetaDataStore, FormMetaData } from '../../../FormMetaDataStore';

/**
 * App imports
 */
import { StepFormDirective } from '../../../step-form/step-form';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { CeipStepMapping } from './ceip-step.fieldmapping';
import { StepMapping } from '../../../field-mapping/FieldMapping';

@Component({
    selector: 'app-shared-ceip-step',
    templateUrl: './ceip-step.component.html',
    styleUrls: ['./ceip-step.component.scss']
})
export class SharedCeipStepComponent extends StepFormDirective implements OnInit {
    constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    protected supplyStepMapping(): StepMapping {
        return CeipStepMapping;
    }

    ngOnInit() {
        super.ngOnInit();

        this.initFormWithSavedData();
    }
}
