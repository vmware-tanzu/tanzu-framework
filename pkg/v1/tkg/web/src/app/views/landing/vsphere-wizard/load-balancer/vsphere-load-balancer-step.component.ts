// Angular imports
import { Validators } from '@angular/forms';
import { Component, OnInit } from '@angular/core';
// App imports
import AppServices from '../../../../shared/service/appServices';
import {
    KUBE_VIP,
    NSX_ADVANCED_LOAD_BALANCER,
    SharedLoadBalancerStepComponent
} from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import { LoadBalancerField } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.fieldmapping';
import { TanzuEventType } from '../../../../shared/service/Messenger';

const HA_REQUIRED_FIELDS = [
    LoadBalancerField.CONTROLLER_HOST,
    LoadBalancerField.USERNAME,
    LoadBalancerField.PASSWORD,
    LoadBalancerField.CONTROLLER_CERT,
    LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_NAME,
    LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_CIDR
]

@Component({
    selector: 'app-vsphere-load-balancer-step',
    templateUrl: '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/load-balancer/load-balancer-step.component.scss']
})
export class VsphereLoadBalancerStepComponent extends SharedLoadBalancerStepComponent implements OnInit {
    static readonly description = 'Specify VMware NSX Advanced Load Balancer settings';
    static readonly descriptionKubeVip = 'Optionally specify VMware NSX Advanced Load Balancer settings';

    currentControlPlaneEndpointProvider: string;

    ngOnInit() {
        super.ngOnInit();
        this.setLoadBalancerLabel(this.currentControlPlaneEndpointProvider);
        this.registerStepDescriptionTriggers({fields: [LoadBalancerField.CONTROLLER_HOST]});
    }

    private setLoadBalancerLabel(controlPlaneEndpointProvider: string) {
        let prependLabel = '';
        if (controlPlaneEndpointProvider === NSX_ADVANCED_LOAD_BALANCER) {
            prependLabel = 'Manual ';
        } else if (controlPlaneEndpointProvider === KUBE_VIP) {
            prependLabel = 'Optional ';
        }
        this.loadBalancerLabel = prependLabel + 'VMware NSX Advanced Load Balancer Settings';
    }

    protected customizeForm() {
        super.customizeForm();
        AppServices.messenger.subscribe<string>(TanzuEventType.VSPHERE_CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED, ({ payload }) => {
                this.onControlPlaneEndpointProviderChange(payload);
            });
    }

    private onControlPlaneEndpointProviderChange(newProvider: string) {
        this.currentControlPlaneEndpointProvider = newProvider;
        if (this.currentControlPlaneEndpointProvider === NSX_ADVANCED_LOAD_BALANCER) {
            HA_REQUIRED_FIELDS.forEach(fieldName => this.resurrectField(fieldName, [Validators.required]));
        } else {
            HA_REQUIRED_FIELDS.forEach(fieldName => this.disarmField(fieldName, true));
        }
        this.errorNotification = '';
        this.setLoadBalancerLabel(newProvider);
        this.triggerStepDescriptionChange();
    }

    dynamicDescription(): string {
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        const controllerHost = this.getFieldValue( LoadBalancerField.CONTROLLER_HOST);
        if (controllerHost) {
            return 'Controller: ' + controllerHost;
        }

        if (this.currentControlPlaneEndpointProvider === KUBE_VIP) {
            return VsphereLoadBalancerStepComponent.descriptionKubeVip;
        }
        return VsphereLoadBalancerStepComponent.description;
    }
}
