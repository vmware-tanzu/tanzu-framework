import { KUBE_VIP } from './../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
// Angular imports
import { Component, OnInit, ElementRef, AfterViewInit, ViewChild } from '@angular/core';
import { FormGroup, FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

// Third party imports
import { Observable } from 'rxjs';

// App imports
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import { APIClient } from '../../../swagger/api-client.service';
import { PROVIDERS, Providers } from '../../../shared/constants/app.constants';
import { AppDataService } from '../../../shared/service/app-data.service';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { CliFields, CliGenerator } from '../wizard/shared/utils/cli-generator';
import { WizardBaseDirective } from '../wizard/shared/wizard-base/wizard-base';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';
import { VsphereRegionalClusterParams } from 'src/app/swagger/models/vsphere-regional-cluster-params.model';

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
    deploymentPending: boolean = false;
    disableDeployButton = false;

    show = false;

    constructor(
        private apiClient: APIClient,
        router: Router,
        public wizardFormService: VSphereWizardFormService,
        private appDataService: AppDataService,
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
            registerTmcForm: this.formBuilder.group({
            }),
            ceipOptInForm: this.formBuilder.group({
            }),
            identityForm: this.formBuilder.group({
            })
        });

        this.provider = this.appDataService.getProviderType();
        this.tkrVersion = this.appDataService.getTkrVersion();
    }

    ngOnInit() {
        super.ngOnInit();

        // delay showing first panel to avoid panel not defined console err
        setTimeout(_ => {
            this.show = true;
        }, 100)

        this.titleService.setTitle(this.title + ' vSphere');
    }

    getStepDescription(stepName: string): string {
        if (stepName === 'provider') {
            if (this.getFieldValue('vsphereProviderForm', 'vcenterAddress') &&
                this.getFieldValue('vsphereProviderForm', 'datacenter')) {
                return 'vCenter ' + this.getFieldValue('vsphereProviderForm', 'vcenterAddress') + ' connected';
            } else {
                return 'Validate the vSphere provider account for Tanzu Kubernetes Grid';
            }
        } else if (stepName === 'nodeSetting') {
            if (this.getFieldValue('vsphereNodeSettingForm', 'controlPlaneSetting')) {
                let mode = 'Development cluster selected: 1 node control plane';
                if (this.getFieldValue('vsphereNodeSettingForm', 'controlPlaneSetting') === 'prod') {
                    mode = 'Production cluster selected: 3 node control plane';
                }
                return mode;
            } else {
                return `Specify the resources backing the ${this.clusterType} cluster`;
            }
        } else if (stepName === 'resource') {
            if (this.getFieldValue('resourceForm', 'vmFolder') &&
                this.getFieldValue('resourceForm', 'datastore') &&
                this.getFieldValue('resourceForm', 'resourcePool')) {
                return 'Resource Pool: ' + this.getFieldValue('resourceForm', 'resourcePool') +
                    ', VM Folder: ' + this.getFieldValue('resourceForm', 'vmFolder') +
                    ', Datastore: ' + this.getFieldValue('resourceForm', 'datastore');
            } else {
                return `Specify the resources for this ${this.clusterType}} cluster`;
            }
        } else if (stepName === 'network') {
            if (this.getFieldValue('networkForm', 'networkName')) {
                return 'Network: ' + this.getFieldValue('networkForm', 'networkName');
            } else {
                return 'Specify how Tanzu Kubernetes Grid networking is provided and any global network settings';
            }
        } else if (stepName === 'loadBalancer') {
            if (this.getFieldValue('loadBalancerForm', 'controllerHost')) {
                return 'Controller: ' + this.getFieldValue('loadBalancerForm', 'controllerHost');
            } else {
                const endpointProvider = this.getFieldValue("vsphereNodeSettingForm", "controlPlaneEndpointProvider");
                if (endpointProvider === KUBE_VIP) {
                    return 'Optionally specify VMware NSX Advanced Load Balancer settings';
                } else {
                    return 'Specify VMware NSX Advanced Load Balancer settings';
                }
            }
        } else if (stepName === 'osImage') {
            if (this.getFieldValue('osImageForm', 'osImage') && this.getFieldValue('osImageForm', 'osImage').name) {
                return 'OS Image: ' + (this.getFieldValue('osImageForm', 'osImage').name);
            } else {
                return 'Specify the OS Image';
            }
        } else if (stepName === 'metadata') {
            if (this.getFieldValue('metadataForm', 'clusterLocation')) {
                return 'Location: ' + this.getFieldValue('metadataForm', 'clusterLocation');
            } else {
                return `Specify metadata for the ${this.clusterType} cluster`;
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
        }
    }

    getPayload(): VsphereRegionalClusterParams {
        const payload: VsphereRegionalClusterParams = {};
        this.initPayloadWithCommons(payload);
        const mappings = [
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
        payload.workerNodeType = (this.clusterType !== 'standalone') ?
            this.getFieldValue('vsphereNodeSettingForm', 'workerNodeInstanceType') : payload.controlPlaneNodeType;
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
            clusterType: this.clusterType
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

}
