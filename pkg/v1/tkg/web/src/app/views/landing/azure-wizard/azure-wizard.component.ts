import { Component, OnInit, ElementRef } from '@angular/core';
import { Router } from '@angular/router';
import { FormBuilder } from '@angular/forms';
import { Title } from '@angular/platform-browser';
import { APIClient } from 'src/app/swagger';

import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { Observable, EMPTY, throwError, of } from 'rxjs';
import { CliGenerator, CliFields } from '../wizard/shared/utils/cli-generator';
import { AzureRegionalClusterParams } from 'src/app/swagger/models';
import { AzureAccountParamsKeys } from './provider-step/azure-provider-step.component';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { EXISTING } from './vnet-step/vnet-step.component';
import Broker from 'src/app/shared/service/broker';

@Component({
    selector: 'app-azure-wizard',
    templateUrl: './azure-wizard.component.html',
    styleUrls: ['./azure-wizard.component.scss']
})
export class AzureWizardComponent extends WizardBaseDirective implements OnInit {

    // The region user selected
    region: string;

    constructor(
        router: Router,
        public wizardFormService: AzureWizardFormService,
        private formBuilder: FormBuilder,
        private apiClient: APIClient,
        titleService: Title,
        formMetaDataService: FormMetaDataService,
        el: ElementRef) {

        super(router, el, formMetaDataService, titleService);

        this.form = this.formBuilder.group({
            azureProviderForm: this.formBuilder.group({
            }),
            vnetForm: this.formBuilder.group({
            }),
            azureNodeSettingForm: this.formBuilder.group({
            }),
            workerazureNodeSettingForm: this.formBuilder.group({
            }),
            metadataForm: this.formBuilder.group({
            }),
            networkForm: this.formBuilder.group({
            }),
            identityForm: this.formBuilder.group({
            }),
            osImageForm: this.formBuilder.group({
            }),
            ceipOptInForm: this.formBuilder.group({
            })
        });
    }

    ngOnInit() {
        super.ngOnInit();

        this.titleService.setTitle(this.title + ' Azure');
    }

    getStepDescription(stepName: string): string {
        if (stepName === 'azureProviderForm') {
            const tenant = this.getFieldValue('azureProviderForm', 'tenantId');
            return tenant ? `Azure tenant: ${tenant}` : 'Validate the Azure provider credentials for Tanzu';
        } else if (stepName === 'vnetForm') {
            const vnetCidrBlock = this.getFieldValue(stepName, "vnetCidrBlock");
            if (vnetCidrBlock) {
                return `Subnet: ${vnetCidrBlock}`;
            }
            return "Specify a Azure VNET CIDR";
        } else if (stepName === 'azureNodeSettingForm') {
            const controlPlaneSetting = this.getFieldValue(stepName, "controlPlaneSetting");
            if (controlPlaneSetting) {
                return `Control plane type: ${controlPlaneSetting}`;
            }
            return `Specifying the resources backing the ${this.clusterTypeDescriptor} cluster`;
        } else if (stepName === 'metadataForm') {
            const location = this.getFieldValue(stepName, "clusterLocation");
            if (location) {
                return `Location: ${location}`;
            }
            return `Specify metadata for the ${this.clusterTypeDescriptor} cluster`;
        } else if (stepName === 'networkForm') {
            const serviceCidr = this.getFieldValue(stepName, "clusterServiceCidr");
            const podCidr = this.getFieldValue(stepName, "clusterPodCidr");
            if (serviceCidr && podCidr) {
                return `Cluster service CIDR: ${serviceCidr} Cluster POD CIDR: ${podCidr}`;
            }
            return "Specify how TKG networking is provided and global network settings";
        } else if (stepName === 'ceipOptInForm') {
            return "Join the CEIP program for TKG";
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
        return `Step ${stepName} is not supported yet`;
    }

    getPayload(): any {
        const payload: AzureRegionalClusterParams = {};

        payload.azureAccountParams = {};
        AzureAccountParamsKeys.forEach(key => {
            payload.azureAccountParams[key] = this.getFieldValue("azureProviderForm", key);
        });

        const mappings = [
            ["location", "azureProviderForm", "region"],
            ["sshPublicKey", "azureProviderForm", "sshPublicKey"],
        ];

        mappings.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));

        payload.controlPlaneMachineType = this.getControlPlaneNodeType("azure");
        payload.controlPlaneFlavor = this.getControlPlaneFlavor("azure");
        payload.workerMachineType = Broker.appDataService.isModeClusterStandalone() ? payload.controlPlaneMachineType :
            this.getFieldValue('azureNodeSettingForm', 'workerNodeInstanceType');
        payload.machineHealthCheckEnabled = this.getBooleanFieldValue("azureNodeSettingForm", "machineHealthChecksEnabled");

        const resourceGroupOption = this.getFieldValue("azureProviderForm", "resourceGroupOption");
        let resourceGroupField = "resourceGroupCustom";
        if (resourceGroupOption === "existing") {
            resourceGroupField = "resourceGroupExisting";
        }
        payload.resourceGroup = this.getFieldValue("azureProviderForm", resourceGroupField);
        payload.clusterName = this.getMCName();

        // Retrieve vnet info
        payload.vnetResourceGroup = this.getFieldValue("vnetForm", "resourceGroup");

        const vnetOption = this.getFieldValue("vnetForm", "vnetOption");
        let vnetAttrs = [       // For new vnet
            ["vnetName", "vnetForm", "vnetNameCustom"],
            ["vnetCidr", "vnetForm", "vnetCidrBlock"],
            ["controlPlaneSubnet", "vnetForm", "controlPlaneSubnetNew"],
            ["controlPlaneSubnetCidr", "vnetForm", "controlPlaneSubnetCidrNew"],
            ["workerNodeSubnet", "vnetForm", "workerNodeSubnetNew"],
            ["workerNodeSubnetCidr", "vnetForm", "workerNodeSubnetCidrNew"],
        ];

        if (vnetOption === EXISTING) {        // for existing vnet
            vnetAttrs = [
                ["vnetName", "vnetForm", "vnetNameExisting"],
                ["vnetCidr", "vnetForm", "vnetCidrBlock"],
                ["controlPlaneSubnet", "vnetForm", "controlPlaneSubnet"],
                ["controlPlaneSubnetCidr", "vnetForm", "controlPlaneSubnetCidr"],
                ["workerNodeSubnet", "vnetForm", "workerNodeSubnet"],
            ];
        }
        vnetAttrs.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));

        payload.enableAuditLogging = this.getBooleanFieldValue("azureNodeSettingForm", "enableAuditLogging");

        this.initPayloadWithCommons(payload);

        // private Azure cluster support
        payload.isPrivateCluster = this.getBooleanFieldValue("vnetForm", "privateAzureCluster");

        payload.frontendPrivateIp = "";
        if (payload.isPrivateCluster) {
            payload.frontendPrivateIp = this.getFieldValue("vnetForm", "privateIP");
        }

        return payload;
    }

    /**
     * @method method to trigger deployment
     */
    createRegionalCluster(payload: any): Observable<any> {
        return this.apiClient.createAzureRegionalCluster(payload);
    }

    /**
     * Return management/standalone cluster name
     */
    getMCName() {
        return this.getFieldValue("azureNodeSettingForm", "managementClusterName");
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
            extendCliCmds: []
        };
        return cliG.getCli(cliParams);
    }

    getCliEnvVariables() {
        let envVariableString = '';
        const resourceGroupOption = this.getFieldValue('azureProviderForm', 'resourceGroupOption')
        const azureResourceGroup = resourceGroupOption === 'existing' ? 'resourceGroupExisting' : 'resourceGroupCustom';
        const vnetOption = this.getFieldValue('vnetForm', 'vnetOption');
        const azureVnetName = vnetOption === 'existing' ? 'vnetNameExisting' : 'vnetNameCustom';
        const azureControlPlaneSubnetName = vnetOption === 'existing' ? 'controlPlaneSubnet' : 'controlPlaneSubnetNew';
        const azureNodeSubnetName = vnetOption === 'existing' ? 'workerNodeSubnet' : 'workerNodeSubnetNew';
        const fieldsMapping = {
            AZURE_RESOURCE_GROUP: ['azureProviderForm', azureResourceGroup],
            AZURE_VNET_RESOURCE_GROUP: ['vnetForm', 'resourceGroup'],
            AZURE_VNET_NAME: ['vnetForm', azureVnetName],
            AZURE_VNET_CIDR: ['vnetForm', 'vnetCidrBlock'],
            AZURE_CONTROL_PLANE_SUBNET_NAME: ['vnetForm', azureControlPlaneSubnetName],
            AZURE_CONTROL_PLANE_SUBNET_CIDR: ['vnetForm', 'controlPlaneSubnetCidrNew'],
            AZURE_NODE_SUBNET_NAME: ['vnetForm', azureNodeSubnetName],
            AZURE_NODE_SUBNET_CIDR: ['vnetForm', 'workerNodeSubnetCidrNew']
        }
        let fields = [];
        if (vnetOption === 'existing') {
            fields = [
                'AZURE_RESOURCE_GROUP',
                'AZURE_VNET_RESOURCE_GROUP',
                'AZURE_VNET_NAME',
                'AZURE_CONTROL_PLANE_SUBNET_NAME',
                'AZURE_NODE_SUBNET_NAME'
            ];
        } else {
            fields = [
                'AZURE_RESOURCE_GROUP',
                'AZURE_VNET_RESOURCE_GROUP',
                'AZURE_VNET_NAME',
                'AZURE_VNET_CIDR',
                'AZURE_CONTROL_PLANE_SUBNET_NAME',
                'AZURE_CONTROL_PLANE_SUBNET_CIDR',
                'AZURE_NODE_SUBNET_NAME',
                'AZURE_NODE_SUBNET_CIDR'
            ];
        }
        fields.forEach(field => {
            const temp =  fieldsMapping[field];
            envVariableString += `${field}="${this.getFieldValue(temp[0], temp[1])}" `;
        });
        return envVariableString;
    }
    applyTkgConfig() {
        return this.apiClient.applyTKGConfigForAzure({ params: this.getPayload() });
    }

    /**
     * Retrieve the config file from the backend and return as a string
     */
    retrieveExportFile() {
        return this.apiClient.exportTKGConfigForAzure({ params: this.getPayload() });
    }

    getAdditionalNoProxyInfo() {
        const vnetCidr = this.getFieldValue('vpcForm', 'vnetCidrBlock');
        return (vnetCidr ? vnetCidr + ',' : '')  + '169.254.0.0/16,168.63.129.16';
    }
}
