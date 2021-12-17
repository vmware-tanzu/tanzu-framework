/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';

/**
 * App imports
 */
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { FormUtils } from '../../wizard/shared/utils/form-utils';

@Component({
    selector: 'app-worker-node-setting-step',
    templateUrl: './worker-node-setting-step.component.html',
    styleUrls: ['./worker-node-setting-step.component.scss']
})
export class WorkerNodeSettingStepComponent extends StepFormDirective implements OnInit {
    currentRegion = "US-WEST";
    workderNodeInstanceTypes = ["large", "medium", "small"];
    azs = ["US-WEST", "US-EAST"];

    buildForm() {
        FormUtils.addControl(
            this.formGroup,
            'workerNodeInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );

        ['az1', 'az2', 'az3'].forEach(id =>
            FormUtils.addControl(
                this.formGroup,
                id,
                new FormControl('', [
                    Validators.required
                ])
            )
        )
    }

    ngOnInit() {
        super.ngOnInit();
        this.buildForm();
        this.initFormWithSavedData();
    }
}
