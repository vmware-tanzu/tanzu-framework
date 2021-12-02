import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { takeUntil, debounceTime, distinctUntilChanged } from 'rxjs/operators';

import { TkgEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { Vpc } from '../../../../swagger/models/vpc.model';
import { AwsWizardFormService } from '../../../../shared/service/aws-wizard-form.service';
import Broker from 'src/app/shared/service/broker';

@Component({
    selector: 'app-vpc-step',
    templateUrl: './vpc-step.component.html',
    styleUrls: ['./vpc-step.component.scss']
})
export class VpcStepComponent extends StepFormDirective implements OnInit {

    defaultVpcHasChanged: boolean = false;
    existingVpcs: Array<Vpc>;
    loadingExistingVpcs: boolean = false;

    defaultVpcAddress: string = '10.0.0.0/16';

    constructor(private validationService: ValidationService,
        private awsWizardFormService: AwsWizardFormService) {
        super();
    }

    ngOnInit() {
        super.ngOnInit();

        this.formGroup.addControl(
            'vpcType',
            new FormControl(
                'new', [
                Validators.required
            ])
        );

        this.formGroup.addControl(
            'vpc',
            new FormControl('', [])
        );

        this.formGroup.addControl(
            'existingVpcCidr',
            new FormControl('', [])
        );

        this.formGroup.addControl(
            'existingVpcId',
            new FormControl('', [])
        );

        this.formGroup.addControl(
            'nonInternetFacingVPC',
            new FormControl(false, [])
        );

        this.formGroup.get('vpcType').valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                if (val === 'existing') {
                    Broker.messenger.publish({
                        type: TkgEventType.AWS_VPC_TYPE_CHANGED,
                        payload: { vpcType: 'existing' }
                    });
                    if (this.existingVpcs && this.existingVpcs.length === 1) {
                        this.formGroup.get('existingVpcId').setValue(this.existingVpcs[0].id);
                        this.formGroup.get('existingVpcCidr').setValue(this.existingVpcs[0].cidr);
                    }
                    this.formGroup.get('vpc').clearValidators();
                    this.formGroup.get('vpc').setValue('');
                    this.clearFieldSavedData('vpc');
                    this.setExistingVpcValidators();
                } else {
                    this.formGroup.get('existingVpcId').setValue('');
                    this.formGroup.get('existingVpcId').clearValidators();
                    this.formGroup.get('existingVpcId').updateValueAndValidity();
                    this.formGroup.get('existingVpcCidr').setValue('');
                    this.formGroup.get('existingVpcCidr').clearValidators();
                    this.formGroup.get('existingVpcCidr').updateValueAndValidity();
                    this.clearFieldSavedData('existingVpcCidr');
                    this.clearFieldSavedData('existingVpcId');
                    this.setNewVpcValidators();
                    Broker.messenger.publish({
                        type: TkgEventType.AWS_VPC_TYPE_CHANGED,
                        payload: { vpcType: 'new' }
                    });

                }
            }
            );

        const vpcCidrs = ['vpc', 'existingVpcCidr'];
        vpcCidrs.forEach(vpcCidr => {
            this.formGroup.get(vpcCidr).valueChanges.pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((cidr) => {
                Broker.messenger.publish({
                    type: TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO,
                    payload: { info: (cidr ? cidr + ',' : '') + '169.254.0.0/16' }
                });
            });
        });

        /**
         * Whenever aws region selection changes, update AZ subregion
         */
        Broker.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                if (this.formGroup.get('existingVpcId')) {
                    this.existingVpcs = [];
                    this.formGroup.get('existingVpcId').setValue('');
                    this.formGroup.get('existingVpcCidr').setValue('');
                }
            });

        this.awsWizardFormService.getErrorStream(TkgEventType.AWS_GET_EXISTING_VPCS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(error => {
                this.errorNotification = error;
            });

        this.awsWizardFormService.getDataStream(TkgEventType.AWS_GET_EXISTING_VPCS)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((vpcs: Array<Vpc>) => {
                this.existingVpcs = vpcs;
                this.loadingExistingVpcs = false;
            });

        // init vpc type to new
        Broker.messenger.publish({
            type: TkgEventType.AWS_VPC_TYPE_CHANGED,
            payload: { vpcType: 'new' }
        });

        this.registerOnValueChange('nonInternetFacingVPC', this.onNonInternetFacingVPCChange.bind(this));
    }

    onNonInternetFacingVPCChange(checked: boolean) {
        Broker.messenger.publish({
            type: TkgEventType.AWS_AIRGAPPED_VPC_CHANGE,
            payload: checked === true
        });
    }

    setSavedDataAfterLoad() {
        if (!this.hasSavedData() || this.getSavedValue('vpc', '') !== '') {
            this.setNewVpcValidators();
        } else {
            this.formGroup.get('vpcType').setValue('existing');
            this.setExistingVpcValidators();
        }
        super.setSavedDataAfterLoad();
    }

    /**
     * @method setNewVpcValidators
     * helper method to consolidate setting validators for new vpc fields and
     * re-subscribe to vpc value changes
     */
    setNewVpcValidators() {
        this.defaultVpcHasChanged = false;

        this.formGroup.get('vpc').setValue(this.getSavedValue('vpc', this.defaultVpcAddress));
        this.formGroup.get('vpc').setValidators([
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ]);

    }

    setExistingVpcValidators() {
        this.formGroup.get('existingVpcId').setValidators([Validators.required]);
        this.formGroup.get('existingVpcId').updateValueAndValidity();
    }

    /**
     * @method existingVpcOnChange
     * helper method to manually set existing VPC CIDR read-only value, and
     * dispatch message to retrieve VPC subnets by VPC ID
     * @param existingVpcId
     */
    existingVpcOnChange(existingVpcId: any) {
        const existingVpc: Array<Vpc> = this.existingVpcs.filter((vpc) => {
            return vpc.id === existingVpcId;
        });
        if (existingVpc && existingVpc.length > 0) {
            this.formGroup.get('existingVpcCidr').setValue(existingVpc[0].cidr);
        } else {
            this.formGroup.get('existingVpcCidr').setValue('');
        }

        Broker.messenger.publish({
            type: TkgEventType.AWS_GET_SUBNETS,
            payload: { vpcId: existingVpcId }
        });

        Broker.messenger.publish(({
            type: TkgEventType.AWS_VPC_CHANGED
        }));
    }
}
