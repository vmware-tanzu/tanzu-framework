// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// App imports
import { AzureInstanceType } from 'src/app/swagger/models';
import { AzureNodeSettingStandaloneStepMapping, AzureNodeSettingStepMapping } from './node-setting-step.fieldmapping';
import AppServices from '../../../../shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { ClusterPlan } from '../../wizard/shared/constants/wizard.constants';
import { AzureField } from '../azure-wizard.constants';

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {

    nodeTypes: AzureInstanceType[] = [];
    clusterPlan: string;
    currentRegion = 'US-WEST';
    displayForm = false;
    clusterNameInstruction: string;

    constructor(private validationService: ValidationService) {
        super();
        this.nodeTypes = [];
    }

    private supplyStepMapping(): StepMapping {
        const fieldMappings = this.modeClusterStandalone ? AzureNodeSettingStandaloneStepMapping : AzureNodeSettingStepMapping;
        AppServices.fieldMapUtilities.getFieldMapping(AzureField.NODESETTING_MANAGEMENT_CLUSTER_NAME, fieldMappings).required =
            AppServices.appDataService.isClusterNameRequired();
        // dynamically modify the cluster name label based on the type descriptor and whether the cluster name is required
        const clusterNameMapping = AppServices.fieldMapUtilities.getFieldMapping(AzureField.NODESETTING_MANAGEMENT_CLUSTER_NAME,
            fieldMappings);
        let clusterNameLabel = this.clusterTypeDescriptor.toUpperCase() + ' CLUSTER NAME';
        if (!AppServices.appDataService.isClusterNameRequired()) {
            clusterNameLabel += ' (OPTIONAL)';
        }
        clusterNameMapping.label = clusterNameLabel;
        return fieldMappings;
    }

    private subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe(this,
            TanzuEventType.AZURE_GET_INSTANCE_TYPES, this.onFetchedInstanceTypes.bind(this))
    }

    private onFetchedInstanceTypes(instanceTypes: AzureInstanceType[]) {
        this.nodeTypes = instanceTypes.sort();
        if (!this.modeClusterStandalone && this.nodeTypes.length === 1) {
            this.formGroup.get(AzureField.NODESETTING_WORKERTYPE).setValue(this.nodeTypes[0].name);
        }
    }

    listenOnChangeClusterPlan() {
        setTimeout(_ => {
            this.displayForm = true;
            const controlPlaneSettingControl = this.formGroup.get(AzureField.NODESETTING_CONTROL_PLANE_SETTING);
            if (controlPlaneSettingControl) {
                controlPlaneSettingControl.valueChanges.subscribe(data => {
                    if (data === ClusterPlan.DEV) {
                        this.setDevCardValidations();
                    } else if (data === ClusterPlan.PROD) {
                        this.setProdCardValidations();
                    }
                });
            } else {
                console.log('WARNING: azure-wizard.node-setting-step.listenOnChangeClusterPlan() cannot find controlPlaneSettingControl!');
            }
        });
    }

    setDevCardValidations() {
        this.clusterPlan = ClusterPlan.DEV;
        this.formGroup.markAsPending();
        this.resurrectField(
            AzureField.NODESETTING_INSTANCE_TYPE_DEV,
            [Validators.required],
            this.nodeTypes.length === 1 ? this.nodeTypes[0].name : '',
            { onlySelf: true, emitEvent: false }
        );
        this.disarmField(AzureField.NODESETTING_INSTANCE_TYPE_PROD, true);
    }

    setProdCardValidations() {
        this.clusterPlan = ClusterPlan.PROD;
        this.disarmField(AzureField.NODESETTING_INSTANCE_TYPE_DEV, true);
        this.formGroup.markAsPending();
        this.resurrectField(
            AzureField.NODESETTING_INSTANCE_TYPE_PROD,
            [Validators.required],
            this.nodeTypes.length === 1 ? this.nodeTypes[0].name : '',
            { onlySelf: true, emitEvent: false }
        );
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, this.supplyStepMapping());
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(this.supplyStepMapping());
        this.storeDefaultLabels(this.supplyStepMapping());
        this.setClusterNameInstruction();
        this.subscribeToServices();
        this.registerStepDescriptionTriggers({ clusterTypeDescriptor: true, fields: ['controlPlaneSetting']});
        this.registerDefaultFileImportedHandler(this.eventFileImported, this.supplyStepMapping());
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);
        this.listenOnChangeClusterPlan();

        this.chooseInitialClusterPlan();
    }

    chooseInitialClusterPlan() {
        const isProdClusterPlan = !this.getStoredValue(AzureField.NODESETTING_INSTANCE_TYPE_DEV, this.supplyStepMapping());
        this.cardClick(isProdClusterPlan ? ClusterPlan.PROD : ClusterPlan.DEV);
        isProdClusterPlan ? this.setProdCardValidations() : this.setDevCardValidations()
    }

    cardClick(envType: string) {
        this.setControlValueSafely(AzureField.NODESETTING_CONTROL_PLANE_SETTING, envType);
    }

    getEnvType(): string {
        return this.formGroup.controls[AzureField.NODESETTING_CONTROL_PLANE_SETTING].value;
    }

    dynamicDescription(): string {
        if (this.clusterPlan) {
            return 'Control plane type: ' + this.clusterPlan;
        }
        return 'Specify the resources backing the ' + this.clusterTypeDescriptor + ' cluster';
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(this.supplyStepMapping());
        this.storeDefaultDisplayOrder(this.supplyStepMapping());
    }

    private setClusterNameInstruction() {
        if (AppServices.appDataService.isClusterNameRequired()) {
            this.clusterNameInstruction = 'Specify a name for the ' + this.clusterTypeDescriptor + ' cluster.';
        } else {
            this.clusterNameInstruction = 'Optionally specify a name for the ' + this.clusterTypeDescriptor + ' cluster. ' +
                'If left blank, the installer names the cluster automatically.';
        }
    }
}
