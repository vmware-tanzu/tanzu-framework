// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
// Third party imports
import { Observable } from 'rxjs';
// App imports
import { APIClient } from 'src/app/swagger';
import AppServices from "../../../shared/service/appServices";
import {
    AWSAvailabilityZone,
    AWSNodeAz,
    AWSRegionalClusterParams,
    AWSSubnet,
    AWSVirtualMachine,
    AWSVpc,
    Vpc
} from 'src/app/swagger/models';
import { AWSAccountParamsKeys, AwsProviderStepComponent } from './provider-step/aws-provider-step.component';
import { AwsField, AwsForm, AwsStep } from "./aws-wizard.constants";
import { AwsOsImageStepComponent } from './os-image-step/aws-os-image-step.component';
import { BASTION_HOST_DISABLED, BASTION_HOST_ENABLED, NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { ImportParams, ImportService } from "../../../shared/service/import.service";
import { InstanceType } from '../../../shared/constants/app.constants';
import { TkgEventType } from '../../../shared/service/Messenger';
import { Utils } from '../../../shared/utils';
import { VpcStepComponent } from './vpc-step/vpc-step.component';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';

export interface AzRelatedFields {
    az: string,
    workerNodeInstanceType: string,
    vpcPublicSubnet: string,
    vpcPrivateSubnet: string
}

export const AzRelatedFieldsArray: AzRelatedFields[] = [
    { az: AwsField.NODESETTING_AZ_1, vpcPrivateSubnet: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_1,
        vpcPublicSubnet: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1, workerNodeInstanceType: AwsField.NODESETTING_WORKERTYPE_1 },
    { az: AwsField.NODESETTING_AZ_2, vpcPrivateSubnet: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_2,
        vpcPublicSubnet: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_1, workerNodeInstanceType: AwsField.NODESETTING_WORKERTYPE_1 },
    { az: AwsField.NODESETTING_AZ_3, vpcPrivateSubnet: AwsField.NODESETTING_VPC_PRIVATE_SUBNET_3,
        vpcPublicSubnet: AwsField.NODESETTING_VPC_PUBLIC_SUBNET_3, workerNodeInstanceType: AwsField.NODESETTING_WORKERTYPE_3 },
];

@Component({
    selector: 'aws-wizard',
    templateUrl: './aws-wizard.component.html',
    styleUrls: ['./aws-wizard.component.scss'],
})
export class AwsWizardComponent extends WizardBaseDirective implements OnInit {
    constructor(
        router: Router,
        formBuilder: FormBuilder,
        private importService: ImportService,
        private apiClient: APIClient,
        formMetaDataService: FormMetaDataService,
        titleService: Title,
        el: ElementRef) {

        super(router, el, formMetaDataService, titleService, formBuilder);
    }

    protected supplyStepData(): FormDataForHTML[] {
        return [
            this.AwsProviderForm,
            this.AwsVpcForm,
            this.AwsNodeSettingForm,
            this.MetadataForm,
            this.AwsNetworkForm,
            this.IdentityForm,
            this.AwsOsImageForm,
            this.CeipForm,
        ];
    }

    ngOnInit() {
        super.ngOnInit();
        this.registerServices();
        this.subscribeToServices();

        // To avoid re-open issue for AWS provider step.
        this.form.markAsDirty();

        this.titleService.setTitle(this.title + ' AWS');
    }

    getPayload(): AWSRegionalClusterParams {
        const payload: AWSRegionalClusterParams = {};

        payload.awsAccountParams = {};
        AWSAccountParamsKeys.forEach(key => {
            payload.awsAccountParams[key] = this.getFieldValue(AwsForm.PROVIDER, key);
        });
        payload.loadbalancerSchemeInternal = this.getBooleanFieldValue(AwsForm.VPC, 'nonInternetFacingVPC');
        payload.sshKeyName = this.getFieldValue(AwsForm.NODESETTING, 'sshKeyName');
        payload.createCloudFormationStack = this.getFieldValue(AwsForm.NODESETTING, 'createCloudFormation') || false;
        payload.clusterName = this.getFieldValue(AwsForm.NODESETTING, 'clusterName');
        payload.controlPlaneNodeType = this.getControlPlaneNodeType('aws');
        payload.controlPlaneFlavor = this.getControlPlaneFlavor('aws');
        const bastionHostEnabled = this.getFieldValue(AwsForm.NODESETTING, 'bastionHostEnabled');
        payload.bastionHostEnabled = bastionHostEnabled === BASTION_HOST_ENABLED;
        const machineHealthChecksEnabled = this.getFieldValue(AwsForm.NODESETTING, 'machineHealthChecksEnabled');
        payload.machineHealthCheckEnabled = (machineHealthChecksEnabled === true);
        payload.vpc = {
            cidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                this.getFieldValue(AwsForm.VPC, 'existingVpcCidr') :
                this.getFieldValue(AwsForm.VPC, 'vpc'),
            vpcID: this.getFieldValue(AwsForm.VPC, 'existingVpcId'),
            azs: this.getAwsNodeAzs(payload)
        };

        payload.enableAuditLogging = this.getBooleanFieldValue("awsNodeSettingForm", "enableAuditLogging");
        this.initPayloadWithCommons(payload);

        return payload;
    }

    setFromPayload(payload: AWSRegionalClusterParams) {
        if (payload !== undefined) {
            if (payload.awsAccountParams !== undefined) {
                for (const key of Object.keys(payload.awsAccountParams)) {
                    this.saveFormField(AwsForm.PROVIDER, key, payload.awsAccountParams[key]);
                }
            }
            this.saveFormField(AwsForm.NODESETTING, 'sshKeyName', payload.sshKeyName);
            this.saveFormField(AwsForm.NODESETTING, 'createCloudFormation', payload.createCloudFormationStack);
            this.saveFormField(AwsForm.NODESETTING, 'clusterName', payload.clusterName);

            this.saveFormField(AwsForm.NODESETTING, AwsField.NODESETTING_CONTROL_PLANE_SETTING, payload.controlPlaneFlavor);
            if (payload.controlPlaneFlavor === InstanceType.DEV) {
                this.saveFormField(AwsForm.NODESETTING, AwsField.NODESETTING_INSTANCE_TYPE_DEV, payload.controlPlaneNodeType);
            } else if (payload.controlPlaneFlavor === InstanceType.PROD) {
                this.saveFormField(AwsForm.NODESETTING, AwsField.NODESETTING_INSTANCE_TYPE_PROD, payload.controlPlaneNodeType);
            }
            const bastionHost = payload.bastionHostEnabled ? BASTION_HOST_ENABLED : BASTION_HOST_DISABLED;
            this.saveFormField(AwsForm.NODESETTING, 'bastionHostEnabled', bastionHost);
            this.saveFormField(AwsForm.NODESETTING, 'machineHealthChecksEnabled', payload.machineHealthCheckEnabled);
            this.saveFormField(AwsForm.VPC, 'existingVpcId', (payload.vpc) ? payload.vpc.vpcID : '');
            this.saveFormField(AwsForm.VPC, 'nonInternetFacingVPC', payload.loadbalancerSchemeInternal)
            this.saveFormField(AwsForm.NODESETTING, "enableAuditLogging", payload.enableAuditLogging);
            this.saveVpcFields(payload.vpc);

            this.saveCommonFieldsFromPayload(payload);
        }
    }

    private saveVpcFields(vpc: AWSVpc) {
        if (vpc) {
            if (vpc.vpcID) {
                this.saveFormField(AwsForm.VPC, 'vpcType', 'existing');
                this.saveFormField(AwsForm.VPC, 'existingVpcCidr', vpc.cidr);
                this.saveFormField(AwsForm.VPC, 'publicNodeCidr', '');
                this.saveFormField(AwsForm.VPC, 'privateNodeCidr', '');
                this.saveFormField(AwsForm.VPC, 'existingVpcId', vpc.vpcID);
            } else {
                this.saveFormField(AwsForm.VPC, 'vpcType', 'new');
                this.saveFormField(AwsForm.VPC, 'vpc', vpc.cidr);
            }
            this.saveVpcAzs(vpc);
        }
    }

    private saveVpcAzs(vpc: AWSVpc) {
        if (vpc && vpc.azs && vpc.azs.length > 0) {
            const nodeAzList = vpc.azs;
            const numNodeAz = nodeAzList.length;
            for (let x = 0; x < numNodeAz; x++) {
                this.saveAzNodeFields(nodeAzList[x], AzRelatedFieldsArray[x]);
            }
        }
    }

    private saveAzNodeFields(node: AWSNodeAz, azFields: AzRelatedFields) {
        this.saveFormField(AwsForm.NODESETTING, azFields.az, node.name);
        if (!AppServices.appDataService.isModeClusterStandalone()) {
            this.saveFormField(AwsForm.NODESETTING, azFields.workerNodeInstanceType, node.workerNodeType);
        }
        this.saveFormField(AwsForm.NODESETTING, azFields.vpcPublicSubnet, Utils.safeString(node.publicSubnetID));
        this.saveFormField(AwsForm.NODESETTING, azFields.vpcPrivateSubnet, Utils.safeString(node.privateSubnetID));
    }

    private getAzFieldData(azFields: AzRelatedFields, standaloneControlPlaneNodeType: string) {
        return             {
            name: this.getFieldValue(AwsForm.NODESETTING, azFields.az),
            workerNodeType: AppServices.appDataService.isModeClusterStandalone() ? standaloneControlPlaneNodeType :
                this.getFieldValue(AwsForm.NODESETTING, azFields.workerNodeInstanceType),
            publicNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                this.getFieldValue(AwsForm.VPC, 'publicNodeCidr') : '',
            privateNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                this.getFieldValue(AwsForm.VPC, 'privateNodeCidr') : '',
            publicSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                this.getFieldValue(AwsForm.NODESETTING, azFields.vpcPublicSubnet) : '',
            privateSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                this.getFieldValue(AwsForm.NODESETTING, azFields.vpcPrivateSubnet) : ''
        }
    }

    getAwsNodeAzs(payload) {
        const nodeAzList = [this.getAzFieldData(AzRelatedFieldsArray[0], payload.controlPlaneNodeType)];

        if (this.getFieldValue(AwsForm.NODESETTING, AwsField.NODESETTING_AZ_2)) {
            nodeAzList.push(this.getAzFieldData(AzRelatedFieldsArray[1], payload.controlPlaneNodeType));
        }
        if (this.getFieldValue(AwsForm.NODESETTING, AwsField.NODESETTING_AZ_3)) {
            nodeAzList.push(this.getAzFieldData(AzRelatedFieldsArray[2], payload.controlPlaneNodeType));
        }
        return nodeAzList;
    }

    /**
     * @method method to trigger deployment
     */
    createRegionalCluster(payload: any): Observable<any> {
        return this.apiClient.createAWSRegionalCluster(payload);
    }

    /**
     * Return management/standalone cluster name
     */
    getMCName() {
        return this.getFieldValue(AwsForm.NODESETTING, 'clusterName');
    }

    /**
     * @method getExtendCliCmds to return cli command string according to special selection
     * For AWS, selects Create Cloudformation Stack,
     * should include tanzu management-cluster permissions aws set
     * @returns the array includes cli command object like {isPrefixOfCreateCmd: true, cmdStr: "tanzu ..."}
     */
    getExtendCliCmds(): Array<{ isPrefixOfCreateCmd: boolean, cmdStr: string }> {
        if (this.getFieldValue(AwsForm.NODESETTING, 'createCloudFormation')) {
            const clusterPrefix = (this.getClusterType()) ? this.getClusterType() : 'management';
            const command = `tanzu ${clusterPrefix}-cluster permissions aws set`;
            return [{ isPrefixOfCreateCmd: true, cmdStr: command }]
        }
        return []
    }

    /**
     * Get the CLI used to deploy the management/standalone cluster
     */
    getCli(configPath: string): string {
        const cliG = new CliGenerator();
        const cliParams: CliFields = {
            configPath: configPath,
            clusterType: this.getClusterType(),
            clusterName: this.getMCName(),
            extendCliCmds: this.getExtendCliCmds()
        };
        return cliG.getCli(cliParams);
    }

    applyTkgConfig() {
        return this.apiClient.applyTKGConfigForAWS({ params: this.getPayload() });
    }

    /**
     * Retrieve the config file from the backend and return as a string
     */
    retrieveExportFile() {
        return this.apiClient.exportTKGConfigForAWS({ params: this.getPayload() });
    }

    // returns TRUE if the file contents appear to be a valid config file for AWS
    // returns FALSE if the file is empty or does not appear to be valid. Note that in the FALSE
    // case we also alert the user.
    importFileValidate(nameFile: string, fileContents: string): boolean {
        if (fileContents.includes('AWS_')) {
            return true;
        }
        alert(nameFile + ' is not a valid AWS configuration file!');
        return false;
    }

    importFileRetrieveClusterParams(fileContents: string): Observable<AWSRegionalClusterParams>  {
        return this.apiClient.importTKGConfigForAWS( { params: { filecontents: fileContents } } );
    }

    importFileProcessClusterParams(nameFile: string, awsClusterParams: AWSRegionalClusterParams) {
        this.setFromPayload(awsClusterParams);
        this.resetToFirstStep();
        this.importService.publishImportSuccess(nameFile);
    }

    // returns TRUE if user (a) will not lose data on import, or (b) confirms it's OK
    onImportButtonClick() {
        let result = true;
        if (!this.isOnFirstStep()) {
            result = confirm('Importing will overwrite any data you have entered. Proceed with import?');
        }
        return result;
    }

    onImportFileSelected(event) {
        const params: ImportParams<AWSRegionalClusterParams> = {
            file: event.target.files[0],
            validator: this.importFileValidate,
            backend: this.importFileRetrieveClusterParams.bind(this),
            onSuccess: this.importFileProcessClusterParams.bind(this),
            onFailure: this.importService.publishImportFailure
        }
        this.importService.import(params);

        // clear file reader target so user can re-select same file if needed
        event.target.value = '';
    }
    // HTML convenience methods
    //
    get AwsProviderForm(): FormDataForHTML {
        return {name: 'awsProviderForm', title: 'IaaS Provider',
            description: 'Validate the AWS provider account for ' + this.title,
            i18n: {title: 'IaaS provder step name', description: 'IaaS provder step description'},
        clazz: AwsProviderStepComponent};
    }
    get AwsNodeSettingForm(): FormDataForHTML {
        return { name: AwsForm.NODESETTING, title: FormUtility.titleCase(this.clusterTypeDescriptor) + ' Cluster Settings',
            description: `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`,
            i18n: {title: 'IaaS provder step name', description: 'IaaS provder step description'},
        clazz: NodeSettingStepComponent};
    }
    get AwsVpcForm(): FormDataForHTML {
        return {name: 'vpcForm', title: 'VPC for AWS', description: 'Specify VPC settings for AWS',
        i18n: {title: 'vpc step name', description: 'vpc step description'},
        clazz: VpcStepComponent};
    }
    get AwsOsImageForm(): FormDataForHTML {
        return this.getOsImageForm(AwsOsImageStepComponent);
    }
    get AwsNetworkForm(): FormDataForHTML {
        return FormUtility.formOverrideDescription(this.NetworkForm, 'Specify the cluster Pod CIDR');
    }
    //
    // HTML convenience methods

    private subscribeToServices() {
        AppServices.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .subscribe(event => {
                const region = event.payload;
                AppServices.dataServiceRegistrar.trigger([TkgEventType.AWS_GET_OS_IMAGES], {region: region});
                // NOTE: even though the VPC and AZ endpoints don't take the region as a payload, they DO return different data
                // if the user logs in to AWS using a different region. Therefore, we re-fetch that data if the region changes.
                AppServices.dataServiceRegistrar.trigger([TkgEventType.AWS_GET_EXISTING_VPCS, TkgEventType.AWS_GET_AVAILABILITY_ZONES]);
            });
    }

    private registerServices() {
        const wizard = this;
        AppServices.dataServiceRegistrar.register<Vpc>(TkgEventType.AWS_GET_EXISTING_VPCS,
            () => { return wizard.apiClient.getVPCs() },
            "Failed to retrieve list of existing VPCs from the specified AWS Account." );
        AppServices.dataServiceRegistrar.register<AWSAvailabilityZone>(TkgEventType.AWS_GET_AVAILABILITY_ZONES,
            () => { return wizard.apiClient.getAWSAvailabilityZones(); },
            "Failed to retrieve list of availability zones from the specified AWS Account." );
        AppServices.dataServiceRegistrar.register<AWSSubnet>(TkgEventType.AWS_GET_SUBNETS,
            (payload: { vpcId: string }) => {return wizard.apiClient.getAWSSubnets(payload)},
            "Failed to retrieve list of VPC subnets from the specified AWS Account." );
        AppServices.dataServiceRegistrar.register<string>(TkgEventType.AWS_GET_NODE_TYPES,
            (payload: {az?: string}) => { return wizard.apiClient.getAWSNodeTypes(payload); },
            "Failed to retrieve list of node types from the specified AWS Account." );
        AppServices.dataServiceRegistrar.register<AWSVirtualMachine>(TkgEventType.AWS_GET_OS_IMAGES,
            (payload: {region: string}) => { return wizard.apiClient.getAWSOSImages(payload); },
            "Failed to retrieve list of OS images from the specified AWS Server." );
    }
}
