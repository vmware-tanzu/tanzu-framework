import { TkgEventType } from 'src/app/shared/service/Messenger';
/**
 * Angular Modules
 */
import { Component, Input, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
/**
 * App imports
 */
import { InstanceType, IpFamilyEnum, PROVIDERS, Providers } from '../../../../shared/constants/app.constants';
import { NodeType } from '../../wizard/shared/constants/wizard.constants';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import Broker from 'src/app/shared/service/broker';
import { AppEdition } from 'src/app/shared/constants/branding.constants';
import { VsphereField, VsphereNodeTypes } from '../vsphere-wizard.constants';
import { FormUtils } from '../../wizard/shared/utils/form-utils';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {
    @Input() providerType: string;

    nodeTypes: Array<NodeType> = [];
    PROVIDERS: Providers = PROVIDERS;
    vSphereNodeTypes: Array<NodeType> = VsphereNodeTypes;
    nodeType: string;

    displayForm = false;

    controlPlaneEndpointProviders = [KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER];
    currentControlPlaneEndpoingProvider = KUBE_VIP;
    controlPlaneEndpointOptional = "";

    constructor(private validationService: ValidationService,
        private wizardFormService: VSphereWizardFormService) {

        super();
        this.nodeTypes = [...VsphereNodeTypes];
    }

    ngOnInit() {
        super.ngOnInit();
        FormUtils.addControl(
            this.formGroup,
            VsphereField.NODESETTING_CONTROL_PLANE_SETTING,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.NODESETTING_INSTANCE_TYPE_DEV,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.NODESETTING_INSTANCE_TYPE_PROD,
            new FormControl('', [
                Validators.required
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED,
            new FormControl(true, [])
        );
        if (!this.modeClusterStandalone) {
            FormUtils.addControl(
            this.formGroup,
                VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE,
                new FormControl('', [
                    Validators.required
                ])
            );
        }
        FormUtils.addControl(
            this.formGroup,
            VsphereField.NODESETTING_CLUSTER_NAME,
            new FormControl('', [
                this.validationService.isValidClusterName()
            ])
        );
        FormUtils.addControl(
            this.formGroup,
            VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP,
            new FormControl('', [
                Validators.required,
                this.validationService.isValidIpOrFqdn()
            ])
        );

        FormUtils.addControl(
            this.formGroup,
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
                if (!this.modeClusterStandalone && data) {
                    // The user has just selected a new instance type for the DEV management cluster.
                    // If the worker node type hasn't been set, default to the same node type
                    const currentWorkerInstanceType = this.getFieldValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE);
                    if (!currentWorkerInstanceType) {
                        this.setControlValueSafely(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, data);
                    }
                    this.clearControlValue(VsphereField.NODESETTING_INSTANCE_TYPE_PROD);
                }
                this.formGroup.controls[VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE].updateValueAndValidity();
            });

            this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_PROD).valueChanges.subscribe(data => {
                if (!this.modeClusterStandalone && data) {
                    // The user has just selected a new instance type for the PROD management cluster.
                    // If the worker node type hasn't been set, default to the same node type
                    const currentWorkerInstanceType = this.getFieldValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE);
                    if (!currentWorkerInstanceType) {
                        this.setControlValueSafely(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, data);
                    }
                    this.clearControlValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV);
                }
                this.formGroup.controls[VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE].updateValueAndValidity();
            });

            if (this.edition !== AppEdition.TKG) {
                this.resurrectField(VsphereField.NODESETTING_CLUSTER_NAME,
                    [Validators.required, this.validationService.isValidClusterName()],
                    this.formGroup.get(VsphereField.NODESETTING_CLUSTER_NAME).value);
            }
        });

        this.initFormWithSavedData();
    }

    // findNodeTypeByNameOrId accommodates the fact that when we save the node type in local storage, we may either be saving the
    // name only (ie the display value - the 'old' method), or we may have saved the key (ie the node type id - the 'new' method).
    // Whichever value was saved, they are all unique, so we check if the saved value matches the name OR the id.
    private findNodeTypeByNameOrId(nameOrId: string): NodeType {
        return this.nodeTypes.find(n => n.name === nameOrId || n.id === nameOrId);
    }

    initFormWithSavedData() {
        if (this.hasSavedData()) {
            // Is the configuration using a DEV or a PROD instance type? We check the saved value of the DEV node instance to determine
            // if the user was in DEV node instance mode
            const savedDevInstanceType = this.getSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, '');
            const managementClusterType = savedDevInstanceType !== '' ? InstanceType.DEV : InstanceType.PROD;
            this.cardClick(managementClusterType);
            super.initFormWithSavedData();
            // because it's in its own component, the enable audit logging field does not get initialized in the above call to
            // super.initFormWithSavedData()
            setTimeout( () => {
                this.setControlWithSavedValue('enableAuditLogging', false);
            })

            if (managementClusterType === InstanceType.DEV) {
                // set the node type ID by finding it by the node type name OR the id
                const savedNameOrId = this.getSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, '');
                const savedNodeType = this.findNodeTypeByNameOrId(savedNameOrId);
                if (savedNodeType) {
                    this.setControlValueSafely(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, savedNodeType.id);
                }
                this.disarmField(VsphereField.NODESETTING_INSTANCE_TYPE_PROD, true);
            } else {
                // set the node type ID by finding it by the node type name OR the id
                const savedNameOrId = this.getSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_PROD, '');
                const savedNodeType = this.findNodeTypeByNameOrId(savedNameOrId);
                if (savedNodeType) {
                    this.setControlValueSafely(VsphereField.NODESETTING_INSTANCE_TYPE_PROD, savedNodeType.id);
                }
                this.disarmField(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, true);
            }
            const savedWorkerNodeNameOrId = this.getSavedValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, '');
            const savedWorkerNodeType = this.findNodeTypeByNameOrId(savedWorkerNodeNameOrId);
            const valueToUse = savedWorkerNodeType ? savedWorkerNodeType.id : '';
            this.setControlValueSafely(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, valueToUse);
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
        const controlPlaneSetting = this.formGroup.controls[VsphereField.NODESETTING_CONTROL_PLANE_SETTING];
        if (controlPlaneSetting === undefined) {
            console.log('===> WARNING: cardClick() unable to set controlPlaneSetting to "' + envType + '" because no controlPlaneSetting control was found!');
        } else {
            controlPlaneSetting.setValue(envType);
        }
    }

    getEnvType(): string {
        const controlPlaneSetting = this.formGroup.controls[VsphereField.NODESETTING_CONTROL_PLANE_SETTING];
        if (controlPlaneSetting === undefined) {
            console.log('getEnvType() unable to return env type because no controlPlaneSetting control was found!');
            return '';
        }
        return this.formGroup.controls['controlPlaneSetting'].value;
    }

    dynamicDescription(): string {
        const ctlPlaneFlavor = this.getFieldValue('controlPlaneSetting', true);
        if (ctlPlaneFlavor) {
            let mode = 'Development cluster selected: 1 node control plane';
            if (ctlPlaneFlavor === 'prod') {
                mode = 'Production cluster selected: 3 node control plane';
            }
            return mode;
        }
        return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
    }
}
