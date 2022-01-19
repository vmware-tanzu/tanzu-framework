// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

// Library imports
import { Observable } from 'rxjs';
import { APIClient,
    AzureInstanceType,
    AzureRegionalClusterParams,
    AzureResourceGroup,
    AzureVirtualMachine,
    AzureVirtualNetwork
} from 'tanzu-ui-api-lib';

// App imports
import AppServices from '../../../shared/service/appServices';
import { AzureForm } from './azure-wizard.constants';
import { AzureAccountParamsKeys, AzureProviderStepComponent } from './provider-step/azure-provider-step.component';
import { AzureOsImageStepComponent } from './os-image-step/azure-os-image-step.component';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { EXISTING, VnetStepComponent } from './vnet-step/vnet-step.component';
import { ExportService } from '../../../shared/service/export.service';
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { ImportParams, ImportService } from '../../../shared/service/import.service';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { TkgEventType } from '../../../shared/service/Messenger';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';

@Component({
    selector: 'app-azure-wizard',
    templateUrl: './azure-wizard.component.html',
    styleUrls: ['./azure-wizard.component.scss']
})
export class AzureWizardComponent extends WizardBaseDirective implements OnInit {
    region: string;

    constructor(
        router: Router,
        private importService: ImportService,
        private exportService: ExportService,
        formBuilder: FormBuilder,
        private apiClient: APIClient,
        titleService: Title,
        formMetaDataService: FormMetaDataService,
        el: ElementRef) {

        super(router, el, formMetaDataService, titleService, formBuilder);
    }

    protected supplyWizardName(): string {
        return 'Azure Wizard';
    }

    protected supplyStepData(): FormDataForHTML[] {
        return [
            this.AzureProviderForm,
            this.AzureVnetForm,
            this.AzureNodeSettingForm,
            this.MetadataForm,
            this.AzureNetworkForm,
            this.IdentityForm,
            this.AzureOsImageForm,
            this.CeipForm,
        ];
    }

    ngOnInit() {
        super.ngOnInit();
        this.titleService.setTitle(this.title + ' Azure');
        this.registerServices();
        this.subscribeToServices();
    }

    getPayload(): any {
        const payload: AzureRegionalClusterParams = {};

        payload.azureAccountParams = {};
        AzureAccountParamsKeys.forEach(key => {
            payload.azureAccountParams[key] = this.getFieldValue(AzureForm.PROVIDER, key);
        });

        const mappings = [
            ["location", AzureForm.PROVIDER, "region"],
            ["sshPublicKey", AzureForm.PROVIDER, "sshPublicKey"],
        ];

        mappings.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));

        payload.controlPlaneMachineType = this.getControlPlaneNodeType("azure");
        payload.controlPlaneFlavor = this.getControlPlaneFlavor("azure");
        payload.workerMachineType = AppServices.appDataService.isModeClusterStandalone() ? payload.controlPlaneMachineType :
            this.getFieldValue(AzureForm.NODESETTING, 'workerNodeInstanceType');
        payload.machineHealthCheckEnabled = this.getBooleanFieldValue(AzureForm.NODESETTING, "machineHealthChecksEnabled");

        const resourceGroupOption = this.getFieldValue(AzureForm.PROVIDER, "resourceGroupOption");
        const resourceGroupField = resourceGroupOption === 'existing' ? 'resourceGroupExisting' : 'resourceGroupCustom';
        payload.resourceGroup = this.getFieldValue(AzureForm.PROVIDER, resourceGroupField);
        payload.clusterName = this.getMCName();

        // Retrieve vnet info
        payload.vnetResourceGroup = this.getFieldValue(AzureForm.VNET, 'vnetResourceGroup');

        const vnetOption = this.getFieldValue(AzureForm.VNET, "vnetOption");
        let vnetAttrs = [       // For new vnet
            ["vnetName", AzureForm.VNET, "vnetNameCustom"],
            ["vnetCidr", AzureForm.VNET, "vnetCidrBlock"],
            ["controlPlaneSubnet", AzureForm.VNET, "controlPlaneSubnetNew"],
            ["controlPlaneSubnetCidr", AzureForm.VNET, "controlPlaneSubnetCidrNew"],
            ["workerNodeSubnet", AzureForm.VNET, "workerNodeSubnetNew"],
            ["workerNodeSubnetCidr", AzureForm.VNET, "workerNodeSubnetCidrNew"],
        ];

        if (vnetOption === EXISTING) {        // for existing vnet
            vnetAttrs = [
                ["vnetName", AzureForm.VNET, "vnetNameExisting"],
                ["vnetCidr", AzureForm.VNET, "vnetCidrBlock"],
                ["controlPlaneSubnet", AzureForm.VNET, "controlPlaneSubnet"],
                ["controlPlaneSubnetCidr", AzureForm.VNET, "controlPlaneSubnetCidr"],
                ["workerNodeSubnet", AzureForm.VNET, "workerNodeSubnet"],
            ];
        }
        vnetAttrs.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));

        payload.enableAuditLogging = this.getBooleanFieldValue(AzureForm.NODESETTING, "enableAuditLogging");

        this.initPayloadWithCommons(payload);

        // private Azure cluster support
        payload.isPrivateCluster = this.getBooleanFieldValue(AzureForm.VNET, "privateAzureCluster");

        payload.frontendPrivateIp = "";
        if (payload.isPrivateCluster) {
            payload.frontendPrivateIp = this.getFieldValue(AzureForm.VNET, "privateIP");
        }

        return payload;
    }

    setFromPayload(payload: AzureRegionalClusterParams) {
        if (payload !== undefined) {
            if (payload.azureAccountParams !== undefined) {
                for (const accountFieldName of Object.keys(payload.azureAccountParams)) {
                    // we treat azureCloud differently because it's a listbox selection where the label != key
                    if (accountFieldName !== 'azureCloud') {
                        this.saveFormField(AzureForm.PROVIDER, accountFieldName, payload.azureAccountParams[accountFieldName]);
                    }
                }
                this.saveFormListbox(AzureForm.PROVIDER, 'azureCloud', payload.azureAccountParams['azureCloud']);
            }
            this.saveFormField(AzureForm.PROVIDER, "sshPublicKey", payload["sshPublicKey"]);
            this.saveFormListbox(AzureForm.PROVIDER, "region", payload["location"]);

            this.saveControlPlaneFlavor('azure', payload.controlPlaneFlavor);
            this.saveControlPlaneNodeType('azure', payload.controlPlaneFlavor, payload.controlPlaneMachineType);

            if (!AppServices.appDataService.isModeClusterStandalone()) {
                this.saveFormField(AzureForm.NODESETTING, 'workerNodeInstanceType', payload.workerMachineType);
            }
            this.saveFormField(AzureForm.NODESETTING, "machineHealthChecksEnabled", payload.machineHealthCheckEnabled);

            // Since we cannot tell if the resource group is custom or existing, we load it into the custom field.
            // When the resource groups are retrieved, we have code that will detect if the resource group is existing.
            // See azure-provider-step.component.ts's handleIfSavedCustomResourceGroupIsNowExisting()
            this.saveFormField(AzureForm.PROVIDER, 'resourceGroupCustom', payload.resourceGroup);

            this.saveMCName(payload.clusterName);

            // We canot tell if the vnet is custom or existing, so we load it into the custom field.
            // When the vnet resource groups are retrieved, we have code that will detect if the vnet is existing.
            // See vnet-step.component.ts's handleIfSavedVnetCustomNameIsNowExisting()
            const vnetAttrs = [
                ['vnetResourceGroup', 'vnetResourceGroup'],
                ["vnetName", "vnetNameCustom"],
                ["vnetCidr", "vnetCidrBlock"],
                ["controlPlaneSubnet", "controlPlaneSubnetNew"],
                ["controlPlaneSubnetCidr", "controlPlaneSubnetCidrNew"],
                ["workerNodeSubnet", "workerNodeSubnetNew"],
                ["workerNodeSubnetCidr", "workerNodeSubnetCidrNew"],
            ];
            vnetAttrs.forEach(attr => payload[attr[0]] = this.saveFormField(AzureForm.VNET, attr[1], payload[attr[0]]));
            this.saveFormField(AzureForm.VNET, 'privateAzureCluster', payload.isPrivateCluster);
            if (payload.isPrivateCluster) {
                this.saveFormField(AzureForm.VNET, 'privateIP', payload.frontendPrivateIp);
            }
            this.saveFormField(AzureForm.NODESETTING, 'enableAuditLogging', payload.enableAuditLogging);
            this.saveCommonFieldsFromPayload(payload);
        }
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
        return this.getFieldValue(AzureForm.NODESETTING, "managementClusterName");
    }

    saveMCName(clusterName: string) {
        this.saveFormField(AzureForm.NODESETTING, "managementClusterName", clusterName);
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
        const resourceGroupOption = this.getFieldValue(AzureForm.PROVIDER, 'resourceGroupOption')
        const azureResourceGroup = resourceGroupOption === 'existing' ? 'resourceGroupExisting' : 'resourceGroupCustom';
        const vnetOption = this.getFieldValue(AzureForm.VNET, 'vnetOption');
        const azureVnetName = vnetOption === 'existing' ? 'vnetNameExisting' : 'vnetNameCustom';
        const azureControlPlaneSubnetName = vnetOption === 'existing' ? 'controlPlaneSubnet' : 'controlPlaneSubnetNew';
        const azureNodeSubnetName = vnetOption === 'existing' ? 'workerNodeSubnet' : 'workerNodeSubnetNew';
        const fieldsMapping = {
            AZURE_RESOURCE_GROUP: [AzureForm.PROVIDER, azureResourceGroup],
            AZURE_VNET_RESOURCE_GROUP: [AzureForm.VNET, 'vnetResourceGroup'],
            AZURE_VNET_NAME: [AzureForm.VNET, azureVnetName],
            AZURE_VNET_CIDR: [AzureForm.VNET, 'vnetCidrBlock'],
            AZURE_CONTROL_PLANE_SUBNET_NAME: [AzureForm.VNET, azureControlPlaneSubnetName],
            AZURE_CONTROL_PLANE_SUBNET_CIDR: [AzureForm.VNET, 'controlPlaneSubnetCidrNew'],
            AZURE_NODE_SUBNET_NAME: [AzureForm.VNET, azureNodeSubnetName],
            AZURE_NODE_SUBNET_CIDR: [AzureForm.VNET, 'workerNodeSubnetCidrNew']
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
            const temp = fieldsMapping[field];
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

    exportConfiguration() {
        const wizard = this;    // capture 'this' outside the context of the closure below
        this.exportService.export(
            this.retrieveExportFile(),
            (failureMessage) => { wizard.displayError(failureMessage); }
        );
    }

    getAdditionalNoProxyInfo() {
        const vnetCidr = this.getFieldValue('vpcForm', 'vnetCidrBlock');
        return (vnetCidr ? vnetCidr + ',' : '')  + '169.254.0.0/16,168.63.129.16';
    }

    // HTML convenience methods
    //
    get AzureProviderForm(): FormDataForHTML {
        return { name: AzureForm.PROVIDER, title: 'IaaS Provider', description: 'Validate the Azure provider credentials for Tanzu',
        i18n: {title: 'IaaS provder step name', description: 'IaaS provder step description'},
        clazz: AzureProviderStepComponent};
    }
    get AzureVnetForm(): FormDataForHTML {
        return { name: AzureForm.VNET, title: 'Azure VNET Settings', description: 'Specify an Azure VNET CIDR',
            i18n: {title: 'vnet step name', description: 'vnet step description'},
        clazz: VnetStepComponent};
    }
    get AzureNodeSettingForm(): FormDataForHTML {
        return { name: AzureForm.NODESETTING, title: FormUtility.titleCase(this.clusterTypeDescriptor) + ' Cluster Settings',
            description: `Specifying the resources backing the ${this.clusterTypeDescriptor} cluster`,
            i18n: {title: 'node setting step name', description: 'node setting step description'},
        clazz: NodeSettingStepComponent};
    }
    get AzureOsImageForm(): FormDataForHTML {
        return this.getOsImageForm(AzureOsImageStepComponent);
    }
    get AzureNetworkForm(): FormDataForHTML {
        return FormUtility.formWithOverrides(this.NetworkForm, {description: 'Specify an Azure VNET CIDR'});
    }
    //
    // HTML convenience methods

    // returns TRUE if the file contents appear to be a valid config file for Azure
    // returns FALSE if the file is empty or does not appear to be valid. Note that in the FALSE
    // case we also alert the user.
    importFileValidate(nameFile: string, fileContents: string): boolean {
        if (fileContents.includes('AZURE_')) {
            return true;
        }
        alert(nameFile + ' is not a valid Azure configuration file!');
        return false;
    }

    importFileRetrieveClusterParams(fileContents: string): Observable<AzureRegionalClusterParams> {
        return this.apiClient.importTKGConfigForAzure( { params: { filecontents: fileContents } } );
    }

    importFileProcessClusterParams(nameFile: string, azureClusterParams: AzureRegionalClusterParams) {
        this.setFromPayload(azureClusterParams);
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
        const params: ImportParams<AzureRegionalClusterParams> = {
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

    private subscribeToServices() {
        AppServices.messenger.getSubject(TkgEventType.AZURE_REGION_CHANGED)
            .subscribe(event => {
                const region = event.payload;
                if (this.region) {
                    AppServices.dataServiceRegistrar.trigger([
                        TkgEventType.AZURE_GET_RESOURCE_GROUPS,
                        TkgEventType.AZURE_GET_INSTANCE_TYPES
                    ], { location: region });
                    AppServices.dataServiceRegistrar.trigger([TkgEventType.AZURE_GET_OS_IMAGES]);
                } else {
                    AppServices.dataServiceRegistrar.clear<AzureResourceGroup>(TkgEventType.AZURE_GET_RESOURCE_GROUPS);
                    AppServices.dataServiceRegistrar.clear<AzureInstanceType>(TkgEventType.AZURE_GET_INSTANCE_TYPES);
                    AppServices.dataServiceRegistrar.clear<AzureVirtualMachine>(TkgEventType.AZURE_GET_OS_IMAGES);
                }
            });
    }

    private registerServices() {
        const wizard = this;
        AppServices.dataServiceRegistrar.register<AzureResourceGroup>(TkgEventType.AZURE_GET_RESOURCE_GROUPS,
            (payload: {location: string}) => { return wizard.apiClient.getAzureResourceGroups(payload); },
            "Failed to retrieve resource groups for the particular region." );
        AppServices.dataServiceRegistrar.register<AzureInstanceType>(TkgEventType.AZURE_GET_INSTANCE_TYPES,
            (payload: {location: string}) => { return wizard.apiClient.getAzureInstanceTypes(payload); },
            "Failed to retrieve Azure VM sizes" );
        AppServices.dataServiceRegistrar.register<AzureVirtualMachine>(TkgEventType.AZURE_GET_OS_IMAGES,
            () => { return wizard.apiClient.getAzureOSImages(); },
            "Failed to retrieve list of OS images from the specified Azure Server." );
        AppServices.dataServiceRegistrar.register<AzureVirtualNetwork>(TkgEventType.AZURE_GET_VNETS,
            (payload: {resourceGroupName: string, location: string}) => { return wizard.apiClient.getAzureVnets(payload)},
            "Failed to retrieve list of VNETs from the specified Azure Server." );
    }
}
