import { Component, OnInit } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import Broker from "../../../../shared/service/broker";
import { TkgEvent, TkgEventType } from "../../../../shared/service/Messenger";
import { takeUntil } from "rxjs/operators";
import { FormMetaDataStore } from "../../wizard/shared/FormMetaDataStore";
import { NotificationTypes } from "../../../../shared/components/alert-notification/alert-notification.component";
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { DockerNodeSettingStepMapping, TkgDockerNodeSettingStepMapping } from './node-setting-step.fieldmapping';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    constructor(private validationService: ValidationService, private fieldMapUtilities: FieldMapUtilities) {
        super();
    }

    ngOnInit(): void {
        super.ngOnInit();
        const fieldMapping = this.edition === AppEdition.TKG ? TkgDockerNodeSettingStepMapping : DockerNodeSettingStepMapping;
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, fieldMapping);

        this.initFormWithSavedData();

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
