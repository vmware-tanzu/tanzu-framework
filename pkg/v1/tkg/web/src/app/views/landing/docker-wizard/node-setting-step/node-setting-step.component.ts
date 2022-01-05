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
import { TanzuEvent, TanzuEventType } from "../../../../shared/service/Messenger";

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    clusterNameInstruction: string;

    private customizeForm() {
        this.registerDefaultFileImportedHandler(this.supplyStepMapping());
    }

    private supplyStepMapping() {
        const mapping = DockerNodeSettingStepMapping;
        // dynamically modify the cluster name label based on the type descriptor and whether the cluster name is required
        const clusterNameMapping = AppServices.fieldMapUtilities.getFieldMapping('clusterName', mapping);
        let clusterNameLabel = this.clusterTypeDescriptor.toUpperCase() + ' CLUSTER NAME';
        if (!AppServices.appDataService.isClusterNameRequired()) {
            clusterNameLabel += ' (OPTIONAL)';
        }
        clusterNameMapping.label = clusterNameLabel;
        return mapping;
    }

    ngOnInit(): void {
        super.ngOnInit();
        AppServices.fieldMapUtilities.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.registerDefaultFileImportedHandler(this.supplyStepMapping());

        if (AppServices.appDataService.isClusterNameRequired()) {
            this.clusterNameInstruction = 'Specify a name for the ' + this.clusterTypeDescriptor + ' cluster.';
        } else {
            this.clusterNameInstruction = 'Optionally specify a name for the ' + this.clusterTypeDescriptor + ' cluster. ' +
                'If left blank, the installer names the cluster automatically.';
        }
        this.customizeForm();
        this.initFormWithSavedData();
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(DockerNodeSettingStepMapping);
        this.storeDefaultDisplayOrder(DockerNodeSettingStepMapping);
    }
}
