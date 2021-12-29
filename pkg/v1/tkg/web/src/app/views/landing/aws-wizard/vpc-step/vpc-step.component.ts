// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
// App imports
import AppServices from 'src/app/shared/service/appServices';
import { AwsVpcStepMapping } from './vpc-step.fieldmapping';
import { AwsField, VpcType } from "../aws-wizard.constants";
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';
import { Vpc } from '../../../../swagger/models/vpc.model';

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
                private fieldMapUtilities: FieldMapUtilities) {
        super();
    }

    private onVpcTypeChange(newVpcType: VpcType) {
        const existingVpcControl = this.formGroup.get(AwsField.VPC_EXISTING_ID);
        const existingVpcCidrControl = this.formGroup.get(AwsField.VPC_EXISTING_CIDR);
        if (newVpcType === VpcType.EXISTING) {
            Broker.messenger.publish({
                type: TkgEventType.AWS_VPC_TYPE_CHANGED,
                payload: { vpcType: VpcType.EXISTING.toString() }
            });
            if (this.existingVpcs && this.existingVpcs.length === 1) {
                existingVpcControl.setValue(this.existingVpcs[0].id);
                existingVpcCidrControl.setValue(this.existingVpcs[0].cidr);
            }
            this.formGroup.get(AwsField.VPC_NEW_CIDR).clearValidators();
            this.clearControlValue(AwsField.VPC_NEW_CIDR);
            this.clearFieldSavedData(AwsField.VPC_NEW_CIDR);
            this.setExistingVpcValidators();
        } else {
            existingVpcControl.setValue('');
            existingVpcControl.clearValidators();
            existingVpcControl.updateValueAndValidity();
            existingVpcCidrControl.setValue('');
            existingVpcCidrControl.clearValidators();
            existingVpcCidrControl.updateValueAndValidity();
            this.clearFieldSavedData(AwsField.VPC_EXISTING_CIDR);
            this.clearFieldSavedData(AwsField.VPC_EXISTING_ID);
            this.setNewVpcValidators();
            Broker.messenger.publish({
                type: TkgEventType.AWS_VPC_TYPE_CHANGED,
                payload: { vpcType: VpcType.NEW.toString() }
            });
        }
        this.triggerStepDescriptionChange();
    }
    ngOnInit() {
        super.ngOnInit();
        this.fieldMapUtilities.buildForm(this.formGroup, this.formName, AwsVpcStepMapping);
        // NOTE: we don't call this.registerFieldsAffectingStepDescription() with the other fields, because the other relevant fields
        // already trigger a step description change event in their own onChange handlers
        this.registerFieldsAffectingStepDescription([AwsField.VPC_EXISTING_ID]);

        this.registerOnValueChange(AwsField.VPC_TYPE, this.onVpcTypeChange.bind(this));

        const cidrFields = [AwsField.VPC_NEW_CIDR, AwsField.VPC_EXISTING_CIDR];
        cidrFields.forEach(cidrField => {
            this.registerOnValueChange(cidrField, (cidr) => {
                AppServices.messenger.publish({
                    type: TkgEventType.NETWORK_STEP_GET_NO_PROXY_INFO,
                    payload: { info: (cidr ? cidr + ',' : '') + '169.254.0.0/16' }
                });
                this.triggerStepDescriptionChange();
            });
        });

        /**
         * Whenever aws region selection changes, update AZ subregion
         */
        AppServices.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                if (this.formGroup.get(AwsField.VPC_EXISTING_ID)) {
                    this.existingVpcs = [];
                    this.clearControlValue(AwsField.VPC_EXISTING_ID);
                    this.clearControlValue(AwsField.VPC_EXISTING_CIDR);
                }
            });

        AppServices.dataServiceRegistrar.stepSubscribe<Vpc>(this, TkgEventType.AWS_GET_EXISTING_VPCS, this.onFetchedVpcs.bind(this));

        // init vpc type to new
        AppServices.messenger.publish({
            type: TkgEventType.AWS_VPC_TYPE_CHANGED,
            payload: { vpcType: VpcType.NEW.toString() }
        });
        this.registerOnValueChange(AwsField.VPC_NON_INTERNET_FACING, this.onNonInternetFacingVPCChange.bind(this));
        this.initFormWithSavedData();
    }

    private onFetchedVpcs(vpcs: Array<Vpc>) {
        this.existingVpcs = vpcs;
        this.loadingExistingVpcs = false;
    }

    onNonInternetFacingVPCChange(checked: boolean) {
        AppServices.messenger.publish({
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

        this.formGroup.get(AwsField.VPC_NEW_CIDR).setValue(this.getSavedValue(AwsField.VPC_NEW_CIDR, this.defaultVpcAddress));
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
            // onlySelf onption changes value for the current control only.
            this.formGroup.get(AwsField.VPC_EXISTING_CIDR).setValue('', { onlySelf: true});
        }

        AppServices.messenger.publish({
            type: TkgEventType.AWS_GET_SUBNETS,
            payload: { vpcId: existingVpcId }
        });

        AppServices.messenger.publish(({
            type: TkgEventType.AWS_VPC_CHANGED
        }));
    }

    protected dynamicDescription(): string {
        const vpcType = this.getFieldValue(AwsField.VPC_TYPE);
        const vpcExistingCidr = this.getFieldValue(AwsField.VPC_EXISTING_CIDR, true);
        const vpcExistingId = this.getFieldValue(AwsField.VPC_EXISTING_ID, true);
        const vpcNewCidr = this.getFieldValue(AwsField.VPC_NEW_CIDR, true);

        if (vpcType === VpcType.EXISTING && vpcExistingCidr && vpcExistingId) {
            return 'VPC: ' + vpcExistingId + ' CIDR: ' + vpcExistingCidr;
        }
        if (vpcType === VpcType.NEW && vpcNewCidr) {
            return 'VPC: (new) CIDR: ' + vpcNewCidr;
        }
        return 'Specify VPC settings for AWS';
    }
}
