// Angular imports
import { Component, ElementRef, OnInit } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';
// Third party imports
import { Observable } from 'rxjs';
// App imports
import { APIClient } from '../../../swagger/api-client.service';
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import AppServices from "../../../shared/service/appServices";
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { ClusterPlan, WizardForm, WizardStep } from '../wizard/shared/constants/wizard.constants';
import { ExportService } from '../../../shared/service/export.service';
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';
import { ImportParams, ImportService } from "../../../shared/service/import.service";
import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from './../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import { LoadBalancerField } from '../wizard/shared/components/steps/load-balancer/load-balancer-step.fieldmapping';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { PROVIDERS, Providers } from '../../../shared/constants/app.constants';
import { ResourceStepComponent } from './resource-step/resource-step.component';
import { TanzuEventType } from '../../../shared/service/Messenger';
import { VsphereField, VsphereForm, VsphereNodeTypes } from './vsphere-wizard.constants';
import {
    VSphereDatastore,
    VSphereFolder,
    VSphereManagementObject,
    VSphereNetwork,
    VSphereResourcePool,
    VSphereVirtualMachine
} from '../../../swagger/models';
import { VsphereLoadBalancerStepComponent } from './load-balancer/vsphere-load-balancer-step.component';
import { VsphereNetworkStepComponent } from './vsphere-network-step/vsphere-network-step.component';
import { VsphereOsImageStepComponent } from './vsphere-os-image-step/vsphere-os-image-step.component';
import { VSphereProviderStepComponent } from './provider-step/vsphere-provider-step.component';
import { VsphereRegionalClusterParams } from 'src/app/swagger/models/vsphere-regional-cluster-params.model';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';

@Component({
    selector: 'app-wizard',
    templateUrl: './vsphere-wizard.component.html',
    styleUrls: ['./vsphere-wizard.component.scss'],
})
export class VSphereWizardComponent extends WizardBaseDirective implements OnInit {
    APP_ROUTES: Routes = APP_ROUTES;
    PROVIDERS: Providers = PROVIDERS;

    tkrVersion: Observable<string>;
    vsphereVersion: string;
    deploymentPending: boolean = false;
    disableDeployButton = false;

    constructor(
        private apiClient: APIClient,
        router: Router,
        private exportService: ExportService,
        private importService: ImportService,
        formBuilder: FormBuilder,
        titleService: Title,
        el: ElementRef) {

        super(router, el, titleService, formBuilder);

        this.provider = AppServices.appDataService.getProviderType();
        this.tkrVersion = AppServices.appDataService.getTkrVersion();
        AppServices.appDataService.getVsphereVersion().subscribe(version => {
            this.vsphereVersion = version ? version + ' ' : '';
        });
    }

    protected supplyFileImportedEvent(): TanzuEventType {
        return TanzuEventType.VSPHERE_CONFIG_FILE_IMPORTED;
    }

    protected supplyFileImportErrorEvent(): TanzuEventType {
        return TanzuEventType.VSPHERE_CONFIG_FILE_IMPORT_ERROR;
    }

    protected supplyWizardName(): string {
        return 'vSphereWizard';
    }

    protected supplyStepData(): FormDataForHTML[] {
        return [
            this.VsphereProviderForm,
            this.VsphereNodeSettingForm,
            this.VsphereLoadBalancerForm,
            this.MetadataForm,
            this.VsphereResourceForm,
            this.VsphereNetworkForm,
            this.IdentityForm,
            this.VsphereOsImageForm,
            this.CeipForm,
        ];
    }

    ngOnInit() {
        super.ngOnInit();

        this.titleService.setTitle(this.title ? this.title + ' vSphere' : 'vSphere');
        this.registerServices();
        this.subscribeToServices();
    }

    getPayload(): VsphereRegionalClusterParams {
        const payload: VsphereRegionalClusterParams = {};
        this.initPayloadWithCommons(payload);
        const mappings = [
            ['ipFamily', VsphereForm.PROVIDER, VsphereField.PROVIDER_IP_FAMILY],
            ['datacenter', VsphereForm.PROVIDER, VsphereField.PROVIDER_DATA_CENTER],
            ['ssh_key', VsphereForm.PROVIDER, VsphereField.PROVIDER_SSH_KEY],
            ['clusterName', VsphereForm.NODESETTING, VsphereField.NODESETTING_CLUSTER_NAME],
            ['controlPlaneFlavor', VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_SETTING],
            ['controlPlaneEndpoint', VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP],
            ['datastore', VsphereForm.RESOURCE, VsphereField.RESOURCE_DATASTORE],
            ['folder', VsphereForm.RESOURCE, VsphereField.RESOURCE_VMFOLDER],
            ['resourcePool', VsphereForm.RESOURCE, VsphereField.RESOURCE_POOL]
        ];
        mappings.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));
        payload.controlPlaneNodeType =
            this.getControlPlaneType(this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_SETTING));
        payload.workerNodeType = AppServices.appDataService.isModeClusterStandalone() ? payload.controlPlaneNodeType :
            this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE);
        payload.machineHealthCheckEnabled =
            this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED) === true;

        const vsphereCredentialsMappings = [
            ['host', VsphereForm.PROVIDER, VsphereField.PROVIDER_VCENTER_ADDRESS],
            ['password', VsphereForm.PROVIDER, VsphereField.PROVIDER_USER_PASSWORD],
            ['username', VsphereForm.PROVIDER, VsphereField.PROVIDER_USER_NAME],
            ['thumbprint', VsphereForm.PROVIDER, VsphereField.PROVIDER_THUMBPRINT]
        ];
        payload.vsphereCredentials = {};

        payload.enableAuditLogging = this.getBooleanFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_ENABLE_AUDIT_LOGGING);

        vsphereCredentialsMappings.forEach(attr => payload.vsphereCredentials[attr[0]] = this.getFieldValue(attr[1], attr[2]));
        payload.vsphereCredentials['insecure'] = this.getBooleanFieldValue(VsphereForm.PROVIDER, VsphereField.PROVIDER_CONNECTION_INSECURE);

        const endpointProvider = this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER);
        if (endpointProvider === KUBE_VIP) {
            payload.aviConfig['controlPlaneHaProvider'] = false;
        } else {
            payload.aviConfig['controlPlaneHaProvider'] = true;
        }
        payload.aviConfig['managementClusterVipNetworkName'] =
            this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_NAME);
        if (!payload.aviConfig['managementClusterVipNetworkName']) {
            payload.aviConfig['managementClusterVipNetworkName'] =
                this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.NETWORK_NAME);
        }
        payload.aviConfig['managementClusterVipNetworkCidr'] =
            this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_CIDR);
        if (!payload.aviConfig['managementClusterVipNetworkCidr']) {
            payload.aviConfig['managementClusterVipNetworkCidr'] =
                this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.NETWORK_CIDR)
        }

        return payload;
    }

    setFromPayload(payload: VsphereRegionalClusterParams) {
        const mappings = [
            ['ipFamily', VsphereForm.PROVIDER, VsphereField.PROVIDER_IP_FAMILY],
            ['datacenter', VsphereForm.PROVIDER, VsphereField.PROVIDER_DATA_CENTER],
            ['ssh_key', VsphereForm.PROVIDER, VsphereField.PROVIDER_SSH_KEY],
            ['clusterName', VsphereForm.NODESETTING, VsphereField.NODESETTING_CLUSTER_NAME],
            ['controlPlaneFlavor', VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_SETTING],
            ['controlPlaneEndpoint', VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP],
            ['datastore', VsphereForm.RESOURCE, VsphereField.RESOURCE_DATASTORE],
            ['folder', VsphereForm.RESOURCE, VsphereField.RESOURCE_VMFOLDER],
            ['resourcePool', VsphereForm.RESOURCE, VsphereField.RESOURCE_POOL]
        ];
        mappings.forEach(attr => this.storeFieldString(attr[1], attr[2], payload[attr[0]]));

        this.storeFieldString(VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_SETTING, payload.controlPlaneFlavor);
        if (payload.controlPlaneNodeType) {
            const instanceTypeField = payload.controlPlaneFlavor === 'prod' ? VsphereField.NODESETTING_INSTANCE_TYPE_PROD
                : VsphereField.NODESETTING_INSTANCE_TYPE_DEV;
            const vSphereNode = this.getVSphereNode(payload.controlPlaneNodeType);
            if (vSphereNode) {
                this.storeFieldString('vsphereNodeSettingForm', instanceTypeField, vSphereNode.id, vSphereNode.name);
            }
        }

        this.storeFieldBoolean(VsphereForm.NODESETTING, VsphereField.NODESETTING_ENABLE_AUDIT_LOGGING, payload.enableAuditLogging);
        this.storeFieldBoolean(VsphereForm.NODESETTING, VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED,
            payload.machineHealthCheckEnabled);

        const workerNodeType = payload.workerNodeType;
        if (workerNodeType) {
            const vSphereNode = this.getVSphereNode(workerNodeType);
            if (vSphereNode) {
                this.storeFieldString(VsphereForm.NODESETTING, VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, vSphereNode.id,
                    vSphereNode.name);
            }
        }

        if (payload.vsphereCredentials !== undefined) {
            const vsphereCredentialsMappings = [
                ['host', VsphereForm.PROVIDER, VsphereField.PROVIDER_VCENTER_ADDRESS],
                ['username', VsphereForm.PROVIDER, VsphereField.PROVIDER_USER_NAME],
                ['thumbprint', VsphereForm.PROVIDER, VsphereField.PROVIDER_THUMBPRINT]
            ];
            vsphereCredentialsMappings.forEach(attr => this.storeFieldString(attr[1], attr[2], payload.vsphereCredentials[attr[0]]));
            const decodedPassword = AppServices.appDataService.decodeBase64(payload.vsphereCredentials['password']);
            this.storeFieldString(VsphereForm.PROVIDER, VsphereField.PROVIDER_USER_PASSWORD, decodedPassword);
        }

        if (payload.aviConfig !== undefined) {
            const endpointProvider = payload.aviConfig['controlPlaneHaProvider'] ? NSX_ADVANCED_LOAD_BALANCER : KUBE_VIP;
            this.storeFieldString(VsphereForm.NODESETTING, VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER, endpointProvider);
            // Set (or clear) the network name (based on whether it's different from the aviConfig value
            const managementClusterVipNetworkName = payload.aviConfig['managementClusterVipNetworkName'];
            let uiMcNetworkName = '';
            if (managementClusterVipNetworkName !== payload.aviConfig.network.name) {
                uiMcNetworkName = payload.aviConfig['managementClusterVipNetworkName'];
            }

            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_NAME, uiMcNetworkName);
            // Set (or clear) the CIDR setting (based on whether it's different from the aviConfig value
            const managementClusterNetworkCIDR = payload.aviConfig['managementClusterVipNetworkCidr'];
            let uiMcCidr = '';
            if (managementClusterNetworkCIDR !== payload.aviConfig.network.cidr) {
                uiMcCidr = managementClusterNetworkCIDR;
            }
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_CIDR, uiMcCidr)
        }

        this.storeFieldString('osImageForm', 'osImage', payload.os.moid, payload.os.name);

        this.saveCommonFieldsFromPayload(payload);
        AppServices.userDataService.updateWizardTimestamp(this.wizardName);
    }

    private getVSphereNode(workerNodeType: string) {
        return VsphereNodeTypes.find(node => node.id === workerNodeType);
    }

    /**
     * @method method to trigger deployment
     */
    createRegionalCluster(payload: any): Observable<any> {
        return this.apiClient.createVSphereRegionalCluster(payload);
    }

    /**
     * Return management/standalone cluster name
     */
    getMCName() {
        return this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_CLUSTER_NAME);
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

    /**
     * Apply the settings captured via UI to backend TKG config without
     * actually creating the management/standalone cluster.
     */
    applyTkgConfig() {
        return this.apiClient.applyTKGConfigForVsphere({ params: this.getPayload() });
    }

    /**
     * Retrieve the config file from the backend and return as a string
     */
    retrieveExportFile() {
        return this.apiClient.exportTKGConfigForVsphere({ params: this.getPayload() });
    }

    exportConfiguration() {
        const wizard = this;    // capture 'this' outside the context of the closure below
        this.exportService.export(
            this.retrieveExportFile(),
            (failureMessage) => { wizard.displayError(failureMessage); }
        );
    }

    /**
     * @method getControlPlaneType
     * helper method to return value of dev instance type or prod instance type
     * depending on what type of control plane is selected
     * @param controlPlaneType {string} the control plane type (dev/prod)
     * @returns {any}
     */
    getControlPlaneType(controlPlaneType: string) {
        if (controlPlaneType === ClusterPlan.DEV) {
            return this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_INSTANCE_TYPE_DEV);
        } else if (controlPlaneType === ClusterPlan.PROD) {
            return this.getFieldValue(VsphereForm.NODESETTING, VsphereField.NODESETTING_INSTANCE_TYPE_PROD);
        } else {
            return null;
        }
    }

    // vSphere-specific forms
    get VsphereLoadBalancerForm(): FormDataForHTML {
        return { name: WizardForm.LOADBALANCER, title: 'VMware NSX Advanced Load Balancer',
            description: VsphereLoadBalancerStepComponent.description,
            i18n: { title: 'load balancer step name', description: 'load balancer step description' },
        clazz: VsphereLoadBalancerStepComponent };
    }
    get VsphereNodeSettingForm(): FormDataForHTML {
        return { name: VsphereForm.NODESETTING, title: FormUtility.titleCase(this.clusterTypeDescriptor) + ' Cluster Settings',
            description: `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`,
            i18n: { title: 'node setting step name', description: 'node setting step description' },
        clazz: NodeSettingStepComponent };
    }
    get VsphereNetworkForm(): FormDataForHTML {
        return  FormUtility.formWithOverrides(super.NetworkForm,
            { clazz: VsphereNetworkStepComponent, description: VsphereNetworkStepComponent.description});
    }
    get VsphereProviderForm(): FormDataForHTML {
        return { name: VsphereForm.PROVIDER, title: 'IaaS Provider',
            description: 'Validate the vSphere ' + this.vsphereVersion + 'provider account for Tanzu',
            i18n: { title: 'IaaS provider step name', description: 'IaaS provider step description' },
        clazz: VSphereProviderStepComponent };
    }
    get VsphereResourceForm(): FormDataForHTML {
        return { name: VsphereForm.RESOURCE, title: 'Resources',
            description: `Specify the resources for this ${this.clusterTypeDescriptor} cluster`,
            i18n: { title: 'Resource step name', description: 'Resource step description' },
        clazz: ResourceStepComponent};
    }
    get VsphereOsImageForm(): FormDataForHTML {
        return this.getOsImageForm(VsphereOsImageStepComponent);
    }

    // returns TRUE if the file contents appear to be a valid config file for vSphere
    // returns FALSE if the file is empty or does not appear to be valid. Note that in the FALSE
    // case we also alert the user.
    importFileValidate(nameFile: string, fileContents: string): boolean {
        if (fileContents.includes('VSPHERE_')) {
            return true;
        }
        alert(nameFile + ' is not a valid vSphere configuration file!');
        return false;
    }

    importFileRetrieveClusterParams(fileContents) {
        return this.apiClient.importTKGConfigForVsphere( { params: { filecontents: fileContents } } );
    }

    importFileProcessClusterParams(event: TanzuEventType, nameFile: string, vsphereClusterParams: VsphereRegionalClusterParams) {
        this.setFromPayload(vsphereClusterParams);
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
        const params: ImportParams<VsphereRegionalClusterParams> = {
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
        AppServices.messenger.subscribe<string>(TanzuEventType.VSPHERE_DATACENTER_CHANGED, event => {
            AppServices.dataServiceRegistrar.trigger( [
                TanzuEventType.VSPHERE_GET_RESOURCE_POOLS,
                TanzuEventType.VSPHERE_GET_COMPUTE_RESOURCE,
                TanzuEventType.VSPHERE_GET_VM_NETWORKS,
                TanzuEventType.VSPHERE_GET_DATA_STORES,
                TanzuEventType.VSPHERE_GET_VM_FOLDERS,
                TanzuEventType.VSPHERE_GET_OS_IMAGES
                ], { dc: event.payload });
            });
    }

    private registerServices() {
        const wizard = this;
        AppServices.dataServiceRegistrar.register<VSphereResourcePool>(TanzuEventType.VSPHERE_GET_RESOURCE_POOLS,
            (payload: { dc: string }) => { return wizard.apiClient.getVSphereResourcePools(payload) },
            "Failed to retrieve list of resource pools from the specified vCenter Server."
            );
        AppServices.dataServiceRegistrar.register<VSphereManagementObject>(TanzuEventType.VSPHERE_GET_COMPUTE_RESOURCE,
            (payload: { dc: string }) => { return wizard.apiClient.getVSphereComputeResources(payload) },
            "Failed to retrieve list of compute resources from the specified datacenter."
        );
        AppServices.dataServiceRegistrar.register<VSphereNetwork>(TanzuEventType.VSPHERE_GET_VM_NETWORKS,
            (payload: { dc: string }) => { return wizard.apiClient.getVSphereNetworks(payload) },
            "Failed to retrieve list of VM networks from the specified vCenter Server."
        );
        AppServices.dataServiceRegistrar.register<VSphereDatastore>(TanzuEventType.VSPHERE_GET_DATA_STORES,
            (payload: { dc: string }) => { return wizard.apiClient.getVSphereDatastores(payload) },
            "Failed to retrieve list of datastores from the specified vCenter Server."
        );
        AppServices.dataServiceRegistrar.register<VSphereFolder>(TanzuEventType.VSPHERE_GET_VM_FOLDERS,
            (payload: { dc: string }) => { return wizard.apiClient.getVSphereFolders(payload) },
            "Failed to retrieve list of vm folders from the specified vCenter Server."
        );
        AppServices.dataServiceRegistrar.register<VSphereVirtualMachine>(TanzuEventType.VSPHERE_GET_OS_IMAGES,
            (payload: { dc: string }) => { return wizard.apiClient.getVSphereOSImages(payload) },
            "Failed to retrieve list of OS images from the specified vCenter Server."
        );
    }
}
