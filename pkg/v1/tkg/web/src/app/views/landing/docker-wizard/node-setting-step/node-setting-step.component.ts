import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { AppEdition } from 'src/app/shared/constants/branding.constants';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {

    constructor(private validationService: ValidationService) {
        super();
    }

    ngOnInit(): void {
        super.ngOnInit();
        this.formGroup.addControl('clusterName', new FormControl('', [this.validationService.isValidClusterName()]));

        if (this.edition !== AppEdition.TKG) {
            this.resurrectField('clusterName',
                [Validators.required, this.validationService.isValidClusterName()],
                this.formGroup.get('clusterName').value);
        }
    }
}
