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

@Component({
    selector: 'app-node-setting-step',
    templateUrl: './node-setting-step.component.html',
    styleUrls: ['./node-setting-step.component.scss']
})
export class NodeSettingStepComponent extends StepFormDirective implements OnInit {

    nodeTypes: AzureInstanceType[] = [];
    nodeType: string;   // 'prod' or 'dev'
    currentRegion = "US-WEST";
    displayForm = false;

    constructor(private validationService: ValidationService,
        private fieldMapUtilities: FieldMapUtilities) {
        super();
        this.nodeTypes = [];
    }

    private supplyStepMapping(): StepMapping {
        const fieldMappings = this.modeClusterStandalone ? AzureNodeSettingStandaloneStepMapping : AzureNodeSettingStepMapping;
        FieldMapUtilities.getFieldMapping('managementClusterName', fieldMappings).required =
            AppServices.appDataService.isClusterNameRequired();
        return fieldMappings;
    }

    private subscribeToServices() {
        AppServices.dataServiceRegistrar.stepSubscribe(this,
            TanzuEventType.AZURE_GET_INSTANCE_TYPES, this.onFetchedInstanceTypes.bind(this))
    }

    private onFetchedInstanceTypes(instanceTypes: AzureInstanceType[]) {
        this.nodeTypes = instanceTypes.sort();
        if (!this.modeClusterStandalone && this.nodeTypes.length === 1) {
            this.formGroup.get('workerNodeInstanceType').setValue(this.nodeTypes[0].name);
        }
    }

    toggleValidations() {
        setTimeout(_ => {
            this.displayForm = true;
            const controlPlaneSettingControl = this.formGroup.get('controlPlaneSetting');
            if (controlPlaneSettingControl) {
                controlPlaneSettingControl.valueChanges.subscribe(data => {
                    if (data === 'dev') {
                        this.setDevCardValidations();
                    } else if (data === 'prod') {
                        this.setProdCardValidations();
                    }
                });
            } else {
                console.log('WARNING: azure-wizard.node-setting-step.toggleValidations() unable to find controlPlaneSettingControl!');
            }
        });
    }

    setDevCardValidations() {
        this.nodeType = 'dev';
        this.formGroup.markAsPending();
        this.resurrectField(
            'devInstanceType',
            [Validators.required],
            this.nodeTypes.length === 1 ? this.nodeTypes[0].name : '',
            { onlySelf: true, emitEvent: false }
        );
        this.disarmField('prodInstanceType', true);
    }

    setProdCardValidations() {
        this.nodeType = 'prod';
        this.disarmField('devInstanceType', true);
        this.formGroup.markAsPending();
        this.resurrectField(
            'prodInstanceType',
            [Validators.required],
            this.nodeTypes.length === 1 ? this.nodeTypes[0].name : '',
            { onlySelf: true, emitEvent: false }
        );
    }

    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, this.supplyStepMapping());
        this.subscribeToServices();
        this.registerStepDescriptionTriggers({ clusterTypeDescriptor: true, fields: ['controlPlaneSetting'] });
        this.toggleValidations();
        this.initFormWithSavedData();
    }

    initFormWithSavedData() {
        this.cardClick(this.getSavedValue('devInstanceType', '') === '' ? 'prod' : 'dev');
        this.getSavedValue('devInstanceType', '') === '' ? this.setProdCardValidations() : this.setDevCardValidations()
        super.initFormWithSavedData();
        // because it's in its own component, the enable audit logging field does not get initialized in the above call to
        // super.initFormWithSavedData()
        setTimeout(() => {
            this.setControlWithSavedValue('enableAuditLogging', false);
        })
    }

    cardClick(envType: string) {
        this.setControlValueSafely('controlPlaneSetting', envType);
    }

    getEnvType(): string {
        return this.formGroup.controls['controlPlaneSetting'].value;
    }

    dynamicDescription(): string {
        if (this.nodeType) {
            return 'Control plane type: ' + this.nodeType;
        }
        return 'Specify the resources backing the ' + this.clusterTypeDescriptor + ' cluster';
    }
}
