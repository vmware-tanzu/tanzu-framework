// Angular imports
import { OnInit, ElementRef, AfterViewInit, ViewChild, Directive } from '@angular/core';
import { FormGroup } from '@angular/forms';
import { Router } from '@angular/router';

// Third party imports
import { Observable } from 'rxjs';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { APP_ROUTES, Routes } from 'src/app/shared/constants/routes.constants';
import { Providers, PROVIDERS } from 'src/app/shared/constants/app.constants';
import { FormMetaDataStore } from '../FormMetaDataStore';
import { debounceTime, takeUntil } from 'rxjs/operators';
import { TkgEventType } from './../../../../../shared/service/Messenger';
import { ClrStepper } from '@clr/angular';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { ConfigFileInfo } from '../../../../../swagger/models/config-file-info.model';
import Broker from 'src/app/shared/service/broker';

@Directive()
export abstract class WizardBaseDirective extends BasicSubscriber implements AfterViewInit, OnInit {

    APP_ROUTES: Routes = APP_ROUTES;
    PROVIDERS: Providers = PROVIDERS;

    @ViewChild('wizard', { read: ClrStepper, static: true })
    wizard: ClrStepper;

    form: FormGroup;
    errorNotification: string;
    provider: Observable<string>;
    providerType: string;
    deploymentPending: boolean = false;
    disableDeployButton = false;

    steps = [true, false, false, false, false, false, false, false, false, false, false];
    review = false;

    constructor(
        protected router: Router,
        protected el: ElementRef,
        protected formMetaDataService: FormMetaDataService
    ) {

        super();
    }

    ngOnInit() {
        // work around an issue within StepperModel
        this.wizard['stepperService']['accordion']['openFirstPanel'] = function () {
            const firstPanel = this.getFirstPanel();
            if (firstPanel) {
                this._panels[firstPanel.id].open = true;
                this._panels[firstPanel.id].disabled = true;
            }
        }
        this.watchFieldsChange();

        FormMetaDataStore.resetStepList();
        FormMetaDataStore.resetFormList();
    }

    ngAfterViewInit(): void {
        this.getStepMetadata();
    }

    watchFieldsChange() {
        const formNames = Object.keys(this.form.controls);
        formNames.forEach((formName) => {
            this.form.controls[formName].valueChanges.pipe(debounceTime(200)).subscribe(() => {
                if (this.form.controls[formName].status === 'VALID') {
                    this.formMetaDataService.saveFormMetadata(formName,
                        this.el.nativeElement.querySelector(`clr-stepper-panel[formgroupname=${formName}]`));
                }
            });
        });
    }
    /**
     * Collect step meta data (title, description etc.) for all steps
     */
    getStepMetadata() {
        let wizard = this.el.nativeElement;
        wizard = wizard.querySelector('form[clrstepper]');
        const panels: any[] = Array.from(wizard.querySelectorAll('clr-stepper-panel'));
        const stepMetadataList = [];
        panels.forEach((panel => {
            const stepMetadata = {};
            const title = panel.querySelector('clr-step-title');
            if (title) {
                stepMetadata['title'] = title.innerText;
            }
            const description = panel.querySelector('clr-step-description');
            if (description) {
                stepMetadata['description'] = description.innerText;
            }
            stepMetadataList.push(stepMetadata);
        }));
        FormMetaDataStore.setStepList(stepMetadataList);
    }

    /**
     * @method getControlPlaneType
     * helper method to return value of dev instance type or prod instance type
     * depending on what type of control plane is selected
     * @param controlPlaneType {string} the control plane type (dev/prod)
     * @returns {any}
     */
    getControlPlaneNodeType(provider: string) {
        const controlPlaneType = this.getFieldValue(`${provider}NodeSettingForm`, 'controlPlaneSetting');
        if (controlPlaneType === 'dev') {
            return this.getFieldValue(`${provider}NodeSettingForm`, 'devInstanceType');
        } else if (controlPlaneType === 'prod') {
            return this.getFieldValue(`${provider}NodeSettingForm`, 'prodInstanceType');
        } else {
            return null;
        }
    }

    getControlPlaneFlavor(provider: string) {
        return this.getFieldValue(`${provider}NodeSettingForm`, 'controlPlaneSetting');
    }

    /**
     * Apply the settings captured via UI to backend TKG config without
     * actually creating the management cluster.
     */
    abstract applyTkgConfig(): Observable<ConfigFileInfo>;

    /**
     * Switch the mode between "Review Configuration" and "Edit Configuration"
     * @param review In "Review Configuration" mode if true; otherwise in "Edit Configuration" mode
     */
    reviewConfiguration(review) {
        this.disableDeployButton = false;
        this.applyTkgConfig()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                (data) => {
                    this.updateCli(data.path); // Generate CLI based on latest settings
                },
                err => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification = `Failed to apply tkg config. ${error}`;
                }
            );
        this.review = review;
    }

    getWizardValidity(): boolean {
        if (!FormMetaDataStore.getStepList()) {
            return false;
        }
        const totalSteps = FormMetaDataStore.getStepList().length;
        const stepsVisisted = this.steps.filter(step => step).length;
        return stepsVisisted > totalSteps && this.form.status === 'VALID';
    }

    /**
     * @method method to trigger deployment
     */
    abstract createRegionalCluster(params: any): Observable<any>;
    abstract getPayload(): any;

    deploy(): void {
        this.deploymentPending = true;
        this.disableDeployButton = true;
        const params = this.getPayload();
        this.createRegionalCluster({
            params: params
        })
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(
                (() => {
                    this.navigate(APP_ROUTES.WIZARD_PROGRESS);
                }),
                ((err) => {
                    const error = err.error.message || err.message || JSON.stringify(err);
                    this.errorNotification = `Failed to initiate cluster deployment. ${error}`;
                    this.deploymentPending = false;
                    this.disableDeployButton = false;
                })
            );
    }

    /**
     * @method navigate
     * @desc helper method to trigger router navigation to specified route
     * @param route - the route to navigate to
     */
    navigate(route: string): void {
        this.router.navigate([route]);
    }

    /**
     * @method navigateToWizard
     * @desc helper method to trigger router navigation to wizard
     * @param route - the route to navigate to
     */
    navigateToWizard(route: string): void {
        this.router.navigate([route]);
    }

    /**
     * Set the next step to be rendered. In initial wizard walkthrouh,
     * each step content is rendered sequentially, but in subsequent walkthrough,
     * a.k.a. "Edit Configuration" mode, each step widget is no longer re-created,
     * and therefore it reuses its previous component and form states.
     */
    onNextStep() {
        for (let i = 0; i < this.steps.length; i++) {
            if (!this.steps[i]) {
                this.steps[i] = true;
                break;
            }
        }
        this.getStepMetadata();
    }

    /**
     * Return the current value of the specified field
     * @param formName the form to get the field from
     * @param fieldName the name of the field to get
     */
    getFieldValue(formName, fieldName) {
        return this.form.get(formName) && this.form.get(formName).get(fieldName) && this.form.get(formName).get(fieldName).value || '';
    }

    /**
     * Return the field value as a boolean type
     * @param formName the form to get the field from
     * @param fieldName the name of the field to get
     */
    getBooleanFieldValue(formName, fieldName): boolean {
        return this.getFieldValue(formName, fieldName) ? true : false;
    }

    /**
     * Return CLI based on latest user input
     */
    abstract getCli(configPath: string): string;

    /**
     * Notify others that the CLI has changed
     */
    updateCli(configPath: string) {
        const cli = this.getCli(configPath);
        Broker.messenger.publish({
            type: TkgEventType.CLI_CHANGED,
            payload: cli
        });
    }

    /**
     * Converts ES6 map to stringifyable object
     * @param strMap ES6 map that will be converted
     */
    strMapToObj(strMap: Map<string, string>): { [key: string]: string; } {
        const obj = Object.create(null);
        for (const [k, v] of strMap) {
            obj[k] = v;
        }
        return obj;
    }

    /**
     * Fill in payload with values from all common steps
     * @param payload
     */
    initPayloadWithCommons(payload: any) {
        payload.networking = {
            networkName: this.getFieldValue('networkForm', 'networkName'),
            clusterDNSName: '',
            clusterNodeCIDR: '',
            clusterServiceCIDR: this.getFieldValue('networkForm', 'clusterServiceCidr'),
            clusterPodCIDR: this.getFieldValue('networkForm', 'clusterPodCidr'),
            cniType: this.getFieldValue('networkForm', 'cniType')
        };

        if (this.getFieldValue('networkForm', 'proxySettings')) {
            let proxySettingsMap = null;
            proxySettingsMap = [
                ['HTTPProxyURL', 'networkForm', 'httpProxyUrl'],
                ['HTTPProxyUsername', 'networkForm', 'httpProxyUsername'],
                ['HTTPProxyPassword', 'networkForm', 'httpProxyPassword'],
                ['noProxy', 'networkForm', 'noProxy']
            ];
            if (this.getFieldValue('networkForm', 'isSameAsHttp')) {
                proxySettingsMap = [
                    ...proxySettingsMap,
                    ['HTTPSProxyURL', 'networkForm', 'httpProxyUrl'],
                    ['HTTPSProxyUsername', 'networkForm', 'httpProxyUsername'],
                    ['HTTPSProxyPassword', 'networkForm', 'httpProxyPassword']
                ];
            } else {
                proxySettingsMap = [
                    ...proxySettingsMap,
                    ['HTTPSProxyURL', 'networkForm', 'httpsProxyUrl'],
                    ['HTTPSProxyUsername', 'networkForm', 'httpsProxyUsername'],
                    ['HTTPSProxyPassword', 'networkForm', 'httpsProxyPassword']
                ];
            }
            payload.networking.httpProxyConfiguration = {
                enabled: true
            };
            proxySettingsMap.forEach(attr => {
                let val = this.getFieldValue(attr[1], attr[2]);
                if (attr[0] === 'noProxy') {
                    val = val.replace(/\s/g, ''); // remove all spaces
                }
                payload.networking.httpProxyConfiguration[attr[0]] = val;
            });
        }

        payload.ceipOptIn = this.getFieldValue('ceipOptInForm', 'ceipOptIn') || false;
        payload.tmc_registration_url = this.getFieldValue('registerTmcForm', 'tmcRegUrl');
        payload.labels = this.strMapToObj(this.getFieldValue('metadataForm', 'clusterLabels'));
        payload.os = this.getFieldValue('osImageForm', 'osImage');
        payload.annotations = {
            'description': this.getFieldValue('metadataForm', 'clusterDescription'),
            'location': this.getFieldValue('metadataForm', 'clusterLocation')
        };

        let ldap_url = '';
        if (this.getFieldValue('identityForm', 'endpointIp')) {
            ldap_url = this.getFieldValue('identityForm', 'endpointIp') +
                ':' + this.getFieldValue('identityForm', 'endpointPort');
        }

        payload.identityManagement = {
            'idm_type': this.getFieldValue('identityForm', 'identityType') || 'none'
        }

        if (this.getFieldValue('identityForm', 'identityType') === 'oidc') {
            payload.identityManagement = Object.assign({
                    'oidc_provider_name': '',
                    'oidc_provider_url': this.getFieldValue('identityForm', 'issuerURL'),
                    'oidc_client_id': this.getFieldValue('identityForm', 'clientId'),
                    'oidc_client_secret': this.getFieldValue('identityForm', 'clientSecret'),
                    'oidc_scope': this.getFieldValue('identityForm', 'scopes'),
                    'oidc_claim_mappings': {
                        'username': this.getFieldValue('identityForm', 'oidcUsernameClaim'),
                        'groups': this.getFieldValue('identityForm', 'oidcGroupsClaim')
                    }

                }
                , payload.identityManagement);
        } else if (this.getFieldValue('identityForm', 'identityType') === 'ldap') {
            payload.identityManagement = Object.assign({
                    'ldap_url': ldap_url,
                    'ldap_bind_dn': this.getFieldValue('identityForm', 'bindDN'),
                    'ldap_bind_password': this.getFieldValue('identityForm', 'bindPW'),
                    'ldap_user_search_base_dn': this.getFieldValue('identityForm', 'userSearchBaseDN'),
                    'ldap_user_search_filter': this.getFieldValue('identityForm', 'userSearchFilter'),
                    'ldap_user_search_username': this.getFieldValue('identityForm', 'userSearchUsername'),
                    'ldap_user_search_name_attr': this.getFieldValue('identityForm', 'userSearchUsername'),
                    'ldap_group_search_base_dn': this.getFieldValue('identityForm', 'groupSearchBaseDN'),
                    'ldap_group_search_filter': this.getFieldValue('identityForm', 'groupSearchFilter'),
                    'ldap_group_search_user_attr': this.getFieldValue('identityForm', 'groupSearchUserAttr'),
                    'ldap_group_search_group_attr': this.getFieldValue('identityForm', 'groupSearchGroupAttr'),
                    'ldap_group_search_name_attr': this.getFieldValue('identityForm', 'groupSearchNameAttr'),
                    'ldap_root_ca': this.getFieldValue('identityForm', 'ldapRootCAData')
                }
                , payload.identityManagement);
        }

        payload.aviConfig = {
            'controller': this.getFieldValue('loadBalancerForm', 'controllerHost'),
            'username': this.getFieldValue('loadBalancerForm', 'username'),
            'password': this.getFieldValue('loadBalancerForm', 'password'),
            'cloud': this.getFieldValue('loadBalancerForm', 'cloudName'),
            'service_engine': this.getFieldValue('loadBalancerForm', 'serviceEngineGroupName'),
            'ca_cert': this.getFieldValue('loadBalancerForm', 'controllerCert'),
            'network': {
                'name': this.getFieldValue('loadBalancerForm', 'networkName'),
                'cidr': this.getFieldValue('loadBalancerForm', 'networkCIDR')
            },
            'labels': this.strMapToObj(this.getFieldValue('loadBalancerForm', 'clusterLabels'))
        }

        return payload;
    }
}
