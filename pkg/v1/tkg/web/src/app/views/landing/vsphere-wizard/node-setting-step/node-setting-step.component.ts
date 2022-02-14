// Angular imports
import { Component, Input, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// App imports
import AppServices from 'src/app/shared/service/appServices';
import { NodeType } from '../../wizard/shared/constants/wizard.constants';
import { IpFamilyEnum, PROVIDERS, Providers } from '../../../../shared/constants/app.constants';
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VsphereField, VsphereNodeTypes } from '../vsphere-wizard.constants';
import { VsphereNodeSettingFieldMappings } from './node-setting-step.fieldmapping';
import { NodeSettingStepDirective } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.component';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends NodeSettingStepDirective<NodeType> implements OnInit {
    controlPlaneEndpointProviders = [KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER];
    currentControlPlaneEndpointProvider = KUBE_VIP;
    controlPlaneEndpointOptional = '';

    constructor(protected validationService: ValidationService) {
        super(validationService);
        this.nodeTypes = [...VsphereNodeTypes];
    }

    protected createStepMapping(): StepMapping {
        const commonStepMapping = super.createStepMapping();
        // We take the inherited field mappings and add our specific vSphere field mappings
        return { fieldMappings: [...commonStepMapping.fieldMappings, ...VsphereNodeSettingFieldMappings] };
    }

    protected listenToEvents() {
        super.listenToEvents();
        this.registerOnValueChange(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER,
            this.onControlPlaneEndpointProviderChange.bind(this));
        this.registerOnIpFamilyChange(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP,
            [ Validators.required, this.validationService.isValidIpOrFqdn() ],
            [ Validators.required, this.validationService.isValidIpv6OrFqdn() ]);
    }

    onControlPlaneEndpointProviderChange(provider: string): void {
        this.currentControlPlaneEndpointProvider = provider;
        AppServices.messenger.publish({
            type: TanzuEventType.VSPHERE_CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
            payload: provider
        });
        this.resurrectField(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, (provider === KUBE_VIP) ? [
            Validators.required,
            this.ipFamily === IpFamilyEnum.IPv4 ? this.validationService.isValidIpOrFqdn() : this.validationService.isValidIpv6OrFqdn()
        ] : [
            this.ipFamily === IpFamilyEnum.IPv4 ? this.validationService.isValidIpOrFqdn() : this.validationService.isValidIpv6OrFqdn()
        ], this.getStoredValue(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, this.supplyStepMapping(), ''));

        this.controlPlaneEndpointOptional = (provider === KUBE_VIP ? '' : '(OPTIONAL)');
    }

    protected getKeyFromNodeInstance(nodeInstance: NodeType): string {
        return nodeInstance.id;
    }

    protected getDisplayFromNodeInstance(nodeInstance: NodeType): string {
        return nodeInstance.name;
    }
}
