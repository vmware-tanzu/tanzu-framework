import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { takeUntil, debounceTime, distinctUntilChanged } from 'rxjs/operators';

import { TkgEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { Vpc } from '../../../../swagger/models/vpc.model';
import { AwsWizardFormService } from '../../../../shared/service/aws-wizard-form.service';
import Broker from 'src/app/shared/service/broker';
import {AwsField} from "../aws-wizard.constants";

enum VpcType {
    EXISTING = 'existing',
    NEW = 'new'
}

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
            AwsField.VPC_TYPE,
            new FormControl(
                VpcType.NEW, [
                Validators.required
            ])
        );

        this.formGroup.addControl(
            AwsField.VPC_NEW_CIDR,
            new FormControl('', [])
        );

        this.formGroup.addControl(
            AwsField.VPC_EXISTING_CIDR,
            new FormControl('', [])
        );

        this.formGroup.addControl(
            AwsField.VPC_EXISTING_ID,
            new FormControl('', [])
        );

        this.formGroup.addControl(
            AwsField.VPC_NON_INTERNET_FACING,
            new FormControl(false, [])
        );

        this.formGroup.get(AwsField.VPC_TYPE).valueChanges
            .pipe(
                distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
                takeUntil(this.unsubscribe)
            ).subscribe((val) => {
                if (val === VpcType.EXISTING) {
                    Broker.messenger.publish({
                        type: TkgEventType.AWS_VPC_TYPE_CHANGED,
                        payload: { vpcType: VpcType.EXISTING.toString() }
                    });
                    if (this.existingVpcs && this.existingVpcs.length === 1) {
                        this.formGroup.get(AwsField.VPC_EXISTING_ID).setValue(this.existingVpcs[0].id);
                        this.formGroup.get(AwsField.VPC_EXISTING_CIDR).setValue(this.existingVpcs[0].cidr);
                    }
                    this.formGroup.get(AwsField.VPC_NEW_CIDR).clearValidators();
                    this.formGroup.get(AwsField.VPC_NEW_CIDR).setValue('');
                    this.clearFieldSavedData(AwsField.VPC_NEW_CIDR);
                    this.setExistingVpcValidators();
                } else {
                    this.formGroup.get(AwsField.VPC_EXISTING_ID).setValue('');
                    this.formGroup.get(AwsField.VPC_EXISTING_ID).clearValidators();
                    this.formGroup.get(AwsField.VPC_EXISTING_ID).updateValueAndValidity();
                    this.formGroup.get(AwsField.VPC_EXISTING_CIDR).setValue('');
                    this.formGroup.get(AwsField.VPC_EXISTING_CIDR).clearValidators();
                    this.formGroup.get(AwsField.VPC_EXISTING_CIDR).updateValueAndValidity();
                    this.clearFieldSavedData(AwsField.VPC_EXISTING_CIDR);
                    this.clearFieldSavedData(AwsField.VPC_EXISTING_ID);
                    this.setNewVpcValidators();
                    Broker.messenger.publish({
                        type: TkgEventType.AWS_VPC_TYPE_CHANGED,
                        payload: { vpcType: VpcType.NEW.toString() }
                    });

                }
            }
            );

        const vpcCidrs = [AwsField.VPC_NEW_CIDR, AwsField.VPC_EXISTING_CIDR];
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
                if (this.formGroup.get(AwsField.VPC_EXISTING_ID)) {
                    this.existingVpcs = [];
                    this.formGroup.get(AwsField.VPC_EXISTING_ID).setValue('');
                    this.formGroup.get(AwsField.VPC_EXISTING_CIDR).setValue('');
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
            payload: { vpcType: VpcType.NEW.toString() }
        });

        this.registerOnValueChange(AwsField.VPC_NON_INTERNET_FACING, this.onNonInternetFacingVPCChange.bind(this));
        this.initFormWithSavedData();
    }

    onNonInternetFacingVPCChange(checked: boolean) {
        Broker.messenger.publish({
            type: TkgEventType.AWS_AIRGAPPED_VPC_CHANGE,
            payload: checked === true
        });
    }

    initFormWithSavedData() {
        if (!this.hasSavedData() || this.getSavedValue(AwsField.VPC_NEW_CIDR, '') !== '') {
            this.setNewVpcValidators();
        } else {
            this.formGroup.get(AwsField.VPC_TYPE).setValue(VpcType.EXISTING);
            this.setExistingVpcValidators();
        }
        super.initFormWithSavedData();
    }

    /**
     * @method setNewVpcValidators
     * helper method to consolidate setting validators for new vpc fields and
     * re-subscribe to vpc value changes
     */
    setNewVpcValidators() {
        this.defaultVpcHasChanged = false;

        this.formGroup.get(AwsField.VPC_NEW_CIDR).setValue(this.getSavedValue('vpc', this.defaultVpcAddress));
        this.formGroup.get(AwsField.VPC_NEW_CIDR).setValidators([
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ]);

    }

    setExistingVpcValidators() {
        this.formGroup.get(AwsField.VPC_EXISTING_ID).setValidators([Validators.required]);
        this.formGroup.get(AwsField.VPC_EXISTING_ID).updateValueAndValidity();
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
            this.formGroup.get(AwsField.VPC_EXISTING_CIDR).setValue(existingVpc[0].cidr);
        } else {
            this.formGroup.get(AwsField.VPC_EXISTING_CIDR).setValue('');
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
