import { Component, OnInit } from '@angular/core';
import { FormControl } from '@angular/forms';
import { takeUntil } from 'rxjs/operators';
import { APIClient } from 'src/app/swagger';
import { DockerDaemonStatus } from 'src/app/swagger/models';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import Broker from "../../../../shared/service/broker";
import { TkgEvent, TkgEventType } from "../../../../shared/service/Messenger";
import { NotificationTypes } from "../../../../shared/components/alert-notification/alert-notification.component";

@Component({
    selector: 'app-daemon-validation-step',
    templateUrl: './daemon-validation-step.component.html',
    styleUrls: ['./daemon-validation-step.component.scss']
})
export class DaemonValidationStepComponent extends StepFormDirective implements OnInit {

    connected: boolean = false;
    connecting: boolean = false;
    errorNotification: string = "";

    constructor(
        private validationService: ValidationService,
        private apiClient: APIClient
    ) {
        super();
    }

    ngOnInit(): void {
        super.ngOnInit();
        this.formGroup.addControl(
            'isConnected',
            new FormControl(
                false,
                this.validationService.isTrue
            )
        );
        Broker.messenger.getSubject(TkgEventType.CONFIG_FILE_IMPORTED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                this.configFileNotification = {
                    notificationType: NotificationTypes.SUCCESS,
                    message: data.payload
                };
                // The file import saves the data to local storage, so we reinitialize this step's form from there
                this.savedMetadata = FormMetaDataStore.getMetaData(this.formName);
                this.initFormWithSavedData();

                // Clear event so that listeners in other provider workflows do not receive false notifications
                Broker.messenger.clearEvent(TkgEventType.CONFIG_FILE_IMPORTED);
            });
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
                this.resurrectField('isConnected', this.validationService.isTrue(), 'true');
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
}
