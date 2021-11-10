// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';

import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { AWSAccountParamsKeys } from './provider-step/aws-provider-step.component';
import { AWSRegionalClusterParams, AWSVpc } from 'src/app/swagger/models';
import { APIClient } from 'src/app/swagger';
import { AwsWizardFormService } from 'src/app/shared/service/aws-wizard-form.service';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import Broker from "../../../shared/service/broker";

enum AwsForm {
    PROVIDER = 'awsProviderForm',
    VPC = 'vpcForm',
    NODESETTING = 'awsNodeSettingForm',
    NETWORK = 'networkForm',
    METADATA = 'metadataForm',
    IDENTITY = 'identityForm',
    OSIMAGE = 'osImageForm'
}

enum AwsStep {
    PROVIDER = 'provider',
    VPC = 'vpc',
    NODESETTING = 'nodeSetting',
    NETWORK = 'network',
    METADATA = 'metadata',
    IDENTITY = 'identity',
    OSIMAGE = 'osImage'
}

@Component({
    selector: 'aws-wizard',
    templateUrl: './aws-wizard.component.html',
    styleUrls: ['./aws-wizard.component.scss'],
})
export class AwsWizardComponent extends WizardBaseDirective implements OnInit {

    // The region user selected
    region: string;
    nodeAzList: Array<any>;

    constructor(
        router: Router,
        public wizardFormService: AwsWizardFormService,
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

    getStepDescription(stepName: string): string {
        if (stepName === AwsStep.PROVIDER) {
            return 'Validate the AWS provider account for Tanzu';
        } else if (stepName === AwsStep.VPC) {
            const vpc = this.getFieldValue(AwsForm.VPC, 'vpc');
            const publicNodeCidr = this.getFieldValue(AwsForm.VPC, 'publicNodeCidr');
            const privateNodeCidr = this.getFieldValue(AwsForm.VPC, 'privateNodeCidr');
            const awsNodeAz = this.getFieldValue(AwsForm.VPC, 'awsNodeAz');

            if (vpc && publicNodeCidr && privateNodeCidr && awsNodeAz) {
                return `VPC CIDR: ${vpc}, Public Node CIDR: ${publicNodeCidr}, ` +
                    `Private Node CIDR: ${privateNodeCidr}, Node AZ: ${awsNodeAz}`;
            } else {
                return 'Specify VPC settings for AWS';
            }
        } else if (stepName === AwsStep.NODESETTING) {
            if (this.getFieldValue(AwsForm.NODESETTING, 'controlPlaneSetting')) {
                let mode = 'Development cluster selected: 1 node control plane';
                if (this.getFieldValue(AwsForm.NODESETTING, 'controlPlaneSetting') === 'prod') {
                    mode = 'Production cluster selected: 3 node control plane';
                }
                return mode;
            } else {
                return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
            }
        } else if (stepName === AwsStep.NETWORK) {
            if (this.getFieldValue(AwsForm.NETWORK, 'clusterPodCidr')) {
                return 'Cluster Pod CIDR: ' + this.getFieldValue(AwsForm.NETWORK, 'clusterPodCidr');
            } else {
                return 'Specify the cluster Pod CIDR';
            }
        } else if (stepName === AwsStep.METADATA) {
            if (this.form.get(AwsForm.METADATA).get('clusterLocation') &&
                this.form.get(AwsForm.METADATA).get('clusterLocation').value) {
                return 'Location: ' + this.form.get(AwsForm.METADATA).get('clusterLocation').value;
            } else {
                return `Specify metadata for the ${this.clusterTypeDescriptor} cluster`;
            }
        } else if (stepName === AwsStep.IDENTITY) {
            if (this.getFieldValue(AwsForm.IDENTITY, 'identityType') === 'oidc' &&
                this.getFieldValue(AwsForm.IDENTITY, 'issuerURL')) {
                return 'OIDC configured: ' + this.getFieldValue(AwsForm.IDENTITY, 'issuerURL')
            } else if (this.getFieldValue(AwsForm.IDENTITY, 'identityType') === 'ldap' &&
                this.getFieldValue(AwsForm.IDENTITY, 'endpointIp')) {
                return 'LDAP configured: ' + this.getFieldValue(AwsForm.IDENTITY, 'endpointIp') + ':' +
                    this.getFieldValue(AwsForm.IDENTITY, 'endpointPort');
            } else {
                return 'Specify identity management'
            }
        } else if (stepName === AwsStep.OSIMAGE) {
            if (this.getFieldValue(AwsForm.OSIMAGE, 'osImage') && this.getFieldValue(AwsForm.OSIMAGE, 'osImage').name) {
                return 'OS Image: ' + this.getFieldValue(AwsForm.OSIMAGE, 'osImage').name;
            } else {
                return 'Specify the OS Image';
            }
        }
    }

    getPayload(): AWSRegionalClusterParams {
        const payload: AWSRegionalClusterParams = {};

        payload.awsAccountParams = {};
        AWSAccountParamsKeys.forEach(key => {
            payload.awsAccountParams[key] = this.getFieldValue(AwsForm.PROVIDER, key);
        });
        payload.loadbalancerSchemeInternal = this.getBooleanFieldValue('vpcForm', 'nonInternetFacingVPC');
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
        // SHIMON TODO:         payload.loadbalancerSchemeInternal = this.getBooleanFieldValue('vpcForm', 'nonInternetFacingVPC');
        if (payload !== undefined) {
            if (payload.awsAccountParams !== undefined) {
                for (const key of Object.keys(payload.awsAccountParams)) {
                    this.saveFormField(AwsForm.PROVIDER, key, payload.awsAccountParams[key]);
                }
            }
            this.saveFormField(AwsForm.NODESETTING, 'sshKeyName', payload.sshKeyName);
            this.saveFormField(AwsForm.NODESETTING, 'createCloudFormation', payload.createCloudFormationStack);
            this.saveFormField(AwsForm.NODESETTING, 'clusterName', payload.clusterName);

            this.saveControlPlaneFlavor('aws', payload.controlPlaneFlavor);
            this.saveControlPlaneNodeType('aws', payload.controlPlaneFlavor, payload.controlPlaneNodeType);
            const bastionHost = payload.bastionHostEnabled ? BASTION_HOST_ENABLED : BASTION_HOST_DISABLED;
            this.saveFormField(AwsForm.NODESETTING, 'bastionHostEnabled', bastionHost);
            this.saveFormField(AwsForm.NODESETTING, 'machineHealthChecksEnabled', payload.machineHealthCheckEnabled);
            this.saveFormField(AwsForm.VPC, 'existingVpcId', (payload.vpc) ? payload.vpc.vpcID : '');
            this.saveFormField("awsNodeSettingForm", "enableAuditLogging", payload.enableAuditLogging);
            this.saveVpcFields(payload.vpc);

            this.saveCommonFieldsFromPayload(payload);
        }
    }

    private saveVpcFields(vpc: AWSVpc) {
        if (vpc) {
            // SHIMON SEZ: verify that this is a good way to determine NEW or EXISTING?
            // we check the first node to determine whether we're dealing with a NEW or EXISTING
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
                const node = this.nodeAzList[x];
                // we set the UI fields highest to lowest because we collected them into the payload
                // by pushing them on to the array, which essentially reversed their order. So if there
                // are 3 AZs, for example, the 3rd one in the UI will have landed as the 0-index in our array.
                const uiIndex = numNodeAz - x;
                this.saveAzNodeFields(node, uiIndex);
            }
        }
    }

    private saveAzNodeFields(node: any, uiIndex: number) {
        this.saveFormField(AwsForm.NODESETTING, `awsNodeAz${uiIndex}`, node.name);
        if (!Broker.appDataService.isModeClusterStandalone()) {
            this.saveFormField(AwsForm.NODESETTING, `workerNodeInstanceType${uiIndex}`, node.workerNodeType);
        }
        const publicSubnetId = node.publicSubnetID === null || node.publicSubnetID.length === 0 ? '' : node.publicSubnetID;
        this.saveFormField(AwsForm.NODESETTING, `vpcPublicSubnet${uiIndex}`, publicSubnetId);
        const privateSubnetId = node.privateSubnetId === null || node.privateSubnetId.length === 0 ? '' : node.privateSubnetId;
        this.saveFormField(AwsForm.NODESETTING, `vpcPrivateSubnet${uiIndex}`, privateSubnetId);
    }

    getAwsNodeAzs(payload) {
        this.nodeAzList = [
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
            this.nodeAzList.push({
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
            this.nodeAzList.push({
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

        return this.nodeAzList;
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

    retrievePayloadFromString(config: string): Observable<any> {
        return this.apiClient.importTKGConfigForAWS( { params: { filecontents: config } } );
    }

    validateImportFile(config: string): string {
        if (config.includes('AWS_')) {
            return '';
        }
        return 'This file is not an AWS configuration file!';
    }
}
