// Angular modules
import { Component, OnInit } from '@angular/core';
import { Validators, FormControl } from '@angular/forms';

// Third party imports
import { takeUntil } from 'rxjs/operators';

// App imports
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { TkgEvent, TkgEventType } from '../../../../shared/service/Messenger';
import { FormMetaDataStore } from '../../wizard/shared/FormMetaDataStore';
import { NotificationTypes } from '../../../../shared/components/alert-notification/alert-notification.component';
import { FormUtils } from '../../wizard/shared/utils/form-utils';
import Broker from '../../../../shared/service/broker';

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
        FormUtils.addControl(
            this.formGroup,
            'clusterName',
            new FormControl('', [this.validationService.isValidClusterName()])
        );
        this.initFormWithSavedData();

        if (this.edition !== AppEdition.TKG) {
            this.resurrectField('clusterName',
                [Validators.required, this.validationService.isValidClusterName()],
                this.formGroup.get('clusterName').value);
        }
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
    }
}
