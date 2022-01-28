// Angular imports
import { Component, Input, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// App imports
import AppServices from 'src/app/shared/service/appServices';
import { ClusterPlan, NodeType } from '../../wizard/shared/constants/wizard.constants';
import { InstanceType, IpFamilyEnum, PROVIDERS, Providers } from '../../../../shared/constants/app.constants';
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from 'src/app/shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VsphereField, VsphereNodeTypes } from '../vsphere-wizard.constants';
import { VsphereNodeSettingStepMapping, VsphereNodeSettingStandaloneStepMapping } from './node-setting-step.fieldmapping';

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

    constructor(private validationService: ValidationService) {
        super();
        this.nodeTypes = [...VsphereNodeTypes];
    }

    private supplyStepMapping(): StepMapping {
        const fieldMappings = this.modeClusterStandalone ? VsphereNodeSettingStandaloneStepMapping : VsphereNodeSettingStepMapping;
        AppServices.fieldMapUtilities.getFieldMapping(VsphereField.NODESETTING_CLUSTER_NAME, fieldMappings).required =
            AppServices.appDataService.isClusterNameRequired();
        // customize the cluster name label
        const clusterNameMapping = AppServices.fieldMapUtilities.getFieldMapping(VsphereField.NODESETTING_CLUSTER_NAME, fieldMappings);
        let customLabel = this.clusterTypeDescriptor.toUpperCase() + ' CLUSTER NAME';
        if (!AppServices.appDataService.isClusterNameRequired()) {
            customLabel += ' (OPTIONAL)';
        }
        clusterNameMapping.label = customLabel;
        return fieldMappings;
    }

    private customizeForm() {
        this.registerStepDescriptionTriggers({ clusterTypeDescriptor: true, fields: [VsphereField.NODESETTING_CONTROL_PLANE_SETTING]});
        this.registerOnValueChange(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER,
            this.onControlPlaneEndpoingProviderChange.bind(this));
        this.registerOnIpFamilyChange(VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, [
            Validators.required,
            this.validationService.isValidIpOrFqdn()
        ], [
            Validators.required,
            this.validationService.isValidIpv6OrFqdn()
        ]);
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.fieldMapUtilities.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.registerDefaultFileImportedHandler(this.eventFileImported, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        this.customizeForm();

        // TODO: can some of these subscriptions be moved to customizeForm()?
        setTimeout(_ => {
            this.displayForm = true;
            this.registerOnValueChange(VsphereField.NODESETTING_CONTROL_PLANE_SETTING, this.onControlPlaneSettingChange.bind(this));

            this.formGroup.get(VsphereField.NODESETTING_INSTANCE_TYPE_DEV).valueChanges.subscribe(data => {
                if (!this.modeClusterStandalone && data) {
                    // The user has just selected a new instance type for the DEV management cluster.
                    // If the worker node type hasn't been set, default to the same node type
                    const currentWorkerInstanceType = this.getFieldValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE);
                    if (!currentWorkerInstanceType) {
                        this.setControlValueSafely(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, data);
                    }
                    this.clearControlValue(VsphereField.NODESETTING_INSTANCE_TYPE_PROD);
                    this.formGroup.controls[VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE].updateValueAndValidity();
                }
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
                    this.formGroup.controls[VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE].updateValueAndValidity();
                }
            });
        });

        this.initFormWithSavedData();
    }

    private onControlPlaneSettingChange(data) {
        if (data === InstanceType.DEV) {
            this.onDevCardClicked();
        } else if (data === InstanceType.PROD) {
            this.onProdCardClicked();
        }
    }

    // findNodeTypeByNameOrId accommodates the fact that when we save the node type in local storage, we may either be saving the
    // name only (ie the display value - the 'old' method), or we may have saved the key (ie the node type id - the 'new' method).
    // Whichever value was saved, they are all unique, so we check if the saved value matches the name OR the id.
    private findNodeTypeByNameOrId(nameOrId: string): NodeType {
        return this.nodeTypes.find(n => n.name === nameOrId || n.id === nameOrId);
    }

    private onDevCardClicked() {
        this.nodeType = InstanceType.DEV;
        this.resurrectFieldWithSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV, [Validators.required]);
        const devInstanceType = this.getFieldValue(VsphereField.NODESETTING_INSTANCE_TYPE_DEV);
        // If there's no worker instance type, set it to the DEV value
        if (!this.modeClusterStandalone && devInstanceType && !this.getFieldValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE)) {
            this.setControlValueSafely(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, devInstanceType);
        }
        this.disarmField(VsphereField.NODESETTING_INSTANCE_TYPE_PROD);
    }

    private onProdCardClicked() {
        this.nodeType = InstanceType.PROD;
        this.resurrectFieldWithSavedValue(VsphereField.NODESETTING_INSTANCE_TYPE_PROD, [Validators.required]);
        if (!this.modeClusterStandalone) {
            this.resurrectFieldWithSavedValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, [Validators.required]);
        }
        this.disarmField(VsphereField.NODESETTING_INSTANCE_TYPE_DEV);
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
                this.setControlWithSavedValue(VsphereField.NODESETTING_ENABLE_AUDIT_LOGGING, false);
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
            if (!this.modeClusterStandalone) {
                const savedWorkerNodeNameOrId = this.getSavedValue(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, '');
                const savedWorkerNodeType = this.findNodeTypeByNameOrId(savedWorkerNodeNameOrId);
                const valueToUse = savedWorkerNodeType ? savedWorkerNodeType.id : '';
                this.setControlValueSafely(VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, valueToUse);
            }
        }
    }

    onControlPlaneEndpoingProviderChange(provider: string): void {
        this.currentControlPlaneEndpoingProvider = provider;
        AppServices.messenger.publish({
            type: TanzuEventType.VSPHERE_CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
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
        return this.formGroup.controls[VsphereField.NODESETTING_CONTROL_PLANE_SETTING].value;
    }

    dynamicDescription(): string {
        const ctlPlaneFlavor = this.getFieldValue(VsphereField.NODESETTING_CONTROL_PLANE_SETTING, true);
        if (ctlPlaneFlavor === ClusterPlan.PROD) {
            return 'Production cluster selected: 3 node control plane';
        } else if (ctlPlaneFlavor === ClusterPlan.DEV) {
            return 'Development cluster selected: 1 node control plane';
        }
        return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDefaultDisplayOrder(this.supplyStepMapping());
    }
}
