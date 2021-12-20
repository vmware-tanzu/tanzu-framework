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

@Component({
    selector: 'app-shared-ceip-step',
    templateUrl: './ceip-step.component.html',
    styleUrls: ['./ceip-step.component.scss']
})
export class SharedCeipStepComponent extends StepFormDirective implements OnInit {
    constructor(private fieldMapUtilities: FieldMapUtilities) {
        super();
    }
    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, CeipStepMapping);

        this.initFormWithSavedData();
    }
}
