import { Component, OnInit } from '@angular/core';
import { takeUntil } from 'rxjs/operators';
import { APIClient } from 'src/app/swagger';
import { DockerDaemonStatus } from 'src/app/swagger/models';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import AppServices from "../../../../shared/service/appServices";
import { TanzuEvent, TanzuEventType } from "../../../../shared/service/Messenger";
import { NotificationTypes } from "../../../../shared/components/alert-notification/alert-notification.component";
import { DaemonStepMapping } from './daemon-validation-step.fieldmapping';

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
        this.registerDefaultFileImportedHandler(DaemonStepMapping);
    }

    ngOnInit(): void {
        super.ngOnInit();
        Broker.fieldMapUtilities.buildForm(this.formGroup, this.formName, DaemonStepMapping);
        this.htmlFieldLabels = Broker.fieldMapUtilities.getFieldLabelMap(DaemonStepMapping);
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
                FormMetaDataStore.saveMetaDataEntry(
                    this.formName,
                    'dockerDeamonValidation',
                    {
                        label: 'DOCKER DAEMON CONNECTED',
                        displayValue: 'yes'
                    });
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
