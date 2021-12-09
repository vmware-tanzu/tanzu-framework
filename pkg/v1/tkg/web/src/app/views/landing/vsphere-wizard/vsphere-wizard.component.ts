import { KUBE_VIP, NSX_ADVANCED_LOAD_BALANCER } from './../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
// Angular imports
import { Component, OnInit, ElementRef } from '@angular/core';
import { FormGroup, FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

// Third party imports
import { Observable } from 'rxjs';

// App imports
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import { APIClient } from '../../../swagger/api-client.service';
import { PROVIDERS, Providers } from '../../../shared/constants/app.constants';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
import { VsphereRegionalClusterParams } from 'src/app/swagger/models/vsphere-regional-cluster-params.model';
import Broker from "../../../shared/service/broker";
import { WizardForm, WizardStep } from '../wizard/shared/constants/wizard.constants';
import { WizardStep } from '../wizard/shared/constants/wizard.constants';
import { StepUtility } from '../wizard/shared/components/steps/step-utility';
import { ImportParams, ImportService } from "../../../shared/service/import.service";
import { VsphereField } from './vsphere-wizard.constants';
import { FormDataForHTML, FormUtility } from '../wizard/shared/components/steps/form-utility';

@Component({
    selector: 'app-wizard',
    templateUrl: './vsphere-wizard.component.html',
    styleUrls: ['./vsphere-wizard.component.scss'],
})
export class VSphereWizardComponent extends WizardBaseDirective implements OnInit {
    APP_ROUTES: Routes = APP_ROUTES;
    PROVIDERS: Providers = PROVIDERS;

    datacenterMoid: Observable<string>;
    tkrVersion: Observable<string>;
    vsphereVersion: string;
    deploymentPending: boolean = false;
    disableDeployButton = false;

    show = false;

    constructor(
        private apiClient: APIClient,
        router: Router,
        public wizardFormService: VSphereWizardFormService,
        private importService: ImportService,
        private formBuilder: FormBuilder,
        formMetaDataService: FormMetaDataService,
        titleService: Title,
        el: ElementRef) {

        super(router, el, formMetaDataService, titleService);

        this.form = this.formBuilder.group({
            vsphereProviderForm: this.formBuilder.group({
            }),
            vsphereNodeSettingForm: this.formBuilder.group({
            }),
            metadataForm: this.formBuilder.group({
            }),
            resourceForm: this.formBuilder.group({
            }),
            networkForm: this.formBuilder.group({
            }),
            loadBalancerForm: this.formBuilder.group({
            }),
            osImageForm: this.formBuilder.group({
            }),
            ceipOptInForm: this.formBuilder.group({
            }),
            identityForm: this.formBuilder.group({
            })
        });

        this.provider = Broker.appDataService.getProviderType();
        this.tkrVersion = Broker.appDataService.getTkrVersion();
        Broker.appDataService.getVsphereVersion().subscribe(version => {
            this.vsphereVersion = version ? version + ' ' : '';
        });
    }

    ngOnInit() {
        super.ngOnInit();

        // delay showing first panel to avoid panel not defined console err
        setTimeout(_ => {
            this.show = true;
        }, 100)

        this.titleService.setTitle(this.title + ' vSphere');
    }

    getPayload(): VsphereRegionalClusterParams {
        const payload: VsphereRegionalClusterParams = {};
        this.initPayloadWithCommons(payload);
        const mappings = [
            ['ipFamily', 'vsphereProviderForm', 'ipFamily'],
            ['datacenter', 'vsphereProviderForm', 'datacenter'],
            ['ssh_key', 'vsphereProviderForm', 'ssh_key'],
            ['clusterName', 'vsphereNodeSettingForm', 'clusterName'],
            ['controlPlaneFlavor', 'vsphereNodeSettingForm', 'controlPlaneSetting'],
            ['controlPlaneEndpoint', 'vsphereNodeSettingForm', 'controlPlaneEndpointIP'],
            ['datastore', 'resourceForm', 'datastore'],
            ['folder', 'resourceForm', 'vmFolder'],
            ['resourcePool', 'resourceForm', 'resourcePool']
        ];
        mappings.forEach(attr => payload[attr[0]] = this.getFieldValue(attr[1], attr[2]));
        payload.controlPlaneNodeType = this.getControlPlaneType(this.getFieldValue('vsphereNodeSettingForm', 'controlPlaneSetting'));
        payload.workerNodeType = Broker.appDataService.isModeClusterStandalone() ? payload.controlPlaneNodeType :
            this.getFieldValue('vsphereNodeSettingForm', VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE);
        payload.machineHealthCheckEnabled = this.getFieldValue("vsphereNodeSettingForm", "machineHealthChecksEnabled") === true;

        const vsphereCredentialsMappings = [
            ['host', 'vsphereProviderForm', 'vcenterAddress'],
            ['password', 'vsphereProviderForm', 'password'],
            ['username', 'vsphereProviderForm', 'username'],
            ['thumbprint', 'vsphereProviderForm', 'thumbprint']
        ];
        payload.vsphereCredentials = {};

        payload.enableAuditLogging = this.getBooleanFieldValue("vsphereNodeSettingForm", "enableAuditLogging");

        vsphereCredentialsMappings.forEach(attr => payload.vsphereCredentials[attr[0]] = this.getFieldValue(attr[1], attr[2]));
        payload.vsphereCredentials['insecure'] = this.getBooleanFieldValue('vsphereProviderForm', 'insecure');

        const endpointProvider = this.getFieldValue("vsphereNodeSettingForm", "controlPlaneEndpointProvider");
        if (endpointProvider === KUBE_VIP) {
            payload.aviConfig['controlPlaneHaProvider'] = false;
        } else {
            payload.aviConfig['controlPlaneHaProvider'] = true;
        }
        payload.aviConfig['managementClusterVipNetworkName'] = this.getFieldValue("loadBalancerForm", "managementClusterNetworkName");
        if (!payload.aviConfig['managementClusterVipNetworkName']) {
            payload.aviConfig['managementClusterVipNetworkName'] = this.getFieldValue('loadBalancerForm', 'networkName');
        }
        payload.aviConfig['managementClusterVipNetworkCidr'] = this.getFieldValue("loadBalancerForm", "managementClusterNetworkCIDR");
        if (!payload.aviConfig['managementClusterVipNetworkCidr']) {
            payload.aviConfig['managementClusterVipNetworkCidr'] = this.getFieldValue('loadBalancerForm', 'networkCIDR')
        }

        return payload;
    }

    setFromPayload(payload: VsphereRegionalClusterParams) {
        const mappings = [
            ['ipFamily', 'vsphereProviderForm', 'ipFamily'],
            ['datacenter', 'vsphereProviderForm', 'datacenter'],
            ['ssh_key', 'vsphereProviderForm', 'ssh_key'],
            ['clusterName', 'vsphereNodeSettingForm', 'clusterName'],
            ['controlPlaneFlavor', 'vsphereNodeSettingForm', 'controlPlaneSetting'],
            ['controlPlaneEndpoint', 'vsphereNodeSettingForm', 'controlPlaneEndpointIP'],
            ['datastore', 'resourceForm', 'datastore'],
            ['folder', 'resourceForm', 'vmFolder'],
            ['resourcePool', 'resourceForm', 'resourcePool']
        ];
        mappings.forEach(attr => this.saveFormField(attr[1], attr[2], payload[attr[0]]));

        this.saveControlPlaneFlavor('vsphere', payload.controlPlaneFlavor);
        this.saveControlPlaneNodeType('vsphere', payload.controlPlaneFlavor, payload.controlPlaneNodeType);

        this.saveFormField("vsphereNodeSettingForm", VsphereField.NODESETTING_ENABLE_AUDIT_LOGGING, payload.enableAuditLogging);
        this.saveFormField("vsphereNodeSettingForm", VsphereField.NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED,
            payload.machineHealthCheckEnabled);
        this.saveFormListbox('vsphereNodeSettingForm', VsphereField.NODESETTING_WORKER_NODE_INSTANCE_TYPE, payload.workerNodeType);

        if (payload.vsphereCredentials !== undefined) {
            const vsphereCredentialsMappings = [
                ['host', 'vsphereProviderForm', 'vcenterAddress'],
                ['username', 'vsphereProviderForm', 'username'],
                ['thumbprint', 'vsphereProviderForm', 'thumbprint']
            ];
            vsphereCredentialsMappings.forEach(attr => this.saveFormField(attr[1], attr[2], payload.vsphereCredentials[attr[0]]));
            const decodedPassword = Broker.appDataService.decodeBase64(payload.vsphereCredentials['password']);
            this.saveFormField('vsphereProviderForm', 'password', decodedPassword);
        }

        if (payload.aviConfig !== undefined) {
            const endpointProvider = payload.aviConfig['controlPlaneHaProvider'] ? NSX_ADVANCED_LOAD_BALANCER : KUBE_VIP;
            this.saveFormField('vsphereNodeSettingForm', 'controlPlaneEndpointProvider', endpointProvider);
            // Set (or clear) the network name (based on whether it's different from the aviConfig value
            const managementClusterVipNetworkName = payload.aviConfig['managementClusterVipNetworkName'];
            let uiMcNetworkName = '';
            if (managementClusterVipNetworkName !== payload.aviConfig.network.name) {
                uiMcNetworkName = payload.aviConfig['managementClusterVipNetworkName'];
            }

            this.saveFormField("loadBalancerForm", "managementClusterNetworkName", uiMcNetworkName);
            // Set (or clear) the CIDR setting (based on whether it's different from the aviConfig value
            const managementClusterNetworkCIDR = payload.aviConfig['managementClusterVipNetworkCidr'];
            let uiMcCidr = '';
            if (managementClusterNetworkCIDR !== payload.aviConfig.network.cidr) {
                uiMcCidr = managementClusterNetworkCIDR;
            }
            this.saveFormField("loadBalancerForm", "managementClusterNetworkCIDR", uiMcCidr)
        }

        this.saveCommonFieldsFromPayload(payload);
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
        return this.getFieldValue('vsphereNodeSettingForm', 'clusterName');
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

    /**
     * @method getControlPlaneType
     * helper method to return value of dev instance type or prod instance type
     * depending on what type of control plane is selected
     * @param controlPlaneType {string} the control plane type (dev/prod)
     * @returns {any}
     */
    getControlPlaneType(controlPlaneType: string) {
        if (controlPlaneType === 'dev') {
            return this.getFieldValue('vsphereNodeSettingForm', 'devInstanceType');
        } else if (controlPlaneType === 'prod') {
            return this.getFieldValue('vsphereNodeSettingForm', 'prodInstanceType');
        } else {
            return null;
        }
    }


    // HTML convenience methods
    //
    // OVERRIDES
    private get NetworkFormDescription(): string {
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        // because vSphere allows for the user to select a network name
        const networkName = this.getFieldValue(WizardForm.NETWORK, 'networkName');
        if (networkName) {
            return 'Network: ' + networkName;
        }
        return 'Specify how Tanzu Kubernetes Grid networking is provided and any global network settings';
    }
    get NetworkForm(): FormDataForHTML {
        return FormUtility.formOverrideDescription(this.NetworkForm, this.NetworkFormDescription);
    }

    // Vsphere-specific
    private get LoadBalancerFormDescription(): string { // TODO: this should be overriding base class implementation
        // NOTE: even though this is a common wizard form, vSphere has a different way of describing it
        const controllerHost = this.getFieldValue('loadBalancerForm', 'controllerHost');
        if (controllerHost) {
            return 'Controller: ' + controllerHost;
        }
        const endpointProvider = this.getFieldValue("vsphereNodeSettingForm", "controlPlaneEndpointProvider");
        if (endpointProvider === KUBE_VIP) {
            return 'Optionally specify VMware NSX Advanced Load Balancer settings';
        }
        return 'Specify VMware NSX Advanced Load Balancer settings';
    }
    get VsphereLoadBalancerForm(): FormDataForHTML {
        return { name: WizardForm.LOADBALANCER, title: 'VMware NSX Advanced Load Balancer', description: this.LoadBalancerFormDescription,
            i18n: { title: 'load balancer step name', description: 'load balancer step description' } };
    }
    get VsphereNodeSettingForm(): FormDataForHTML {
        return { name: 'vsphereNodeSettingForm', title: FormUtility.titleCase(this.clusterTypeDescriptor) + ' Cluster Settings',
            description: this.VsphereNodeSettingFormDescription,
            i18n: { title: 'node setting step name', description: 'node setting step description' } };
    }
    private get VsphereNodeSettingFormDescription(): string {
        if (this.getFieldValue('vsphereNodeSettingForm', 'controlPlaneSetting')) {
            let mode = 'Development cluster selected: 1 node control plane';
            if (this.getFieldValue('vsphereNodeSettingForm', 'controlPlaneSetting') === 'prod') {
                mode = 'Production cluster selected: 3 node control plane';
            }
            return mode;
        }
        return `Specify the resources backing the ${this.clusterTypeDescriptor} cluster`;
    }
    get VsphereProviderForm(): FormDataForHTML {
        return { name: 'vsphereProviderForm', title: 'IaaS Provider', description: this.VsphereProviderFormDescription,
            i18n: { title: 'IaaS provider step name', description: 'IaaS provider step description' } };
    }
    private get VsphereProviderFormDescription(): string {
        const vcenterIP = this.getFieldValue('vsphereProviderForm', 'vcenterAddress');
        const datacenter = this.getFieldValue('vsphereProviderForm', 'datacenter');
        if ( vcenterIP && datacenter) {
            return 'vCenter ' + vcenterIP + ' connected';
        }
        return 'Validate the vSphere ' + this.vsphereVersion + 'provider account for Tanzu';
    }
    get VsphereResourceForm(): string {
        return 'resourceForm';
    }
    get VsphereResourceFormDescription(): string {
        const vmFolder = this.getFieldValue('resourceForm', 'vmFolder');
        const datastore = this.getFieldValue('resourceForm', 'datastore');
        const resourcePool = this.getFieldValue('resourceForm', 'resourcePool');
        if (vmFolder && datastore && resourcePool) {
            return 'Resource Pool: ' + resourcePool + ', VM Folder: ' + vmFolder + ', Datastore: ' + datastore;
        }
        return `Specify the resources for this ${this.clusterTypeDescriptor} cluster`;
    }
    //
    // HTML convenience methods
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

    importFileProcessClusterParams(nameFile: string, vsphereClusterParams: VsphereRegionalClusterParams) {
        this.setFromPayload(vsphereClusterParams);
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
        const params: ImportParams<VsphereRegionalClusterParams> = {
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
}
