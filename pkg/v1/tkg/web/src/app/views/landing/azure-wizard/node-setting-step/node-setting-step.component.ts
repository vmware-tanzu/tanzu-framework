// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import AppServices from '../../../../shared/service/appServices';
import { AzureField } from '../azure-wizard.constants';
import { AzureInstanceType } from 'src/app/swagger/models';
import { NodeSettingStepDirective } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.component';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends NodeSettingStepDirective<AzureInstanceType> implements OnInit {
    constructor(protected validationService: ValidationService) {
        super(validationService);
    }

    protected getKeyFromNodeInstance(nodeInstance: AzureInstanceType): string {
        return nodeInstance.name;
    }

    protected getDisplayFromNodeInstance(nodeInstance: AzureInstanceType): string {
        return nodeInstance.name;
    }

    protected subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe(this,
            TanzuEventType.AZURE_GET_INSTANCE_TYPES, this.onFetchedInstanceTypes.bind(this))
    }

    private onFetchedInstanceTypes(instanceTypes: AzureInstanceType[]) {
        this.nodeTypes = instanceTypes.sort();
        if (!this.modeClusterStandalone && this.nodeTypes.length === 1) {
            this.formGroup.get(AzureField.NODESETTING_WORKERTYPE).setValue(this.nodeTypes[0].name);
        }
    }
}
