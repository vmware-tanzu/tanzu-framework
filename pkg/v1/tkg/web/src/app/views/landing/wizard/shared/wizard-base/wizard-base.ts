// Angular imports
import { OnInit, ElementRef, AfterViewInit, ViewChild, Directive, Type } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';

// Third party imports
import { Observable } from 'rxjs';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { APP_ROUTES, Routes } from 'src/app/shared/constants/routes.constants';
import { Providers, PROVIDERS } from 'src/app/shared/constants/app.constants';
import { FormMetaDataStore } from '../FormMetaDataStore';
import { debounceTime, take, takeUntil } from 'rxjs/operators';
import { TkgEvent, TkgEventType } from './../../../../../shared/service/Messenger';
import { ClrStepper } from '@clr/angular';
import { FormMetaDataService } from 'src/app/shared/service/form-meta-data.service';
import { ConfigFileInfo } from '../../../../../swagger/models/config-file-info.model';
import Broker from 'src/app/shared/service/broker';
import { ClusterType, IdentityManagementType, WizardForm } from "../constants/wizard.constants";
import FileSaver from 'file-saver';
import { FormDataForHTML, FormUtility } from '../components/steps/form-utility';
import { SharedCeipStepComponent } from '../components/steps/ceip-step/ceip-step.component';
import { SharedIdentityStepComponent } from '../components/steps/identity-step/identity-step.component';
import { MetadataStepComponent } from '../components/steps/metadata-step/metadata-step.component';
import { SharedNetworkStepComponent } from '../components/steps/network-step/network-step.component';
import { SharedOsImageStepComponent } from '../components/steps/os-image-step/os-image-step.component';
import { StepFormDirective } from '../step-form/step-form';

// This interface describes a wizard that can register a step component
export interface WizardStepRegistrar {
    registerStep: (nameStep: string, stepComponent: StepFormDirective) => void,
    describeStep: (nameStep, staticDescription: string) => string,
    displayStep:  (nameStep: string) => boolean
}

@Directive()
export abstract class WizardBaseDirective extends BasicSubscriber implements WizardStepRegistrar, AfterViewInit, OnInit {
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

    title: string;
    edition: string;
    clusterTypeDescriptor: string;

    steps = [true, false, false, false, false, false, false, false, false, false, false];
    stepComponents: Map<string, StepFormDirective>;
    stepData: FormDataForHTML[];

    review = false;

    constructor(
        protected router: Router,
        protected el: ElementRef,
        protected formMetaDataService: FormMetaDataService,
        protected titleService: Title,
        protected formBuilder: FormBuilder
    ) {
        super();
    }

    ngOnInit() {
        this.form = this.formBuilder.group({});
        this.stepComponents = new Map<string, StepFormDirective>();
        // loop through stepData definitions and add a new form control for each step and we'll have the step formGroup objects built
        // even before the step components are instantiated (and Clarity will be happy, since it wants to process formGroup directives
        // before the step components are instantiated)
        for (const daStepData of this.stepData) {
            this.form.controls[daStepData.name] = this.formBuilder.group({});
        }

        // set branding and cluster type on branding change for base wizard components
        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                this.edition = data.payload.edition;
                this.clusterTypeDescriptor = data.payload.clusterTypeDescriptor;
                this.title = data.payload.branding.title;
            });

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
                if (this.isFormComplete(formName)) {
                    const stepForm = this.el.nativeElement.querySelector(`clr-stepper-panel[formgroupname=${formName}]`);
                    this.formMetaDataService.saveFormMetadata(formName, stepForm);
                }
            });
        });
    }

    // isFormComplete() is designed to indicate that a form's fields have been validated and are ready to be saved to local storage.
    // However, this method is misleading, because the wizard framework we're using does not take into account deactivated controls,
    // which may be necessary for full form completion.
    // For example, some controls are deactivated until the user enters credentials and we connect to a provider's server (the deactivation
    // prevents premature messages about required fields). After the connect, these controls are populated with values and activated; the
    // user is required to select a value, and once activated, the form validation will insist on a value before reporting VALID.
    // However, before the control is activated (and therefore part of the validation) the form may be validated and
    // the status will return 'VALID', even when the control has no value yet. The form is not yet fully completed,
    // and the values likely should not be stored in local storage, but right now we report the form is complete
    // and we do save the values we have.
    // Rather than change the isFormComplete() return value to reflect reality (by taking into account deactivated controls), we have
    // introduced a 'save-requires-value' attribute for these deactivated controls, so that at least their blank value will not be saved
    // to local storage (potentially overwriting a real value we want to keep, for use when the control is activated).
    private isFormComplete(formName: string): boolean {
        return this.form.controls[formName].status === 'VALID';
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
        const controlPlaneType = this.getControlPlaneFlavor(provider);
        if (controlPlaneType === 'dev') {
            return this.getFieldValue(`${provider}NodeSettingForm`, 'devInstanceType');
        } else if (controlPlaneType === 'prod') {
            return this.getFieldValue(`${provider}NodeSettingForm`, 'prodInstanceType');
        } else {
            return null;
        }
    }

    // Note: provider should be one of [aws,vsphere]; controlPlaneFlavor should be one of [dev, prod]
    saveControlPlaneNodeType(provider: string, controlPlaneFlavor: string, nodeType: string) {
        if (provider != null && controlPlaneFlavor != null) {
            this.saveFormField(`${provider}NodeSettingForm`, `${controlPlaneFlavor}InstanceType`, nodeType);
        }
    }

    getControlPlaneFlavor(provider: string) {
        return this.getFieldValue(`${provider}NodeSettingForm`, 'controlPlaneSetting');
    }

    saveControlPlaneFlavor(provider: string, flavor: string) {
        this.saveFormField(`${provider}NodeSettingForm`, 'controlPlaneSetting', flavor);
    }

    /**
     * Apply the settings captured via UI to backend TKG config without
     * actually creating the management/standalone cluster.
     */
    abstract applyTkgConfig(): Observable<ConfigFileInfo>;

    /**
     * Retrieve the config file from the backend and return as a string
     */
    abstract retrieveExportFile():  Observable<string>;

    /**
     * Switch the mode between "Review Configuration" and "Edit Configuration"
     * @param review In "Review Configuration" mode if true; otherwise in "Edit Configuration" mode
     */
    reviewConfiguration(review) {
        const pageTitle = (review) ? `${this.title} Confirm Settings` : this.title;
        this.titleService.setTitle(pageTitle);
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

    exportConfiguration() {
        this.retrieveExportFile().pipe(take(1)).subscribe(
            ((data) => {
                const blob = new Blob([data], {type: "text/plain;charset=utf-8"});
                FileSaver.saveAs(blob, 'config.yaml');
            }),
            ((err) => {
                this.displayError('Error encountered while creating export file: ' + err.toString());
            })
        )
    }

    displayError(errorMessage) {
        this.errorNotification = errorMessage;
    }

    getWizardValidity(): boolean {
        if (!FormMetaDataStore.getStepList()) {
            return false;
        }
        const totalSteps = FormMetaDataStore.getStepList().length;
        const stepsVisited = this.steps.filter(step => step).length;
        return stepsVisited > totalSteps && this.form.status === 'VALID';
    }

    getClusterType(): ClusterType {
        return Broker.appDataService.isModeClusterStandalone() ? ClusterType.Standalone : ClusterType.Management;
    }

    /**
     * @method method to trigger deployment
     */
    abstract createRegionalCluster(params: any): Observable<any>;
    abstract getPayload(): any;
    abstract setFromPayload(payload: any);

    isOnFirstStep() {
        // we're on the first step if we haven't reached the second step
        return !this.steps[1];
    }

    resetToFirstStep() {
        if (!this.isOnFirstStep()) {
            let activeStep;
            // Reset our steps array which tracks where we are
            this.steps[0] = true;
            // NOTE: we start at the second element
            for (let i = 1; i < this.steps.length; i++) {
                if (this.steps[i]) {
                    activeStep = i;
                    this.steps[i] = false;
                    break;
                }
            }
            this.wizard['stepperService'].resetPanels();
            this.wizard['stepperService']['accordion'].openFirstPanel();
        }
    }

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
     * Set the current value of the specified field
     * @param formName the form to set the field in
     * @param fieldName the name of the field to set
     * @param value the value to set the field to
     * Returns: true if successful; false if unable to get the form or the field
     */
    setFieldValue(formName, fieldName, value) {
        if (this.form.get(formName) && this.form.get(formName).get(fieldName)) {
            this.form.get(formName).get(fieldName).setValue(value);
            return true;
        }
        return false;
    }

    /**
     * Return the current value of the specified field
     * @param formName the form to get the field from
     * @param fieldName the name of the field to get
     */
    getFieldValue(formName, fieldName) {
        return this.form && this.form.get(formName) &&
            this.form.get(formName).get(fieldName) && this.form.get(formName).get(fieldName).value || '';
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
     * Converts iterable object to ES6 map
     * @param iterable object to be converted
     */
    iterableObjToStrMap(obj: any): Map<string, string> {
        const result = new Map<string, string>();
        if (obj !== null) {
            for (const [k, v] of obj) {
                result[k] = v;
            }
        }
        return result;
    }

    /**
     * Converts javascript object to ES6 map
     * @param javascript object to be converted
     */
    objToStrMap(obj: any): Map<string, string> {
        const result = new Map<string, string>();
        if (obj !== null) {
            Object.keys(obj).forEach(key => {
                result[key] = obj[key];
            })
        }
        return result;
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

        payload.ceipOptIn = this.getBooleanFieldValue('ceipOptInForm', 'ceipOptIn');
        payload.labels = this.strMapToObj(this.getFieldValue(WizardForm.METADATA, 'clusterLabels'));
        payload.os = this.getFieldValue(WizardForm.OSIMAGE, 'osImage');
        payload.annotations = {
            'description': this.getFieldValue(WizardForm.METADATA, 'clusterDescription'),
            'location': this.getFieldValue(WizardForm.METADATA, 'clusterLocation')
        };

        let ldap_url = '';
        if (this.getFieldValue(WizardForm.IDENTITY, 'endpointIp')) {
            ldap_url = this.getFieldValue(WizardForm.IDENTITY, 'endpointIp') +
                ':' + this.getFieldValue(WizardForm.IDENTITY, 'endpointPort');
        }

        payload.identityManagement = {
            'idm_type': this.getFieldValue(WizardForm.IDENTITY, 'identityType') || 'none'
        }

        if (this.getFieldValue(WizardForm.IDENTITY, 'identityType') === 'oidc') {
            payload.identityManagement = Object.assign({
                    'oidc_provider_name': '',
                    'oidc_provider_url': this.getFieldValue(WizardForm.IDENTITY, 'issuerURL'),
                    'oidc_client_id': this.getFieldValue(WizardForm.IDENTITY, 'clientId'),
                    'oidc_client_secret': this.getFieldValue(WizardForm.IDENTITY, 'clientSecret'),
                    'oidc_scope': this.getFieldValue(WizardForm.IDENTITY, 'scopes'),
                    'oidc_claim_mappings': {
                        'username': this.getFieldValue(WizardForm.IDENTITY, 'oidcUsernameClaim'),
                        'groups': this.getFieldValue(WizardForm.IDENTITY, 'oidcGroupsClaim')
                    }

                }
                , payload.identityManagement);
        } else if (this.getFieldValue(WizardForm.IDENTITY, 'identityType') === 'ldap') {
            payload.identityManagement = Object.assign({
                    'ldap_url': ldap_url,
                    'ldap_bind_dn': this.getFieldValue(WizardForm.IDENTITY, 'bindDN'),
                    'ldap_bind_password': this.getFieldValue(WizardForm.IDENTITY, 'bindPW'),
                    'ldap_user_search_base_dn': this.getFieldValue(WizardForm.IDENTITY, 'userSearchBaseDN'),
                    'ldap_user_search_filter': this.getFieldValue(WizardForm.IDENTITY, 'userSearchFilter'),
                    'ldap_user_search_username': this.getFieldValue(WizardForm.IDENTITY, 'userSearchUsername'),
                    'ldap_user_search_name_attr': this.getFieldValue(WizardForm.IDENTITY, 'userSearchUsername'),
                    'ldap_group_search_base_dn': this.getFieldValue(WizardForm.IDENTITY, 'groupSearchBaseDN'),
                    'ldap_group_search_filter': this.getFieldValue(WizardForm.IDENTITY, 'groupSearchFilter'),
                    'ldap_group_search_user_attr': this.getFieldValue(WizardForm.IDENTITY, 'groupSearchUserAttr'),
                    'ldap_group_search_group_attr': this.getFieldValue(WizardForm.IDENTITY, 'groupSearchGroupAttr'),
                    'ldap_group_search_name_attr': this.getFieldValue(WizardForm.IDENTITY, 'groupSearchNameAttr'),
                    'ldap_root_ca': this.getFieldValue(WizardForm.IDENTITY, 'ldapRootCAData')
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

    // Methods that fulfill WizardStepRegistrar
    //
    registerStep(stepName: string, stepComponent: StepFormDirective) {
        // create a formGroup for this step, record it internally and set it in the step
        stepComponent.setInputs(stepName, this.form.controls[stepName] as FormGroup);
        // record this step component internally
        this.stepComponents[stepName] = stepComponent;
    }

    // If the component has registered, returns the component's dynamic description (if it has one).
    // If the component has not yet registered, returns the static description (passed in)
    describeStep(stepName, staticDescription: string): string {
        const stepComponent = this.stepComponents[stepName];
        if (!stepComponent) {
            return staticDescription;
        }
        const dynamicDescription = stepComponent.dynamicDescription();
        return dynamicDescription ? dynamicDescription : staticDescription;
    }

    displayStep(stepName: string): boolean {
        const stepIndex = this.getStepIndex(stepName);
        return stepIndex >= 0 && this.steps[stepIndex];
    }
    //
    // Methods that fulfill WizardStepRegistrar

    // saveFormField() is a convenience method to avoid lengthy code lines
    saveFormField(formName, fieldName, value) {
        this.formMetaDataService.saveFormFieldData(formName, fieldName, value);
    }

    // saveFormListbox is a convenience method to avoid lengthy code lines
    saveFormListbox(formName, listboxName, key) {
        this.formMetaDataService.saveFormListboxData(formName, listboxName, key);
    }

    saveProxyFieldsFromPayload(payload: any) {
        if (payload.networking !== undefined && payload.networking.httpProxyConfiguration !== undefined) {
            const proxyConfig = payload.networking.httpProxyConfiguration;
            const hasProxySettings = proxyConfig.enabled;
            this.saveFormField('networkForm', 'proxySettings', hasProxySettings);
            if (hasProxySettings) {
                let proxySettingsMap = [
                    ['HTTPProxyURL', 'networkForm', 'httpProxyUrl'],
                    ['HTTPProxyUsername', 'networkForm', 'httpProxyUsername'],
                    ['HTTPProxyPassword', 'networkForm', 'httpProxyPassword'],
                    ['noProxy', 'networkForm', 'noProxy']
                ];
                // when HTTP matches HTTPS, we check the "matches" UI box and clear the HTTPS fields
                const httpMatchesHttps = this.httpMatchesHttpsSettings(proxyConfig);
                this.saveFormField('networkForm', 'isSameAsHttp', httpMatchesHttps);
                if (httpMatchesHttps) {
                    this.saveFormField('networkForm', 'httpsProxyUrl', '');
                    this.saveFormField('networkForm', 'httpsProxyUsername', '');
                    this.saveFormField('networkForm', 'httpsProxyPassword', '');
                } else {
                    proxySettingsMap = [
                        ...proxySettingsMap,
                        ['HTTPSProxyURL', 'networkForm', 'httpsProxyUrl'],
                        ['HTTPSProxyUsername', 'networkForm', 'httpsProxyUsername'],
                        ['HTTPSProxyPassword', 'networkForm', 'httpsProxyPassword']
                    ];
                }
                proxySettingsMap.forEach(attr => {
                    this.saveFormField(attr[1], attr[2], proxyConfig[attr[0]]);
                });
            }
        }
    }

    /**
     * Fill in payload with values from all common steps
     * @param payload
     */
    saveCommonFieldsFromPayload(payload: any) {
        if (payload.networking !== undefined ) {
            // Networking - general
            this.saveFormField('networkForm', 'networkName', payload.networking.networkName);
            this.saveFormField('networkForm', 'clusterServiceCidr', payload.networking.clusterServiceCIDR);
            this.saveFormField('networkForm', 'clusterPodCidr', payload.networking.clusterPodCIDR);
            this.saveFormField('networkForm', 'cniType', payload.networking.cniType);
        }

        // Proxy settings
        this.saveProxyFieldsFromPayload(payload);

        // Other fields
        this.saveFormField('ceipOptInForm', 'ceipOptIn', payload.ceipOptIn);
        this.saveFormField('registerTmcForm', 'tmcRegUrl', payload.tmc_registration_url);
        if (payload.labels !== undefined) {
            // we construct a label value that mimics how the meta-data step constructs the saved label value
            // when the user creates it label by label
            const labelArray: Array<string> = [];
            Object.keys(payload.labels).forEach(key => {
                const value = payload.labels[key];
                labelArray[labelArray.length] = key + ":" + value;
            });
            const labelValueToSave = labelArray.join(', ');
            this.saveFormField('metadataForm', 'clusterLabels', labelValueToSave);
        }
        this.saveFormField('osImageForm', 'osImage', payload.os);
        if (payload.annotations !== undefined) {
            this.saveFormField('metadataForm', 'clusterDescription', payload.annotations.description);
            this.saveFormField('metadataForm', 'clusterLocation', payload.annotations.location);
        }

        // Identity Management form
        if (payload.identityManagement !== undefined) {
            const idmType = payload.identityManagement.idm_type === 'none' ? '' : payload.identityManagement.idm_type;
            this.saveFormField('identityForm', 'identityType', idmType);
            if (idmType === 'oidc') {
                this.saveFormField('identityForm', 'issuerURL', payload.identityManagement.oidc_provider_url);
                this.saveFormField('identityForm', 'clientId', payload.identityManagement.oidc_client_id);
                this.saveFormField('identityForm', 'clientSecret', payload.identityManagement.oidc_client_secret);
                this.saveFormField('identityForm', 'scopes', payload.identityManagement.oidc_scope);
                this.saveFormField('identityForm', 'oidcUsernameClaim', payload.identityManagement.oidc_claim_mappings.username);
                this.saveFormField('identityForm', 'oidcGroupsClaim', payload.identityManagement.oidc_claim_mappings.groups);
            } else if (idmType === 'ldap') {
                if (payload.id.ldap_url !== undefined) {
                    // separate the IP address from the port in the LDAP URL
                    const ldapUrlPieces = payload.id.ldap_url.split(':');
                    this.saveFormField('identityForm', 'endpointIp', ldapUrlPieces[0]);
                    if (ldapUrlPieces.length() > 1) {
                        this.saveFormField('identityForm', 'endpointPort', ldapUrlPieces[1]);
                    }
                }
                this.saveFormField('identityForm', 'bindDN', payload.identityManagement.ldap_bind_dn);
                this.saveFormField('identityForm', 'bindPW', payload.identityManagement.ldap_bind_password);
                this.saveFormField('identityForm', 'userSearchBaseDN', payload.identityManagement.ldap_user_search_base_dn);
                this.saveFormField('identityForm', 'userSearchFilter', payload.identityManagement.ldap_user_search_filter);
                this.saveFormField('identityForm', 'userSearchUsername', payload.identityManagement.ldap_user_search_username);
                this.saveFormField('identityForm', 'userSearchUsername', payload.identityManagement.ldap_user_search_name_attr);
                this.saveFormField('identityForm', 'groupSearchBaseDN', payload.identityManagement.ldap_group_search_base_dn);
                this.saveFormField('identityForm', 'groupSearchFilter', payload.identityManagement.ldap_group_search_filter);
                this.saveFormField('identityForm', 'groupSearchUserAttr', payload.identityManagement.ldap_group_search_user_attr);
                this.saveFormField('identityForm', 'groupSearchGroupAttr', payload.identityManagement.ldap_group_search_group_attr);
                this.saveFormField('identityForm', 'groupSearchNameAttr', payload.identityManagement.ldap_group_search_name_attr);
                this.saveFormField('identityForm', 'ldapRootCAData', payload.identityManagement.ldap_root_ca);
            }
        }

        if (payload.aviConfig !== undefined) {
            // Load Balancer form
            this.saveFormField('loadBalancerForm', 'controllerHost', payload.aviConfig.controller);
            this.saveFormField('loadBalancerForm', 'username', payload.aviConfig.username);
            this.saveFormField('loadBalancerForm', 'password', payload.aviConfig.password);
            this.saveFormField('loadBalancerForm', 'cloudName', payload.aviConfig.cloud);
            this.saveFormField('loadBalancerForm', 'serviceEngineGroupName', payload.aviConfig.service_engine);
            this.saveFormField('loadBalancerForm', 'controllerCert', payload.aviConfig.ca_cert);
            this.saveFormField('loadBalancerForm', 'networkName', payload.aviConfig.network.name);
            this.saveFormField('loadBalancerForm', 'networkCIDR', payload.aviConfig.network.cidr);
            this.saveFormField('loadBalancerForm', 'clusterLabels', this.objToStrMap(payload.aviConfig.labels));
        }
    }

    private httpMatchesHttpsSettings(httpProxyConfiguration: any) {
        return httpProxyConfiguration['HTTPProxyURL'] === httpProxyConfiguration['HTTPSProxyURL'] &&
            httpProxyConfiguration['HTTPProxyUsername'] === httpProxyConfiguration['HTTPSProxyUsername'] &&
            httpProxyConfiguration['HTTPProxyPassword'] === httpProxyConfiguration['HTTPSProxyPassword'];
    }

    // HTML convenience methods
    //
    get registrar(): WizardStepRegistrar {
        return this;
    }

    get CeipForm(): FormDataForHTML {
        return { name: WizardForm.CEIP, title: 'CEIP Agreement', description: 'Join the CEIP program for TKG',
            i18n: { title: 'ceip agreement step title', description: 'ceip agreement step description' },
        clazz: SharedCeipStepComponent };
    }
    get IdentityForm(): FormDataForHTML {
        return { name: WizardForm.IDENTITY, title: 'Identity Management', description: 'Specify identity management',
            i18n: { title: 'identity step title', description: 'identity step description' },
        clazz: SharedIdentityStepComponent };
    }
    get MetadataForm(): FormDataForHTML {
        return { name: WizardForm.METADATA, title: 'Metadata',
            description: 'Location: Specify metadata for the ' + this.clusterTypeDescriptor + ' cluster',
            i18n: { title: 'metadata step name', description: 'metadata step description' },
        clazz: MetadataStepComponent };
    }
    get NetworkForm(): FormDataForHTML {
        return { name: WizardForm.NETWORK, title: 'Kubernetes Network',
            description: 'Specify how TKG networking is provided and global network settings',
            i18n: { title: 'Kubernetes network step name', description: 'Kubernetes network step description' },
        clazz: SharedNetworkStepComponent };
    }
    getOsImageForm(clazz: Type<StepFormDirective>): FormDataForHTML {
        return { name: WizardForm.OSIMAGE, title: 'OS Image', description: 'Specify the OS Image',
            i18n: { title: 'OS Image step title', description: 'OS Image step description' },
        clazz: clazz };
    }
    get wizardForm(): FormGroup {
        return this.form;
    }
    //
    // HTML convenience methods

    protected getStepIndex(stepName: string): number {
        let result = -1;
        this.stepData.forEach((data, index) => {
            if (data.name === stepName) {
                result = index;
            }
        });
        return result;
    }
}
