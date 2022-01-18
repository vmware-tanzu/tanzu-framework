// App imports
import { SharedNetworkStepComponent } from '../../wizard/shared/components/steps/network-step/network-step.component';

export class DockerNetworkStepComponent extends SharedNetworkStepComponent {
    static readonly description = 'Specify the cluster Pod CIDR';

    protected supplyFieldsAffectingStepDescription(): string[] {
        return ['clusterPodCidr'];
    }

    dynamicDescription(): string {
        if (this.getFieldValue('clusterPodCidr')) {
            return 'Cluster Pod CIDR: ' + this.getFieldValue('clusterPodCidr');
        }
        return DockerNetworkStepComponent.description;
    }
}
