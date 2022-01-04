// Angular imports
import { Component, OnInit } from '@angular/core';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
// Third party imports
import { takeUntil } from "rxjs/operators";
// App imports
import AppServices from "../../../../shared/service/appServices";
import { DockerNodeSettingStepMapping } from './node-setting-step.fieldmapping';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { FormMetaDataStore } from "../../wizard/shared/FormMetaDataStore";
import { NotificationTypes } from "../../../../shared/components/alert-notification/alert-notification.component";
import { TkgEvent, TkgEventType } from "../../../../shared/service/Messenger";

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    constructor(private fieldMapUtilities: FieldMapUtilities) {
        super();
    }

    private customizeForm() {
        AppServices.messenger.getSubject(TkgEventType.CONFIG_FILE_IMPORTED)
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
                AppServices.messenger.clearEvent(TkgEventType.CONFIG_FILE_IMPORTED);
            });
    }

    ngOnInit(): void {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, DockerNodeSettingStepMapping);
        this.customizeForm();
        this.initFormWithSavedData();
    }
}
