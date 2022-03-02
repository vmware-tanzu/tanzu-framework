// Angular imports
import { Component, OnInit } from '@angular/core';
// Third party imports
import { takeUntil } from 'rxjs/operators';
// App imports
import { APIClient } from 'src/app/swagger';
import AppServices from "../../../../shared/service/appServices";
import { DaemonStepMapping } from './daemon-validation-step.fieldmapping';
import { DockerDaemonStatus } from 'src/app/swagger/models';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

@Component({
    selector: 'app-daemon-validation-step',
    templateUrl: './daemon-validation-step.component.html',
    styleUrls: ['./daemon-validation-step.component.scss']
})
export class DaemonValidationStepComponent extends StepFormDirective implements OnInit {

    connected: boolean = false;
    connecting: boolean = false;
    errorNotification: string = "";

    constructor(private validationService: ValidationService,
                private apiClient: APIClient) {
        super();
    }

    private customizeForm() {
        this.registerDefaultFileImportedHandler(this.eventFileImported, DaemonStepMapping);
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);
    }

    ngOnInit(): void {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, DaemonStepMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(DaemonStepMapping);
        this.storeDefaultLabels(DaemonStepMapping);

        this.customizeForm();
        this.connectToDocker();
    }

    initFormWithSavedData() {
        // We don't want to set the isConnected field from saved data, so we override the method's default implementation
    }

    getFormName() {
        super.getFormName();
    }

    connectToDocker() {
        this.connecting = true;
        this.apiClient.checkIfDockerDaemonAvailable()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: DockerDaemonStatus) => {
                this.connected = !!data.status;
                this.connecting = false;
                this.resurrectField('isConnected', this.validationService.isTrue(), 'true', { emitEvent: false });
            }, (err) => {
                this.connected = false;
                this.connecting = false;
                this.errorNotification = err.error.message;
            });
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(DaemonStepMapping);
        this.storeDefaultDisplayOrder(DaemonStepMapping);
    }
}
