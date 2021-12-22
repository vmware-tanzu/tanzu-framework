/**
 * Angular Modules
 */
import { Component, OnInit } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { FormControl } from '@angular/forms';

/**
 * App imports
 */
import { StepFormDirective } from '../../../step-form/step-form';
import { FormUtils } from '../../../utils/form-utils';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { StepMapping } from '../../../field-mapping/FieldMapping';

@Component({
    selector: 'app-shared-register-tmc-step',
    templateUrl: './register-tmc-step.component.html',
    styleUrls: ['./register-tmc-step.component.scss']
})
export class SharedRegisterTmcStepComponent extends StepFormDirective implements OnInit {

    configContent: any;
    emptyUrl: boolean = true;

    constructor(private http: HttpClient, protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    protected supplyStepMapping(): StepMapping {
        return undefined;
    }

    ngOnInit() {
        super.ngOnInit();
        FormUtils.addControl(
            this.formGroup,
            'tmcRegUrl',
            new FormControl('', [])
        );

        this.formGroup.get('tmcRegUrl').valueChanges.subscribe(data => {
            if (data) { this.emptyUrl = false; }
        });
    }

    // TODO: need to validate the registration url prior to making http call
    /**
     * @method GetRemoteConfig
     * makes an http GET request to the provided registration URL; loads JSON/YAML
     * into the ngx codemirror editor window
     * @returns {any|Subscription}
     */
    getRemoteConfig() {
        this.errorNotification = '';

        return this.http.get<string>('/api/integration/tmc', {
            headers: { 'Content-Type': 'application/yaml; charset=utf-8' },
            params: {'url': encodeURIComponent(this.formGroup.controls['tmcRegUrl'].value) },
            responseType: "json"
        }).subscribe(
            (data) => {
                this.configContent = atob(data);
            },
            err => {
                this.errorNotification = `Unable to retrieve Tanzu Mission Control registration data. ${err}`;
            }
        )
    }

    resetConfigContent() {
        this.configContent = '';
    }
}
