// Angular imports
import { Component, OnInit } from '@angular/core';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
// App imports
import AppServices from "../../../../shared/service/appServices";
import { DockerNodeSettingStepMapping } from './node-setting-step.fieldmapping';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    clusterNameInstruction: string;

    private supplyStepMapping() {
        return DockerNodeSettingStepMapping;
    }

    ngOnInit(): void {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.registerDefaultFileImportedHandler(this.eventFileImported, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        this.clusterNameInstruction = 'Specify a name for the ' + this.clusterTypeDescriptor + ' cluster.';
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(DockerNodeSettingStepMapping);
        this.storeDefaultDisplayOrder(DockerNodeSettingStepMapping);
    }
}
