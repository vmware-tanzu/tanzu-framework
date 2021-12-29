// App imports
import { KUBE_VIP, SharedLoadBalancerStepComponent } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';

export class VsphereLoadBalancerStepComponent extends SharedLoadBalancerStepComponent {
    static readonly description = 'Specify VMware NSX Advanced Load Balancer settings';

    protected dynamicDescription(): string {
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        const controllerHost = this.getFieldValue( 'controllerHost');
        if (controllerHost) {
            return 'Controller: ' + controllerHost;
        }
        const endpointProvider = this.getFieldValue( "controlPlaneEndpointProvider");
        if (endpointProvider === KUBE_VIP) {
            return 'Optionally specify VMware NSX Advanced Load Balancer settings';
        }
        return VsphereLoadBalancerStepComponent.description;
    }
}
