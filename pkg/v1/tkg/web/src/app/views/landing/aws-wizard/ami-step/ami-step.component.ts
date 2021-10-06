/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { takeUntil } from 'rxjs/operators';

import { APIClient } from '../../../../swagger/api-client.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { AppDataService } from 'src/app/shared/service/app-data.service';

@Component({
    selector: 'app-ami-step',
    templateUrl: './ami-step.component.html',
    styleUrls: ['./ami-step.component.scss']
})
export class AmiStepComponent extends StepFormDirective implements OnInit {

    constructor(private apiClient: APIClient, appDataService: AppDataService) {
        super(appDataService);
    }

    ngOnInit() {
        super.ngOnInit();
        this.formGroup.addControl(
            'amiOrgId',
            new FormControl('', [])
        );
    }
}
