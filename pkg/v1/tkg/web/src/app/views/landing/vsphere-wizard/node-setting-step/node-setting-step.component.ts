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
import { PROVIDERS, Providers } from '../../../../shared/constants/app.constants';
import { NodeType, vSphereNodeTypes } from '../../wizard/shared/constants/wizard.constants';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import Broker from 'src/app/shared/service/broker';

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
            'controlPlaneSetting',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'devInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'prodInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'machineHealthChecksEnabled',
            new FormControl(true, [])
        );
        this.formGroup.addControl(
            'workerNodeInstanceType',
            new FormControl('', [
                Validators.required
            ])
        );
        this.formGroup.addControl(
            'clusterName',
            new FormControl('', [
                this.validationService.isValidClusterName()
            ])
        );
        this.formGroup.addControl(
            'controlPlaneEndpointIP',
            new FormControl('', [
                Validators.required,
                this.validationService.isValidIpOrFqdn()
            ])
        );

        this.formGroup.addControl(
            'controlPlaneEndpointProvider',
            new FormControl(this.currentControlPlaneEndpoingProvider, [
                Validators.required
            ])
        );

        this.registerOnValueChange("controlPlaneEndpointProvider", this.onControlPlaneEndpoingProviderChange.bind(this));

        setTimeout(_ => {
            this.displayForm = true;
            this.formGroup.get('controlPlaneSetting').valueChanges.subscribe(data => {
                if (data === 'dev') {
                    this.nodeType = 'dev';
                    this.formGroup.get('devInstanceType').setValidators([
                        Validators.required
                    ]);
                    this.formGroup.controls['prodInstanceType'].clearValidators();
                    this.formGroup.controls['prodInstanceType'].setValue('');
                } else if (data === 'prod') {
                    this.nodeType = 'prod';
                    this.formGroup.controls['prodInstanceType'].setValidators([
                        Validators.required
                    ]);
                    this.formGroup.get('devInstanceType').clearValidators();
                    this.formGroup.controls['devInstanceType'].setValue('');
                }
                this.formGroup.get('devInstanceType').updateValueAndValidity();
                this.formGroup.controls['prodInstanceType'].updateValueAndValidity();
            });

            this.formGroup.get('devInstanceType').valueChanges.subscribe(data => {
                this.formGroup.get('workerNodeInstanceType').setValue(data);
                this.formGroup.controls['workerNodeInstanceType'].updateValueAndValidity();
            });

            this.formGroup.get('prodInstanceType').valueChanges.subscribe(data => {
                this.formGroup.get('workerNodeInstanceType').setValue(data);
                this.formGroup.controls['workerNodeInstanceType'].updateValueAndValidity();
            });
        });
    }

    setSavedDataAfterLoad() {
        if (this.hasSavedData()) {
            this.cardClick(this.getSavedValue('devInstanceType', '') === '' ? 'prod' : 'dev');
            super.setSavedDataAfterLoad();
            // set the node type ID by finding it by the node type name
            let savedNodeType = this.nodeTypes.find(n => n.name === this.getSavedValue('devInstanceType', ''));
            if (savedNodeType) {
                this.formGroup.get('devInstanceType').setValue(savedNodeType.id);
            }
            savedNodeType = this.nodeTypes.find(n => n.name === this.getSavedValue('prodInstanceType', ''));
            if (savedNodeType) {
                this.formGroup.get('prodInstanceType').setValue(savedNodeType.id);
            }
            savedNodeType = this.nodeTypes.find(n => n.name === this.getSavedValue('workerNodeInstanceType', ''));
            this.formGroup.get('workerNodeInstanceType').setValue(savedNodeType ? savedNodeType.id : '');
        }
    }

    onControlPlaneEndpoingProviderChange(provider: string): void {
        this.currentControlPlaneEndpoingProvider = provider;
        Broker.messenger.publish({
            type: TkgEventType.CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
            payload: provider
        });
        this.resurrectField("controlPlaneEndpointIP", (provider === KUBE_VIP) ? [
            Validators.required,
            this.validationService.isValidIpOrFqdn()
        ] : [
            this.validationService.isValidIpOrFqdn()
        ], this.getSavedValue("controlPlaneEndpointIP", ""));

        this.controlPlaneEndpointOptional = (provider === KUBE_VIP ? "" : "(OPTIONAL)");
    }

    cardClick(envType: string) {
        this.formGroup.controls['controlPlaneSetting'].setValue(envType);
    }

    getEnvType(): string {
        return this.formGroup.controls['controlPlaneSetting'].value;
    }
}
