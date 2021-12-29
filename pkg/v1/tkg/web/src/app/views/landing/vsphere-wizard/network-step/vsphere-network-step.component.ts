// App imports
import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';
import { WizardForm } from '../../wizard/shared/constants/wizard.constants';

export class VsphereNetworkStepComponent extends SharedNetworkStepComponent {
    static readonly description = 'Specify how Tanzu Kubernetes Grid networking is provided and any global network settings';

    protected supplyFieldsAffectingStepDescription(): string[] {
        return [WizardForm.NETWORK];
    }

    dynamicDescription(): string {
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        // because vSphere allows for the user to select a network name
        const networkName = this.getFieldValue(WizardForm.NETWORK);
        if (networkName) {
            return 'Network: ' + networkName;
        }
        return VsphereNetworkStepComponent.description;
    }
}
