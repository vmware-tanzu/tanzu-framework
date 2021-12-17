/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    FormControl
} from '@angular/forms';

import { APIClient } from '../../../../swagger/api-client.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { FormUtils } from '../../wizard/shared/utils/form-utils';

@Component({
    selector: 'app-ami-step',
    templateUrl: './ami-step.component.html',
    styleUrls: ['./ami-step.component.scss']
})
export class AmiStepComponent extends StepFormDirective implements OnInit {

    constructor(private apiClient: APIClient) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();
        FormUtils.addControl(
            this.formGroup,
            'amiOrgId',
            new FormControl('', [])
        );
        this.initFormWithSavedData();
    }
}
