import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereNetworkFieldMappings } from './vsphere-network-step.fieldmapping';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereWizardFormService } from '../../../../shared/service/vsphere-wizard-form.service';
import { Component } from '@angular/core';
import { VsphereField } from '../vsphere-wizard.constants';

@Component({
    selector: 'vsphere-network-step',
    templateUrl: '../../wizard/shared/components/steps/network-step/network-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/network-step/network-step.component.scss']
})
export class VsphereNetworkStepComponent extends SharedNetworkStepComponent {
    static readonly description = 'Specify how Tanzu Kubernetes Grid networking is provided and any global network settings';

    constructor(protected validationService: ValidationService,
                protected wizardFormService: VSphereWizardFormService) {
        super(validationService, wizardFormService);
    }

    protected supplyFieldsAffectingStepDescription(): string[] {
        return [VsphereField.NETWORK_NAME];
    }

    dynamicDescription(): string {
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        // because vSphere allows for the user to select a network name
        const networkName = this.getFieldValue(VsphereField.NETWORK_NAME);
        if (networkName) {
            return 'Network: ' + networkName;
        }
        return VsphereNetworkStepComponent.description;
    }

    protected supplyEnablesNetworkName(): boolean {
        return true;
    }

    protected supplyEnablesNoProxyWarning(): boolean {
        return true;
    }

    protected supplyStepMapping(): StepMapping {
        // we want to prepend the special vsphere fieldMappings
        const basicStepMapping = super.supplyStepMapping().fieldMappings;
        const vSphereMappings = VsphereNetworkFieldMappings;
        vSphereMappings.push(...basicStepMapping);
        return {
            fieldMappings: vSphereMappings
        }
    }
}
