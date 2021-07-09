/**
 * Angular Modules
 */
import { Component, OnInit, Input } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';

import { StepFormDirective } from '../shared/step-form/step-form';

@Component({
    selector: 'app-storage-step',
    templateUrl: './storage-step.component.html',
    styleUrls: ['./storage-step.component.scss']
})
export class StorageStepComponent extends StepFormDirective implements OnInit {

    constructor() {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
    }
}
