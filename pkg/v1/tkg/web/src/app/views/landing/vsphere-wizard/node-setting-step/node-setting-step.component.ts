import { TkgEventType } from 'src/app/shared/service/Messenger';
/**
 * Angular Modules
 */
import { Component, OnInit, Input } from '@angular/core';
import {
    Validators,
    FormControl
} from '@angular/forms';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';

/**
 * App imports
 */
import { IpFamilyEnum, PROVIDERS, Providers } from '../../../../shared/constants/app.constants';
import { NodeType} from '../../wizard/shared/constants/wizard.constants';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import Broker from 'src/app/shared/service/broker';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { VsphereField, vSphereNodeTypes } from "../vsphere-wizard.constants";

enum InstanceType {
    DEV = 'dev',
    PROD = 'prod'
};

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    @Input() providerType: string;

    nodeTypes: Array<NodeType> = [];
    PROVIDERS: Providers = PROVIDERS;
    vSphereNodeTypes: Array<NodeType> = vSphereNodeTypes;
    nodeType: string;

    displayForm = false;

    controlPlaneEndpointProviders = [KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER];
    currentControlPlaneEndpoingProvider = KUBE_VIP;
    controlPlaneEndpointOptional = "";

    constructor(private validationService: ValidationService,
        private wizardFormService: VSphereWizardFormService) {

        super();
        this.nodeTypes = [...vSphereNodeTypes];
    }

    ngOnInit() {
        super.ngOnInit();
        this.formGroup.addControl(
            VsphereField.NODESETTING_CONTROL_PLANE_SETTING,
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            VsphereField.NODESETTING_INSTANCE_TYPE_DEV,
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            VsphereField.NODESETTING_INSTANCE_TYPE_PROD,
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED,
            new FormControl(true, [])
        );
        if (!this.modeClusterStandalone) {
            this.formGroup.addControl(
                VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE,
                new FormControl('', [
                    Validators.required
                ])
            );
        }
        this.formGroup.addControl(
            VsphereField.NODESETTING_CLUSTER_NAME,
            new FormControl('', [
                this.validationService.isValidClusterName()
            ])
        );
        this.formGroup.addControl(
            VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP,
            new FormControl('', [
                Validators.required,
                this.validationService.isValidIpOrFqdn()
            ])
        );

        this.formGroup.addControl(
            VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER,
            new FormControl(this.currentControlPlaneEndpoingProvider, [
                Validators.required
            ])
        );

        this.registerOnValueChange(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER,
            this.onControlPlaneEndpoingProviderChange.bind(this));
        this.registerOnIpFamilyChange(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, [
            Validators.required,
            this.validationService.isValidIpOrFqdn()
        ], [
            Validators.required,
            this.validationService.isValidIpv6OrFqdn()
        ]);

        setTimeout(_ => {
            this.displayForm = true;
            this.formGroup.get(VsphereField.NODESETTING_CONTROL_PLANE_SETTING).valueChanges.subscribe(data => {
                if (data === InstanceType.DEV) {
                    this.nodeType = InstanceType.DEV;
                    this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_DEV).setValidators([
                        Validators.required
                    ]);
                    this.formGroup.controls[VsphereField.NODESETTING_INSTANCE_TYPE_PROD].clearValidators();
                    this.formGroup.controls[VsphereField.NODESETTING_INSTANCE_TYPE_PROD].setValue('');
                } else if (data === InstanceType.PROD) {
                    this.nodeType = InstanceType.PROD;
                    this.formGroup.controls[VsphereField.NODESETTING_INSTANCE_TYPE_PROD].setValidators([
                        Validators.required
                    ]);
                    this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_DEV).clearValidators();
                    this.formGroup.controls[VsphereField.NODESETTING_INSTANCE_TYPE_DEV].setValue('');
                }
                this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_DEV).updateValueAndValidity();
                this.formGroup.controls[VsphereField.NODESETTING_INSTANCE_TYPE_PROD].updateValueAndValidity();
            });

            this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_DEV).valueChanges.subscribe(data => {
                if (!this.modeClusterStandalone) {
                    this.formGroup.get(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE).setValue(data);
                }
                this.formGroup.controls[VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE].updateValueAndValidity();
            });

            this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_PROD).valueChanges.subscribe(data => {
                if (!this.modeClusterStandalone) {
                    this.formGroup.get(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE).setValue(data);
                }
                this.formGroup.controls[VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE].updateValueAndValidity();
            });

            if (this.edition !== AppEdition.TKG) {
                this.resurrectField(VsphereField.NODESETTING_CLUSTER_NAME,
                    [Validators.required, this.validationService.isValidClusterName()],
                    this.formGroup.get(VsphereField.NODESETTING_CLUSTER_NAME).value);
            }
        });
    }

    setSavedDataAfterLoad() {
        if (this.hasSavedData()) {
            const savedInstanceType = this.getSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, '');
            this.cardClick(savedInstanceType === InstanceType.PROD ? InstanceType.PROD : InstanceType.DEV);
            super.setSavedDataAfterLoad();
            // set the node type ID by finding it by the node type name
            let savedNodeType = this.nodeTypes.find(n => n.name === this.getSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, ''));
            if (savedNodeType) {
                this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_DEV).setValue(savedNodeType.id);
            }
            savedNodeType = this.nodeTypes.find(n => n.name === this.getSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_PROD, ''));
            if (savedNodeType) {
                this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_PROD).setValue(savedNodeType.id);
            }
            savedNodeType = this.nodeTypes.find(n => n.name === this.getSavedValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, ''));
            this.formGroup.get(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE).setValue(savedNodeType ? savedNodeType.id : '');
        }
    }

    onControlPlaneEndpoingProviderChange(provider: string): void {
        this.currentControlPlaneEndpoingProvider = provider;
        Broker.messenger.publish({
            type: TkgEventType.CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
            payload: provider
        });
        this.resurrectField(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, (provider === KUBE_VIP) ? [
            Validators.required,
            this.ipFamily === IpFamilyEnum.IPv4 ? this.validationService.isValidIpOrFqdn() : this.validationService.isValidIpv6OrFqdn()
        ] : [
            this.ipFamily === IpFamilyEnum.IPv4 ? this.validationService.isValidIpOrFqdn() : this.validationService.isValidIpv6OrFqdn()
        ], this.getSavedValue(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, ''));

        this.controlPlaneEndpointOptional = (provider === KUBE_VIP ? '' : '(OPTIONAL)');
    }

    cardClick(envType: string) {
        this.formGroup.controls[VsphereField.NODESETTING_CONTROL_PLANE_SETTING].setValue(envType);
    }

    getEnvType(): string {
        return this.formGroup.controls[VsphereField.NODESETTING_CONTROL_PLANE_SETTING].value;
    }
}
