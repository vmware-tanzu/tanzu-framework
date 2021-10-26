// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

import { Observable } from 'rxjs';

import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { AWSAccountParamsKeys } from './provider-step/aws-provider-step.component';
import { AWSRegionalClusterParams } from 'src/app/swagger/models';
import { APIClient } from 'src/app/swagger';
import { AwsWizardFormService } from 'src/app/shared/service/aws-wizard-form.service';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { BASTION_HOST_ENABLED } from './node-setting-step/node-setting-step.component';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import Broker from "../../../shared/service/broker";

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
            awsProviderForm: this.formBuilder.group({
            }),
            vpcForm: this.formBuilder.group({
            }),
            awsNodeSettingForm: this.formBuilder.group({
            }),
            metadataForm: this.formBuilder.group({
            }),
            networkForm: this.formBuilder.group({
            }),
            identityForm: this.formBuilder.group({
            }),
            ceipOptInForm: this.formBuilder.group({
            }),
            osImageForm: this.formBuilder.group({
            })
        });
    }

    ngOnInit() {
        super.ngOnInit();

        // To avoid re-open issue for AWS provider step.
        this.form.markAsDirty();

        this.titleService.setTitle(this.title + ' AWS');
    }

    getStepDescription(stepName: string): string {
        if (stepName === 'provider') {
            return 'Validate the AWS provider account for Tanzu';
        } else if (stepName === 'vpc') {
            const vpc = this.getFieldValue('vpcForm', 'vpc');
            const publicNodeCidr = this.getFieldValue('vpcForm', 'publicNodeCidr');
            const privateNodeCidr = this.getFieldValue('vpcForm', 'privateNodeCidr');
            const awsNodeAz = this.getFieldValue('vpcForm', 'awsNodeAz');

            if (vpc && publicNodeCidr && privateNodeCidr && awsNodeAz) {
                return `VPC CIDR: ${vpc}, Public Node CIDR: ${publicNodeCidr}, ` +
                    `Private Node CIDR: ${privateNodeCidr}, Node AZ: ${awsNodeAz}`;
            } else {
                return 'Specify VPC settings for AWS';
            }
        } else if (stepName === 'nodeSetting') {
            if (this.getFieldValue('awsNodeSettingForm', 'controlPlaneSetting')) {
                let mode = 'Development cluster selected: 1 node control plane';
                if (this.getFieldValue('awsNodeSettingForm', 'controlPlaneSetting') === 'prod') {
                    mode = 'Production cluster selected: 3 node control plane';
                }
                return mode;
            } else {
                return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
            }
        } else if (stepName === 'network') {
            if (this.getFieldValue('networkForm', 'clusterPodCidr')) {
                return 'Cluster Pod CIDR: ' + this.getFieldValue('networkForm', 'clusterPodCidr');
            } else {
                return 'Specify the cluster Pod CIDR';
            }
        } else if (stepName === 'metadata') {
            if (this.form.get('metadataForm').get('clusterLocation') &&
                this.form.get('metadataForm').get('clusterLocation').value) {
                return 'Location: ' + this.form.get('metadataForm').get('clusterLocation').value;
            } else {
                return `Specify metadata for the ${this.clusterTypeDescriptor} cluster`;
            }
        } else if (stepName === 'identity') {
            if (this.getFieldValue('identityForm', 'identityType') === 'oidc' &&
                this.getFieldValue('identityForm', 'issuerURL')) {
                return 'OIDC configured: ' + this.getFieldValue('identityForm', 'issuerURL')
            } else if (this.getFieldValue('identityForm', 'identityType') === 'ldap' &&
                this.getFieldValue('identityForm', 'endpointIp')) {
                return 'LDAP configured: ' + this.getFieldValue('identityForm', 'endpointIp') + ':' +
                    this.getFieldValue('identityForm', 'endpointPort');
            } else {
                return 'Specify identity management'
            }
        } else if (stepName === 'osImage') {
            if (this.getFieldValue('osImageForm', 'osImage') && this.getFieldValue('osImageForm', 'osImage').name) {
                return 'OS Image: ' + this.getFieldValue('osImageForm', 'osImage').name;
            } else {
                return 'Specify the OS Image';
            }
        }
    }

    getPayload(): AWSRegionalClusterParams {
        const payload: AWSRegionalClusterParams = {};

        payload.awsAccountParams = {};
        AWSAccountParamsKeys.forEach(key => {
            payload.awsAccountParams[key] = this.getFieldValue('awsProviderForm', key);
        });

        payload.sshKeyName = this.getFieldValue('awsNodeSettingForm', 'sshKeyName');
        payload.createCloudFormationStack = this.getFieldValue('awsNodeSettingForm', 'createCloudFormation') || false;
        payload.clusterName = this.getFieldValue('awsNodeSettingForm', 'clusterName');
        payload.controlPlaneNodeType = this.getControlPlaneNodeType('aws');
        payload.controlPlaneFlavor = this.getControlPlaneFlavor('aws');
        const bastionHostEnabled = this.getFieldValue('awsNodeSettingForm', 'bastionHostEnabled');
        payload.bastionHostEnabled = bastionHostEnabled === BASTION_HOST_ENABLED;
        const machineHealthChecksEnabled = this.getFieldValue('awsNodeSettingForm', 'machineHealthChecksEnabled');
        payload.machineHealthCheckEnabled = (machineHealthChecksEnabled === true);
        payload.vpc = {
            cidr: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                this.getFieldValue('vpcForm', 'existingVpcCidr') :
                this.getFieldValue('vpcForm', 'vpc'),
            vpcID: this.getFieldValue('vpcForm', 'existingVpcId'),
            azs: this.getAwsNodeAzs(payload)
        };

        payload.enableAuditLogging = this.getBooleanFieldValue("awsNodeSettingForm", "enableAuditLogging");
        this.initPayloadWithCommons(payload);

        return payload;
    }

    getAwsNodeAzs(payload) {
        this.nodeAzList = [
            {
                name: this.getFieldValue('awsNodeSettingForm', 'awsNodeAz1'),
                workerNodeType: Broker.appDataService.isModeClusterStandalone() ? payload.controlPlaneNodeType :
                    this.getFieldValue('awsNodeSettingForm', 'workerNodeInstanceType1'),
                publicNodeCidr: (this.getFieldValue('vpcForm', 'vpcType') === 'new') ?
                    this.getFieldValue('vpcForm', 'publicNodeCidr') : '',
                privateNodeCidr: (this.getFieldValue('vpcForm', 'vpcType') === 'new') ?
                    this.getFieldValue('vpcForm', 'privateNodeCidr') : '',
                publicSubnetID: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                    this.getFieldValue('awsNodeSettingForm', 'vpcPublicSubnet1') : '',
                privateSubnetID: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                    this.getFieldValue('awsNodeSettingForm', 'vpcPrivateSubnet1') : ''
            }
        ];

        if (this.getFieldValue('awsNodeSettingForm', 'awsNodeAz2')) {
            this.nodeAzList.push({
                name: this.getFieldValue('awsNodeSettingForm', 'awsNodeAz2'),
                workerNodeType: (!Broker.appDataService.isModeClusterStandalone()) ?
                    this.getFieldValue('awsNodeSettingForm', 'workerNodeInstanceType2') : payload.controlPlaneNodeType,
                publicNodeCidr: (this.getFieldValue('vpcForm', 'vpcType') === 'new') ?
                    this.getFieldValue('vpcForm', 'publicNodeCidr') : '',
                privateNodeCidr: (this.getFieldValue('vpcForm', 'vpcType') === 'new') ?
                    this.getFieldValue('vpcForm', 'privateNodeCidr') : '',
                publicSubnetID: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                    this.getFieldValue('awsNodeSettingForm', 'vpcPublicSubnet2') : '',
                privateSubnetID: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                    this.getFieldValue('awsNodeSettingForm', 'vpcPrivateSubnet2') : ''
            });
        }

        if (this.getFieldValue('awsNodeSettingForm', 'awsNodeAz3')) {
            this.nodeAzList.push({
                name: this.getFieldValue('awsNodeSettingForm', 'awsNodeAz3'),
                workerNodeType: (!Broker.appDataService.isModeClusterStandalone()) ?
                    this.getFieldValue('awsNodeSettingForm', 'workerNodeInstanceType3') : payload.controlPlaneNodeType,
                publicNodeCidr: (this.getFieldValue('vpcForm', 'vpcType') === 'new') ?
                    this.getFieldValue('vpcForm', 'publicNodeCidr') : '',
                privateNodeCidr: (this.getFieldValue('vpcForm', 'vpcType') === 'new') ?
                    this.getFieldValue('vpcForm', 'privateNodeCidr') : '',
                publicSubnetID: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                    this.getFieldValue('awsNodeSettingForm', 'vpcPublicSubnet3') : '',
                privateSubnetID: (this.getFieldValue('vpcForm', 'vpcType') === 'existing') ?
                    this.getFieldValue('awsNodeSettingForm', 'vpcPrivateSubnet3') : ''
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
        return this.getFieldValue('awsNodeSettingForm', 'clusterName');
    }

    /**
     * @method getExtendCliCmds to return cli command string according to special selection
     * For AWS, selects Create Cloudformation Stack,
     * should include tanzu management-cluster permissions aws set
     * @returns the array includes cli command object like {isPrefixOfCreateCmd: true, cmdStr: "tanzu ..."}
     */
    getExtendCliCmds(): Array<{ isPrefixOfCreateCmd: boolean, cmdStr: string }> {
        if (this.getFieldValue('awsNodeSettingForm', 'createCloudFormation')) {
            const clusterPrefix = (this.clusterType) ? this.clusterType : 'management';
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

}
