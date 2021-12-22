// Angular modules
import { Component, OnInit } from '@angular/core';
import { FormControl } from '@angular/forms';

// App imports
import { StepFormDirective } from '../../../step-form/step-form';
import { FormUtils } from '../../../utils/form-utils';

@Component({
    selector: 'app-shared-ceip-step',
    templateUrl: './ceip-step.component.html',
    styleUrls: ['./ceip-step.component.scss']
})
export class SharedCeipStepComponent extends StepFormDirective implements OnInit {
    ngOnInit() {
        super.ngOnInit();
        FormUtils.addControl(
            this.formGroup,
            'ceipOptIn',
            new FormControl(true, [])
        );
        this.initFormWithSavedData();
    }
}
