// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';

import { AWSNodeAz, AWSRegionalClusterParams, AWSVpc } from 'src/app/swagger/models';
import { APIClient } from 'src/app/swagger';
import { AwsWizardFormService } from 'src/app/shared/service/aws-wizard-form.service';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { APIClient } from 'src/app/swagger';
import { AWSRegionalClusterParams } from 'src/app/swagger/models';
import Broker from "../../../shared/service/broker";
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { BASTION_HOST_DISABLED, BASTION_HOST_ENABLED } from './node-setting-step/node-setting-step.component';
import { AWSAccountParamsKeys } from './provider-step/aws-provider-step.component';
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';
import { StepUtility } from '../wizard/shared/components/steps/step-utility';
import { AwsField, AwsForm, AwsStep } from "./aws-wizard.constants";
import { ImportParams, ImportService } from "../../../shared/service/import.service";
import { Utils } from '../../../shared/utils';
import { InstanceType } from '../../../shared/constants/app.constants';

@Component({
    selector: 'aws-wizard',
    templateUrl: './aws-wizard.component.html',
    styleUrls: ['./aws-wizard.component.scss'],
})
export class AwsWizardComponent extends WizardBaseDirective implements OnInit {
    constructor(
        router: Router,
        public wizardFormService: AwsWizardFormService,
        private importService: ImportService,
        private formBuilder: FormBuilder,
        private apiClient: APIClient,
        formMetaDataService: FormMetaDataService,
        titleService: Title,
        el: ElementRef) {

        super(router, el, formMetaDataService, titleService);

        this.form = this.formBuilder.group({
            awsProviderForm: this.formBuilder.group({}),
            vpcForm: this.formBuilder.group({}),
            awsNodeSettingForm: this.formBuilder.group({}),
            metadataForm: this.formBuilder.group({}),
            networkForm: this.formBuilder.group({}),
            identityForm: this.formBuilder.group({}),
            ceipOptInForm: this.formBuilder.group({}),
            osImageForm: this.formBuilder.group({})
        });
    }

    ngOnInit() {
        super.ngOnInit();

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
                this.saveAzNodeFields(nodeAzList[x], x + 1);
            }
        }
    }

    private saveAzNodeFields(node: AWSNodeAz, uiIndex: number) {
        // TODO: move away from identifying the fields with ${uiIndex} and use an enum field identifier
        this.saveFormField(AwsForm.NODESETTING, `awsNodeAz${uiIndex}`, node.name);
        if (!Broker.appDataService.isModeClusterStandalone()) {
            this.saveFormField(AwsForm.NODESETTING, `workerNodeInstanceType${uiIndex}`, node.workerNodeType);
        }
        this.saveFormField(AwsForm.NODESETTING, `vpcPublicSubnet${uiIndex}`, Utils.safeString(node.publicSubnetID));
        this.saveFormField(AwsForm.NODESETTING, `vpcPrivateSubnet${uiIndex}`, Utils.safeString(node.privateSubnetID));
    }

    getAwsNodeAzs(payload) {
        // TODO: move away from identifying the fields with literals and use an enum field identifier
        const nodeAzList = [
            {
                name: this.getFieldValue(AwsForm.NODESETTING, 'awsNodeAz1'),
                workerNodeType: Broker.appDataService.isModeClusterStandalone() ? payload.controlPlaneNodeType :
                    this.getFieldValue(AwsForm.NODESETTING, 'workerNodeInstanceType1'),
                publicNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                    this.getFieldValue(AwsForm.VPC, 'publicNodeCidr') : '',
                privateNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                    this.getFieldValue(AwsForm.VPC, 'privateNodeCidr') : '',
                publicSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                    this.getFieldValue(AwsForm.NODESETTING, 'vpcPublicSubnet1') : '',
                privateSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                    this.getFieldValue(AwsForm.NODESETTING, 'vpcPrivateSubnet1') : ''
            }
        ];

        if (this.getFieldValue(AwsForm.NODESETTING, 'awsNodeAz2')) {
            nodeAzList.push({
                name: this.getFieldValue(AwsForm.NODESETTING, 'awsNodeAz2'),
                workerNodeType: (!Broker.appDataService.isModeClusterStandalone()) ?
                    this.getFieldValue(AwsForm.NODESETTING, 'workerNodeInstanceType2') : payload.controlPlaneNodeType,
                publicNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                    this.getFieldValue(AwsForm.VPC, 'publicNodeCidr') : '',
                privateNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                    this.getFieldValue(AwsForm.VPC, 'privateNodeCidr') : '',
                publicSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                    this.getFieldValue(AwsForm.NODESETTING, 'vpcPublicSubnet2') : '',
                privateSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                    this.getFieldValue(AwsForm.NODESETTING, 'vpcPrivateSubnet2') : ''
            });
        }

        if (this.getFieldValue(AwsForm.NODESETTING, 'awsNodeAz3')) {
            nodeAzList.push({
                name: this.getFieldValue(AwsForm.NODESETTING, 'awsNodeAz3'),
                workerNodeType: (!Broker.appDataService.isModeClusterStandalone()) ?
                    this.getFieldValue(AwsForm.NODESETTING, 'workerNodeInstanceType3') : payload.controlPlaneNodeType,
                publicNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                    this.getFieldValue(AwsForm.VPC, 'publicNodeCidr') : '',
                privateNodeCidr: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'new') ?
                    this.getFieldValue(AwsForm.VPC, 'privateNodeCidr') : '',
                publicSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                    this.getFieldValue(AwsForm.NODESETTING, 'vpcPublicSubnet3') : '',
                privateSubnetID: (this.getFieldValue(AwsForm.VPC, 'vpcType') === 'existing') ?
                    this.getFieldValue(AwsForm.NODESETTING, 'vpcPrivateSubnet3') : ''
            });
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
    }}
    // HTML convenience methods
    //
    get AwsProviderForm(): FormDataForHTML {
        return {name: 'awsProviderForm', title: 'IaaS Provider', description: this.AwsProviderFormDescription,
            i18n: {title: 'IaaS provder step name', description: 'IaaS provder step description'}};
    }
    private get AwsProviderFormDescription(): string {
        return 'Validate the AWS provider account for ' + this.title;
    }
    get AwsNodeSettingForm(): FormDataForHTML {
        return { name: 'awsNodeSettings', title: FormUtility.titleCase(this.clusterTypeDescriptor) + ' Cluster Settings',
            description: this.AwsNodeSettingFormDescription,
            i18n: {title: 'IaaS provder step name', description: 'IaaS provder step description'} };
    }
    private get AwsNodeSettingFormDescription(): string {
        if (this.getFieldValue('awsNodeSettingForm', 'controlPlaneSetting')) {
            let mode = 'Development cluster selected: 1 node control plane';
            if (this.getFieldValue('awsNodeSettingForm', 'controlPlaneSetting') === 'prod') {
                mode = 'Production cluster selected: 3 node control plane';
            }
            return mode;
        }
        return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
    }
    get AwsVpcForm(): FormDataForHTML {
        return {name: 'vpcForm', title: 'VPC for AWS', description: this.AwsVpcFormDescription,
        i18n: {title: 'vpc step name', description: 'vpc step description'}};
    }
    private get AwsVpcFormDescription(): string {
        const vpc = this.getFieldValue('vpcForm', 'vpc');
        const publicNodeCidr = this.getFieldValue('vpcForm', 'publicNodeCidr');
        const privateNodeCidr = this.getFieldValue('vpcForm', 'privateNodeCidr');
        const awsNodeAz = this.getFieldValue('vpcForm', 'awsNodeAz');

        if (vpc && publicNodeCidr && privateNodeCidr && awsNodeAz) {
            return `VPC CIDR: ${vpc}, Public Node CIDR: ${publicNodeCidr}, ` +
                `Private Node CIDR: ${privateNodeCidr}, Node AZ: ${awsNodeAz}`;
        }
        return 'Specify VPC settings for AWS';
    }
    //
    // HTML convenience methods
}
