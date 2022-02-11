// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';
// Third party imports
import { Observable } from 'rxjs';
// App imports
import { APIClient } from 'src/app/swagger';
import AppServices from '../../../shared/service/appServices';
import { AzureClouds, AzureField, AzureForm, ResourceGroupOption, VnetOptionType } from './azure-wizard.constants';
import {
    AzureInstanceType,
    AzureRegionalClusterParams,
    AzureResourceGroup,
    AzureVirtualMachine,
    AzureVirtualNetwork
} from 'src/app/swagger/models';
import { AzureAccountParamsKeys, AzureProviderStepComponent } from './provider-step/azure-provider-step.component';
import { AzureOsImageStepComponent } from './os-image-step/azure-os-image-step.component';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { VnetStepComponent } from './vnet-step/vnet-step.component';
import { ExportService } from '../../../shared/service/export.service';
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';
import { ImportParams, ImportService } from "../../../shared/service/import.service";
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { TanzuEventType } from '../../../shared/service/Messenger';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { WizardForm } from '../wizard/shared/constants/wizard.constants';
import { OsImageField } from '../wizard/shared/components/steps/os-image-step/os-image-step.fieldmapping';

@Component({
    selector: 'app-azure-wizard',
    templateUrl: './azure-wizard.component.html',
    styleUrls: ['./azure-wizard.component.scss']
})
export class AzureWizardComponent extends WizardBaseDirective implements OnInit {
    constructor(
        router: Router,
        private importService: ImportService,
        private exportService: ExportService,
        formBuilder: FormBuilder,
        private apiClient: APIClient,
        titleService: Title,
        el: ElementRef) {

        super(router, el, titleService, formBuilder);
    }

    protected supplyFileImportedEvent(): TanzuEventType {
        return TanzuEventType.AZURE_CONFIG_FILE_IMPORTED;
    }

    protected supplyFileImportErrorEvent(): TanzuEventType {
        return TanzuEventType.AZURE_CONFIG_FILE_IMPORT_ERROR;
    }

    protected supplyWizardName(): string {
        return 'AzureWizard';
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
            ["location", AzureForm.PROVIDER, AzureField.PROVIDER_REGION],
            ["sshPublicKey", AzureForm.PROVIDER, AzureField.PROVIDER_SSHPUBLICKEY],
        ];

        mappings.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));

        payload.controlPlaneFlavor = this.getFieldValue(AzureForm.NODESETTING, AzureField.NODESETTING_CONTROL_PLANE_SETTING);
        const nodeTypeField = payload.controlPlaneFlavor === 'prod' ? AzureField.NODESETTING_INSTANCE_TYPE_PROD
            : AzureField.NODESETTING_INSTANCE_TYPE_DEV;
        payload.controlPlaneMachineType = this.getFieldValue(AzureForm.NODESETTING, nodeTypeField);

        payload.workerMachineType = AppServices.appDataService.isModeClusterStandalone() ? payload.controlPlaneMachineType :
            this.getFieldValue(AzureForm.NODESETTING, AzureField.NODESETTING_WORKERTYPE);
        payload.machineHealthCheckEnabled =
            this.getBooleanFieldValue(AzureForm.NODESETTING, AzureField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED);

        const resourceGroupOption = this.getFieldValue(AzureForm.PROVIDER, AzureField.PROVIDER_RESOURCEGROUPOPTION);
        const resourceGroupField = resourceGroupOption === ResourceGroupOption.EXISTING ? AzureField.PROVIDER_RESOURCEGROUPEXISTING :
            AzureField.PROVIDER_RESOURCEGROUPCUSTOM;
        payload.resourceGroup = this.getFieldValue(AzureForm.PROVIDER, resourceGroupField);
        payload.clusterName = this.getMCName();

        // Retrieve vnet info
        payload.vnetResourceGroup = this.getFieldValue(AzureForm.VNET, AzureField.VNET_RESOURCE_GROUP);

        const vnetOption = this.getFieldValue(AzureForm.VNET, AzureField.VNET_EXISTING_OR_CUSTOM);
        let vnetAttrs = [       // For new vnet
            ["vnetName", AzureForm.VNET, AzureField.VNET_CUSTOM_NAME],
            ["vnetCidr", AzureForm.VNET, AzureField.VNET_CUSTOM_CIDR],
            ["controlPlaneSubnet", AzureForm.VNET, AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME],
            ["controlPlaneSubnetCidr", AzureForm.VNET, AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR],
            ["workerNodeSubnet", AzureForm.VNET, AzureField.VNET_WORKER_NEWSUBNET_NAME],
            ["workerNodeSubnetCidr", AzureForm.VNET, AzureField.VNET_WORKER_NEWSUBNET_CIDR],
        ];

        if (vnetOption === VnetOptionType.EXISTING) {        // for existing vnet
            vnetAttrs = [
                ["vnetName", AzureForm.VNET, AzureField.VNET_EXISTING_NAME],
                ["vnetCidr", AzureForm.VNET, AzureField.VNET_CUSTOM_CIDR],
                ["controlPlaneSubnet", AzureForm.VNET, AzureField.VNET_CONTROLPLANE_SUBNET_NAME],
                ["controlPlaneSubnetCidr", AzureForm.VNET, AzureField.VNET_CONTROLPLANE_SUBNET_CIDR],
                ["workerNodeSubnet", AzureForm.VNET, AzureField.VNET_WORKER_SUBNET_NAME],
            ];
        }
        vnetAttrs.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));

        payload.enableAuditLogging = this.getBooleanFieldValue(AzureForm.NODESETTING, AzureField.NODESETTING_ENABLE_AUDIT_LOGGING);

        this.initPayloadWithCommons(payload);

        // private Azure cluster support
        payload.isPrivateCluster = this.getBooleanFieldValue(AzureForm.VNET, AzureField.VNET_PRIVATE_CLUSTER);

        payload.frontendPrivateIp = "";
        if (payload.isPrivateCluster) {
            payload.frontendPrivateIp = this.getFieldValue(AzureForm.VNET, AzureField.VNET_PRIVATE_IP);
        }

        return payload;
    }

    setFromPayload(payload: AzureRegionalClusterParams) {
        if (payload !== undefined) {
            if (payload.azureAccountParams !== undefined) {
                for (const accountFieldName of Object.keys(payload.azureAccountParams)) {
                    // we treat azureCloud differently because it's a listbox selection where the label != key
                    if (accountFieldName !== 'azureCloud') {
                        this.storeFieldString(AzureForm.PROVIDER, accountFieldName, payload.azureAccountParams[accountFieldName]);
                    }
                }
                const azureCloudValue = payload.azureAccountParams['azureCloud'];
                const azureCloud = azureCloudValue ? AzureClouds.find(cloud => cloud.name === azureCloudValue) : undefined;
                if (azureCloud) {
                    this.storeFieldString(AzureForm.PROVIDER, AzureField.PROVIDER_AZURECLOUD, azureCloud.name, azureCloud.displayName);
                }
            }
            this.storeFieldString(AzureForm.PROVIDER, AzureField.PROVIDER_SSHPUBLICKEY, payload["sshPublicKey"]);
            this.storeFieldString(AzureForm.PROVIDER, AzureField.PROVIDER_REGION, payload["location"]);

            this.storeFieldString(AzureForm.NODESETTING, AzureField.NODESETTING_CONTROL_PLANE_SETTING, payload.controlPlaneFlavor);
            const instanceTypeField = payload.controlPlaneFlavor === 'prod' ? AzureField.NODESETTING_INSTANCE_TYPE_PROD
                : AzureField.NODESETTING_INSTANCE_TYPE_DEV;
            this.storeFieldString(AzureForm.NODESETTING, instanceTypeField, payload.controlPlaneMachineType);

            if (!AppServices.appDataService.isModeClusterStandalone()) {
                this.storeFieldString(AzureForm.NODESETTING, AzureField.NODESETTING_WORKERTYPE, payload.workerMachineType);
            }
            this.storeFieldBoolean(AzureForm.NODESETTING, AzureField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED,
                payload.machineHealthCheckEnabled);

            // Since we cannot tell if the resource group is custom or existing, we load it into the custom field.
            // When the resource groups are retrieved, we have code that will detect if the resource group is existing.
            // See azure-provider-step.component.ts's handleIfSavedCustomResourceGroupIsNowExisting()
            this.storeFieldString(AzureForm.PROVIDER, AzureField.PROVIDER_RESOURCEGROUPCUSTOM, payload.resourceGroup);

            this.saveMCName(payload.clusterName);

            // We cannot tell if the vnet is custom or existing, so we load it into the custom field.
            // When the vnet resource groups are retrieved, we have code that will detect if the vnet is existing.
            // See vnet-step.component.ts's handleIfSavedVnetCustomNameIsNowExisting()
            const vnetAttrs = [
                ['vnetResourceGroup', AzureField.VNET_RESOURCE_GROUP],
                ["vnetName", AzureField.VNET_CUSTOM_NAME],
                ["vnetCidr", AzureField.VNET_CUSTOM_CIDR],
                ["controlPlaneSubnet", AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME],
                ["controlPlaneSubnetCidr", AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR],
                ["workerNodeSubnet", AzureField.VNET_WORKER_NEWSUBNET_NAME],
                ["workerNodeSubnetCidr", AzureField.VNET_WORKER_NEWSUBNET_CIDR],
            ];
            vnetAttrs.forEach(attr => payload[attr[0]] = this.storeFieldString(AzureForm.VNET, attr[1], payload[attr[0]]));
            this.storeFieldBoolean(AzureForm.VNET, AzureField.VNET_PRIVATE_CLUSTER, payload.isPrivateCluster);
            if (payload.isPrivateCluster) {
                this.storeFieldString(AzureForm.VNET, AzureField.VNET_PRIVATE_IP, payload.frontendPrivateIp);
            }
            this.storeFieldBoolean(AzureForm.NODESETTING, AzureField.NODESETTING_ENABLE_AUDIT_LOGGING, payload.enableAuditLogging);

            this.storeFieldString(WizardForm.OSIMAGE, OsImageField.IMAGE, payload.os.name);

            this.saveCommonFieldsFromPayload(payload);

            AppServices.userDataService.updateWizardTimestamp(this.wizardName);
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
        return this.getFieldValue(AzureForm.NODESETTING, AzureField.NODESETTING_MANAGEMENT_CLUSTER_NAME);
    }

    saveMCName(clusterName: string) {
        this.storeFieldString(AzureForm.NODESETTING, AzureField.NODESETTING_MANAGEMENT_CLUSTER_NAME, clusterName);
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
        const resourceGroupOption = this.getFieldValue(AzureForm.PROVIDER, AzureField.PROVIDER_RESOURCEGROUPOPTION);
        const azureResourceGroup = resourceGroupOption === ResourceGroupOption.EXISTING ? AzureField.PROVIDER_RESOURCEGROUPEXISTING :
            AzureField.PROVIDER_RESOURCEGROUPCUSTOM;
        const vnetOption = this.getFieldValue(AzureForm.VNET, AzureField.VNET_EXISTING_OR_CUSTOM);
        const azureVnetName = vnetOption === VnetOptionType.EXISTING ? AzureField.VNET_EXISTING_NAME : AzureField.VNET_CUSTOM_NAME;
        const azureControlPlaneSubnetName = vnetOption === VnetOptionType.EXISTING ? AzureField.VNET_CONTROLPLANE_SUBNET_NAME :
            AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME;
        const azureNodeSubnetName = vnetOption === VnetOptionType.EXISTING ? AzureField.VNET_WORKER_SUBNET_NAME :
            AzureField.VNET_WORKER_NEWSUBNET_NAME;
        const fieldsMapping = {
            AZURE_RESOURCE_GROUP: [AzureForm.PROVIDER, azureResourceGroup],
            AZURE_VNET_RESOURCE_GROUP: [AzureForm.VNET, AzureField.VNET_RESOURCE_GROUP],
            AZURE_VNET_NAME: [AzureForm.VNET, azureVnetName],
            AZURE_VNET_CIDR: [AzureForm.VNET, AzureField.VNET_CUSTOM_CIDR],
            AZURE_CONTROL_PLANE_SUBNET_NAME: [AzureForm.VNET, azureControlPlaneSubnetName],
            AZURE_CONTROL_PLANE_SUBNET_CIDR: [AzureForm.VNET, AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR],
            AZURE_NODE_SUBNET_NAME: [AzureForm.VNET, azureNodeSubnetName],
            AZURE_NODE_SUBNET_CIDR: [AzureForm.VNET, AzureField.VNET_WORKER_NEWSUBNET_CIDR]
        }
        let fields = [];
        if (vnetOption === VnetOptionType.EXISTING) {
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

    importFileProcessClusterParams(event: TanzuEventType, nameFile: string, azureClusterParams: AzureRegionalClusterParams) {
        this.setFromPayload(azureClusterParams);
        this.resetToFirstStep();
        this.importService.publishImportSuccess(event, nameFile);
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
            eventSuccess: this.supplyFileImportedEvent(),
            eventFailure: this.supplyFileImportErrorEvent(),
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
        AppServices.messenger.subscribe<string>(TanzuEventType.AZURE_REGION_CHANGED, event => {
            const region = event.payload;
            if (region) {
                AppServices.dataServiceRegistrar.trigger([
                    TanzuEventType.AZURE_GET_RESOURCE_GROUPS,
                    TanzuEventType.AZURE_GET_INSTANCE_TYPES
                ], { location: region });
                AppServices.dataServiceRegistrar.trigger([TanzuEventType.AZURE_GET_OS_IMAGES]);
            } else {
                AppServices.dataServiceRegistrar.clear<AzureResourceGroup>(TanzuEventType.AZURE_GET_RESOURCE_GROUPS);
                AppServices.dataServiceRegistrar.clear<AzureInstanceType>(TanzuEventType.AZURE_GET_INSTANCE_TYPES);
                AppServices.dataServiceRegistrar.clear<AzureVirtualMachine>(TanzuEventType.AZURE_GET_OS_IMAGES);
            }
        });
    }

    private registerServices() {
        const wizard = this;
        AppServices.dataServiceRegistrar.register<AzureResourceGroup>(TanzuEventType.AZURE_GET_RESOURCE_GROUPS,
            (payload: {location: string}) => { return wizard.apiClient.getAzureResourceGroups(payload); },
            "Failed to retrieve resource groups for the particular region." );
        AppServices.dataServiceRegistrar.register<AzureInstanceType>(TanzuEventType.AZURE_GET_INSTANCE_TYPES,
            (payload: {location: string}) => { return wizard.apiClient.getAzureInstanceTypes(payload); },
            "Failed to retrieve Azure VM sizes" );
        AppServices.dataServiceRegistrar.register<AzureVirtualMachine>(TanzuEventType.AZURE_GET_OS_IMAGES,
            () => { return wizard.apiClient.getAzureOSImages(); },
            "Failed to retrieve list of OS images from the specified Azure Server." );
        AppServices.dataServiceRegistrar.register<AzureVirtualNetwork>(TanzuEventType.AZURE_GET_VNETS,
            (payload: {resourceGroupName: string, location: string}) => { return wizard.apiClient.getAzureVnets(payload)},
            "Failed to retrieve list of VNETs from the specified Azure Server." );
    }
}
