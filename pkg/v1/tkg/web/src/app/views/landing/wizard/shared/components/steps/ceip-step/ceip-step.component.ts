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

@Component({
    selector: 'app-shared-ceip-step',
    templateUrl: './ceip-step.component.html',
    styleUrls: ['./ceip-step.component.scss']
})
export class SharedCeipStepComponent extends StepFormDirective implements OnInit {
    ngOnInit() {
        super.ngOnInit();
        this.formGroup.addControl(
            'ceipOptIn',
            new FormControl(true, [])
        );
        this.initFormWithSavedData();
    }
}
