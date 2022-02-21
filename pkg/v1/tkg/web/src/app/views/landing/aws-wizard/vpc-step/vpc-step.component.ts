// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// App imports
import AppServices from 'src/app/shared/service/appServices';
import { AwsVpcStepMapping } from './vpc-step.fieldmapping';
import { AwsField, VpcType } from "../aws-wizard.constants";
import { StepFormDirective } from '../../wizard/shared/step-form/step-form';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { ValidationService } from './../../wizard/shared/validation/validation.service';
import { Vpc } from '../../../../swagger/models/vpc.model';

@Component({
    selector: 'app-vpc-step',
    templateUrl: './vpc-step.component.html',
    styleUrls: ['./vpc-step.component.scss']
})
export class VpcStepComponent extends StepFormDirective implements OnInit {
    defaultVpcHasChanged: boolean = false;
    existingVpcs: Array<Vpc> = [];
    loadingExistingVpcs: boolean = false;
    private curVpcType;
    private curVpcId;

    defaultVpcAddress: string = '10.0.0.0/16';

    constructor(private validationService: ValidationService) {
        super();
    }

    private newVpcTypeIsChangedValue(newVpcType: VpcType): boolean {
        // we want to avoid detecting a change if the new VPC type is '' and the old is undefined
        if (!newVpcType && !this.curVpcType) {
            return false;
        }
        return newVpcType !== this.curVpcType;
    }

    private newVpcIdIsChangedValue(newVpcId: string): boolean {
        // we want to avoid detecting a change if the new VPC id is '' and the old is undefined
        if (!newVpcId && !this.curVpcId) {
            return false;
        }
        return newVpcId !== this.curVpcId;
    }

    private onVpcTypeChange(newVpcType: VpcType) {
        if (!this.newVpcTypeIsChangedValue(newVpcType)) {
            return;
        }
        this.curVpcType = newVpcType;

        if (newVpcType === VpcType.EXISTING) {
            this.clearNewVpcControls();
            if (this.existingVpcs) {
                const possibleVpcIds = this.existingVpcs.map<string>(vpc => vpc.id);
                this.restoreField(AwsField.VPC_EXISTING_ID, AwsVpcStepMapping, possibleVpcIds);
            }
            this.setExistingVpcValidators();
        } else {
            this.clearExistingVpcControls();
            this.setNewVpcValidators();
        }
        AppServices.messenger.publish({
            type: TanzuEventType.AWS_VPC_TYPE_CHANGED,
            payload: { vpcType: newVpcType }
        });
        this.triggerStepDescriptionChange();
    }

    private clearNewVpcControls() {
        this.formGroup.get(AwsField.VPC_NEW_CIDR).clearValidators();
        this.clearControlValue(AwsField.VPC_NEW_CIDR);
        this.clearFieldSavedData(AwsField.VPC_NEW_CIDR);
    }

    private clearExistingVpcControls() {
        const existingVpcControl = this.formGroup.get(AwsField.VPC_EXISTING_ID);
        const existingVpcCidrControl = this.formGroup.get(AwsField.VPC_EXISTING_CIDR);
        existingVpcControl.setValue('');
        existingVpcControl.clearValidators();
        existingVpcControl.updateValueAndValidity();
        existingVpcCidrControl.setValue('');
        existingVpcCidrControl.clearValidators();
        existingVpcCidrControl.updateValueAndValidity();
    }

    ngOnInit() {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, AwsVpcStepMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(AwsVpcStepMapping);
        this.storeDefaultLabels(AwsVpcStepMapping);
        this.registerDefaultFileImportedHandler(this.eventFileImported, AwsVpcStepMapping);
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        // NOTE: we don't call this.registerFieldsAffectingStepDescription() with any fields, because all the relevant fields
        // already trigger a step description change event in their own onChange handlers

        this.registerOnValueChange(AwsField.VPC_TYPE, this.onVpcTypeChange.bind(this));

        const cidrFields = [AwsField.VPC_NEW_CIDR, AwsField.VPC_EXISTING_CIDR];
        cidrFields.forEach(cidrField => {
            this.registerOnValueChange(cidrField, (cidr) => {
                AppServices.messenger.publish({
                    type: TanzuEventType.NETWORK_STEP_GET_NO_PROXY_INFO,
                    payload: { info: (cidr ? cidr + ',' : '') + '169.254.0.0/16' }
                });
                this.triggerStepDescriptionChange();
            });
        });

        /**
         * Whenever aws region selection changes, update AZ subregion
         */
        AppServices.messenger.subscribe(TanzuEventType.AWS_REGION_CHANGED, () => {
                this.existingVpcs = [];
                this.clearControlValue(AwsField.VPC_EXISTING_ID);
                this.clearControlValue(AwsField.VPC_EXISTING_CIDR);
            }, this.unsubscribe);

        AppServices.dataServiceRegistrar.stepSubscribe<Vpc>(this, TanzuEventType.AWS_GET_EXISTING_VPCS, this.onFetchedVpcs.bind(this));

        this.registerOnValueChange(AwsField.VPC_NON_INTERNET_FACING, this.onNonInternetFacingVPCChange.bind(this));
        this.registerOnValueChange(AwsField.VPC_EXISTING_ID, this.onChangeExistingVpc.bind(this));
    }

    protected onStepStarted() {
        this.chooseInitialVpcType();
    }

    private onFetchedVpcs(vpcs: Array<Vpc>) {
        this.existingVpcs = vpcs;
        this.loadingExistingVpcs = false;
    }

    onNonInternetFacingVPCChange(checked: boolean) {
        AppServices.messenger.publish({
            type: TanzuEventType.AWS_AIRGAPPED_VPC_CHANGE,
            payload: checked === true
        });
    }

    chooseInitialVpcType() {
        const vpcType = this.getStoredValue(AwsField.VPC_NEW_CIDR, AwsVpcStepMapping) ? VpcType.NEW : VpcType.EXISTING;
        this.formGroup.get(AwsField.VPC_TYPE).setValue(vpcType);
    }

    /**
     * @method setNewVpcValidators
     * helper method to consolidate setting validators for new vpc fields and
     * re-subscribe to vpc value changes
     */
    setNewVpcValidators() {
        this.defaultVpcHasChanged = false;

        this.setFieldWithStoredValue(AwsField.VPC_NEW_CIDR, AwsVpcStepMapping, this.defaultVpcAddress);
        this.formGroup.get(AwsField.VPC_NEW_CIDR).setValidators([
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpNetworkSegment()
        ]);
    }

    setExistingVpcValidators() {
        if (this.existingVpcs.length > 0) {
            this.formGroup.get(AwsField.VPC_EXISTING_ID).setValidators([Validators.required]);
        }
    }

    /**
     * @method onChangeExistingVpc
     * helper method to manually set existing VPC CIDR read-only value, and
     * dispatch message to retrieve VPC subnets by VPC ID
     * @param existingVpcId
     */
    onChangeExistingVpc(existingVpcId: any) {
        if (!this.newVpcIdIsChangedValue(existingVpcId)) {
            return;
        }
        this.curVpcId = existingVpcId;

        const existingVpc: Array<Vpc> = this.existingVpcs.filter((vpc) => { return vpc.id === existingVpcId; });
        const existingVpcCidr = existingVpc && existingVpc.length > 0 ? existingVpc[0].cidr : '';
        this.setControlValueSafely(AwsField.VPC_EXISTING_CIDR, existingVpcCidr);

        AppServices.messenger.publish({
            type: TanzuEventType.AWS_GET_SUBNETS,
            payload: { vpcId: existingVpcId }
        });

        AppServices.messenger.publish(({
            type: TanzuEventType.AWS_VPC_CHANGED
        }));

        this.triggerStepDescriptionChange();
    }

    dynamicDescription(): string {
        const vpcType = this.getFieldValue(AwsField.VPC_TYPE);
        const vpcExistingCidr = this.getFieldValue(AwsField.VPC_EXISTING_CIDR, true);
        const vpcExistingId = this.getFieldValue(AwsField.VPC_EXISTING_ID, true);
        const vpcNewCidr = this.getFieldValue(AwsField.VPC_NEW_CIDR, true);

        if (vpcType === VpcType.EXISTING && vpcExistingId && vpcExistingCidr) {
            return 'VPC: ' + vpcExistingId + ' CIDR: ' + vpcExistingCidr;
        }
        if (vpcType === VpcType.NEW && vpcNewCidr) {
            return 'VPC: (new) CIDR: ' + vpcNewCidr;
        }
        return 'Specify VPC settings for AWS';
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(AwsVpcStepMapping);
        this.storeDefaultDisplayOrder(AwsVpcStepMapping);
    }
}
