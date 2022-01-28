// Angular imports
import { Component, OnInit } from '@angular/core';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
// App imports
import AppServices from "../../../../shared/service/appServices";
import { DockerNodeSettingStepMapping } from './node-setting-step.fieldmapping';
import { TanzuEventType } from "../../../../shared/service/Messenger";

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    clusterNameInstruction: string;

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
        this.registerDefaultFileImportedHandler(this.eventFileImported, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        if (AppServices.appDataService.isClusterNameRequired()) {
            this.clusterNameInstruction = 'Specify a name for the ' + this.clusterTypeDescriptor + ' cluster.';
        } else {
            this.clusterNameInstruction = 'Optionally specify a name for the ' + this.clusterTypeDescriptor + ' cluster. ' +
                'If left blank, the installer names the cluster automatically.';
        }
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(DockerNodeSettingStepMapping);
        this.storeDefaultDisplayOrder(DockerNodeSettingStepMapping);
    }
}
