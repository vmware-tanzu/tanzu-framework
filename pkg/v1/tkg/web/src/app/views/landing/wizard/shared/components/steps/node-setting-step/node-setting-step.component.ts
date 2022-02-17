import { StepFormDirective } from '../../../step-form/step-form';
import { Directive, OnInit } from '@angular/core';
import { ClusterPlan } from '../../../constants/wizard.constants';
import { StepMapping } from '../../../field-mapping/FieldMapping';
import AppServices from '../../../../../../../shared/service/appServices';
import { NodeSettingField, NodeSettingStepMapping } from './node-setting-step.fieldmapping';
import { Validators } from '@angular/forms';
import { ValidationService } from '../../../validation/validation.service';

@Directive()
export abstract class NodeSettingStepDirective<NODEINSTANCE> extends StepFormDirective implements OnInit {
    nodeTypes: Array<NODEINSTANCE> = [];
    clusterPlan: string;
    clusterNameInstruction: string;
    private stepMapping: StepMapping;

    protected abstract getKeyFromNodeInstance(nodeInstance: NODEINSTANCE): string;
    protected abstract getDisplayFromNodeInstance(nodeInstance: NODEINSTANCE): string;

    protected constructor(protected validationService: ValidationService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();

        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.registerDefaultFileImportedHandler(this.eventFileImported, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);
        this.subscribeToServices();
        this.listenToEvents();

        this.setClusterNameInstruction();
    }

    // available to HTML as handler for clicking on a cluster plan
    cardClickDev() {
        this.setControlPlaneToDev();
        this.triggerStepDescriptionChange();
    }

    cardClickProd() {
        this.setControlPlaneToProd();
        this.triggerStepDescriptionChange();
    }

    get devInstanceTypeValue() {
        return this.getFieldValue(NodeSettingField.INSTANCE_TYPE_DEV);
    }

    get prodInstanceTypeValue() {
        return this.getFieldValue(NodeSettingField.INSTANCE_TYPE_PROD);
    }

    // Extending classes should override this method by calling it first and then adding whatever additional field mappings they need
    protected createStepMapping(): StepMapping {
        return this.createDefaultStepMapping();
    }

    // Extending classes will likely override this method by calling it first and then adding whatever listeners they need
    protected listenToEvents() {
        this.registerOnValueChange(NodeSettingField.INSTANCE_TYPE_DEV, this.onDevInstanceTypeChange.bind(this));
        this.registerOnValueChange(NodeSettingField.INSTANCE_TYPE_PROD, this.onProdInstanceTypeChange.bind(this));
        this.registerStepDescriptionTriggers({ clusterTypeDescriptor: true });
    }

    // Extending classes may override this method if they have service to subscribe to
    protected subscribeToServices() {
    }

    protected onDevInstanceTypeChange(devNodeType: string) {
        if (devNodeType) {
            this.setWorkerInstanceTypeIfNotSet(devNodeType);
            this.clearControlValue(NodeSettingField.INSTANCE_TYPE_PROD);
        }
    }

    protected onProdInstanceTypeChange(prodNodeType: string) {
        if (prodNodeType) {
            this.setWorkerInstanceTypeIfNotSet(prodNodeType);
            this.clearControlValue(NodeSettingField.INSTANCE_TYPE_DEV);
        }
    }

    private setWorkerInstanceTypeIfNotSet(nodeType: string) {
        if (!this.modeClusterStandalone && nodeType) {
            // The user has just selected a new instance type for the management cluster.
            // If the worker node type hasn't been set, default to the same node type
            const workerNodeInstanceTypeControl = this.getControl(NodeSettingField.WORKER_NODE_INSTANCE_TYPE);
            if (!workerNodeInstanceTypeControl.value) {
                workerNodeInstanceTypeControl.setValue(nodeType);
                workerNodeInstanceTypeControl.updateValueAndValidity();
            }
        }
    }

    // Extending classes MAY need to override this method if they have additional changes dependent on cluster plan change to DEV
    protected setControlPlaneToDev() {
        this.clusterPlan = ClusterPlan.DEV;
        let valueToUse;
        if (this.nodeTypes.length === 1) {
            valueToUse = this.getKeyFromNodeInstance(this.nodeTypes[0]);
        } else {
            const existingValue = this.formGroup.get(NodeSettingField.INSTANCE_TYPE_DEV).value;
            valueToUse = this.getStoredValue(NodeSettingField.INSTANCE_TYPE_DEV, this.supplyStepMapping(), existingValue);
        }
        this.resurrectFieldWithStoredValue(NodeSettingField.INSTANCE_TYPE_DEV, this.supplyStepMapping(),
        [Validators.required], valueToUse);
        this.disarmField(NodeSettingField.INSTANCE_TYPE_PROD);
    }

    // Extending classes MAY need to override this method if they have additional changes dependent on cluster plan change to PROD
    protected setControlPlaneToProd() {
        this.clusterPlan = ClusterPlan.PROD;
        let valueToUse;
        if (this.nodeTypes.length === 1) {
            valueToUse = this.getKeyFromNodeInstance(this.nodeTypes[0]);
        } else {
            const existingValue = this.formGroup.get(NodeSettingField.INSTANCE_TYPE_PROD).value;
            valueToUse = this.getStoredValue(NodeSettingField.INSTANCE_TYPE_PROD, this.supplyStepMapping(), existingValue);
        }
        this.resurrectFieldWithStoredValue(NodeSettingField.INSTANCE_TYPE_PROD, this.supplyStepMapping(),
            [Validators.required], valueToUse);
        this.disarmField(NodeSettingField.INSTANCE_TYPE_DEV);
    }

    private createDefaultStepMapping(): StepMapping {
        const stepMapping = AppServices.fieldMapUtilities.cloneStepMapping(NodeSettingStepMapping);
        // if we're in standalone mode, deactivate the worker node instance field mapping (because it isn't used)
        const workerInstanceMapping =
            AppServices.fieldMapUtilities.getFieldMapping(NodeSettingField.WORKER_NODE_INSTANCE_TYPE, stepMapping);
        workerInstanceMapping.deactivated = AppServices.appDataService.isModeClusterStandalone();
        // dynamically modify the cluster field mapping
        const clusterNameMapping = AppServices.fieldMapUtilities.getFieldMapping(NodeSettingField.CLUSTER_NAME, stepMapping);
        clusterNameMapping.label = this.createClusterNameLabel();
        clusterNameMapping.required = this.isClusterNameRequired();

        return stepMapping;
    }

    protected chooseInitialClusterPlan() {
        const devInstanceType = this.getStoredValue(NodeSettingField.INSTANCE_TYPE_DEV, this.supplyStepMapping());
        if (devInstanceType) {
            this.setControlPlaneToDev();
            return;
        }
        const prodInstanceType = this.getStoredValue(NodeSettingField.INSTANCE_TYPE_PROD, this.supplyStepMapping());
        if (prodInstanceType) {
            this.setControlPlaneToProd();
            return;
        }
        this.setControlPlaneToDev();
    }

    // This method may be USED by subclasses, but should not be overwritten; subclasses should overwrite createStepMapping() instead.
    protected supplyStepMapping(): StepMapping {
        if (!this.stepMapping) {
            this.stepMapping = this.createStepMapping();
        }
        if (!this.stepMapping) {
            console.error('NodeSettingStep.supplyStepMapping() returning null');
        } else if (!this.stepMapping.fieldMappings) {
            console.error('NodeSettingStep.supplyStepMapping() returning object missing fieldMappings');
        }
        return this.stepMapping;
    }

    private createClusterNameLabel(): string {
        let clusterNameLabel = this.clusterTypeDescriptor.toUpperCase() + ' CLUSTER NAME';
        if (!AppServices.appDataService.isClusterNameRequired()) {
            clusterNameLabel += ' (OPTIONAL)';
        }
        return clusterNameLabel;
    }

    private setClusterNameInstruction() {
        if (AppServices.appDataService.isClusterNameRequired()) {
            this.clusterNameInstruction = 'Specify a name for the ' + this.clusterTypeDescriptor + ' cluster.';
        } else {
            this.clusterNameInstruction = 'Optionally specify a name for the ' + this.clusterTypeDescriptor + ' cluster. ' +
                'If left blank, the installer names the cluster automatically.';
        }
    }

    // Extending classes may want to override this method
    protected isClusterNameRequired():boolean {
        return AppServices.appDataService.isClusterNameRequired();
    }

    dynamicDescription(): string {
        if (this.isClusterPlanProd) {
            return 'Production cluster selected: 3 node control plane';
        } else if (this.isClusterPlanDev) {
            return 'Development cluster selected: 1 node control plane';
        }
        return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
    }

    get isClusterPlanProd(): boolean {
        return this.clusterPlan === ClusterPlan.PROD;
    }

    get isClusterPlanDev(): boolean {
        return this.clusterPlan === ClusterPlan.DEV;
    }

    // Extending classes should have no reason to override this method.
    protected storeUserData() {
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDisplayOrder(this.getFieldDisplayOrder());
    }

    // Extending classes may want to change the display order of the fields
    protected getFieldDisplayOrder() {
        return this.defaultDisplayOrder(this.supplyStepMapping());
    }

    protected onStepStarted() {
        this.chooseInitialClusterPlan();
    }
}
