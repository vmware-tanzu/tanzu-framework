/**
 * Angular Modules
 */
import { Component, OnInit, Input } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';

import { StepFormDirective } from '../shared/step-form/step-form';
import { FieldMapUtilities } from '../shared/field-mapping/FieldMapUtilities';
import { StepMapping } from '../shared/field-mapping/FieldMapping';

@Component({
    selector: 'app-storage-step',
    templateUrl: './storage-step.component.html',
    styleUrls: ['./storage-step.component.scss']
})
export class StorageStepComponent extends StepFormDirective implements OnInit {

    constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    protected supplyStepMapping(): StepMapping {
        return undefined;
    }

    ngOnInit() {
        super.ngOnInit();
    }
}
