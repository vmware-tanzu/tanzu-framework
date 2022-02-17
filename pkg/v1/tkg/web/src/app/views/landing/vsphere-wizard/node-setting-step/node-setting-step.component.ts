// Angular imports
import { Component, Input, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// App imports
import AppServices from 'src/app/shared/service/appServices';
import { NodeType } from '../../wizard/shared/constants/wizard.constants';
import { IpFamilyEnum } from '../../../../shared/constants/app.constants';
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER, VsphereNodeSettingFieldMappings } from './node-setting-step.fieldmapping';
import { NodeSettingStepDirective } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.component';
import { NodeSettingField } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.fieldmapping';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VsphereField, VsphereNodeTypes } from '../vsphere-wizard.constants';

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
        const result = AppServices.fieldMapUtilities.cloneStepMapping(super.createStepMapping());
        // We take the inherited field mappings and insert our specific vSphere field mappings
        AppServices.fieldMapUtilities.insertFieldMappingsAfter(result, NodeSettingField.WORKER_NODE_INSTANCE_TYPE,
            VsphereNodeSettingFieldMappings);
        return result;
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
