// App imports
import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';
import { NetworkField } from '../../wizard/shared/components/steps/network-step/network-step.fieldmapping';
import Broker from '../../../../shared/service/broker';

export class DockerNetworkStepComponent extends SharedNetworkStepComponent {
    static readonly description = 'Specify the cluster Pod CIDR';

    protected supplyFieldsAffectingStepDescription(): string[] {
        return [NetworkField.CLUSTER_POD_CIDR];
    }

    dynamicDescription(): string {
        if (this.getFieldValue(NetworkField.CLUSTER_POD_CIDR)) {
            return 'Cluster Pod CIDR: ' + this.getFieldValue(NetworkField.CLUSTER_POD_CIDR);
        }
        return DockerNetworkStepComponent.description;
    }

    protected storeUserData() {
        const identifier = this.createUserDataIdentifier('clusterPodCidr');
        Broker.userDataService.storeInputField(identifier, this.formGroup);
        this.storeDisplayOrder(['clusterPodCidr']);
    }
}
