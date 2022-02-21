// Angular imports
import { Directive, ElementRef, OnInit, Type, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';
// Third party imports
import { takeUntil } from 'rxjs/operators';
import { Observable } from 'rxjs';
// App imports
import { APP_ROUTES, Routes } from 'src/app/shared/constants/routes.constants';
import AppServices from '../../../../../shared/service/appServices';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { CeipField } from '../components/steps/ceip-step/ceip-step.fieldmapping';
import { ClusterType, IdentityManagementType, WizardForm } from "../constants/wizard.constants";
import { ConfigFileInfo } from '../../../../../swagger/models/config-file-info.model';
import { EditionData } from '../../../../../shared/service/branding.service';
import { FieldMapping } from '../field-mapping/FieldMapping';
import { FormDataForHTML, FormUtility } from '../components/steps/form-utility';
import { IdentityField } from '../components/steps/identity-step/identity-step.fieldmapping';
import { LoadBalancerField } from '../components/steps/load-balancer/load-balancer-step.fieldmapping';
import { MetadataField } from '../components/steps/metadata-step/metadata-step.fieldmapping';
import { MetadataStepComponent } from '../components/steps/metadata-step/metadata-step.component';
import { NetworkField } from '../components/steps/network-step/network-step.fieldmapping';
import { OsImageField } from '../components/steps/os-image-step/os-image-step.fieldmapping';
import { Providers, PROVIDERS } from 'src/app/shared/constants/app.constants';
import { SharedCeipStepComponent } from '../components/steps/ceip-step/ceip-step.component';
import { SharedIdentityStepComponent } from '../components/steps/identity-step/identity-step.component';
import { SharedNetworkStepComponent } from '../components/steps/network-step/network-step.component';
import { StepFormDirective } from '../step-form/step-form';
import {
    StepCompletedPayload,
    StepDescriptionChangePayload,
    StepStartedPayload,
    TanzuEventType
} from './../../../../../shared/service/Messenger';
import { StepWrapperSetComponent } from '../step-wrapper/step-wrapper-set.component';
import { NodeSettingField } from '../components/steps/node-setting-step/node-setting-step.fieldmapping';

// This interface describes a wizard that can register a step component
export interface WizardStepRegistrar {
    registerStep: (nameStep: string, stepComponent: StepFormDirective) => void,
    stepDescription: Map<string, string>,
}

// This interface contains the data that the wizard sends to a registering step
export interface StepRegistrantData {
    wizard: string,                     // the name of the wizard that is using this step
    step: string,                       // the name the wizard has for this step
    formGroup: FormGroup,               // the FormGroup the wizard has created for the step to use for all its field controls
    eventFileImported: TanzuEventType,    // the event the wizard broadcasts when the user has successfully imported a config file
    eventFileImportError: TanzuEventType, // the event the wizard broadcasts when an error occurs during file import
}

@Directive()
export abstract class WizardBaseDirective extends BasicSubscriber implements WizardStepRegistrar, OnInit {
    APP_ROUTES: Routes = APP_ROUTES;
    PROVIDERS: Providers = PROVIDERS;

    @ViewChild(StepWrapperSetComponent)
    stepWrapperSet: StepWrapperSetComponent;

    form: FormGroup;
    errorNotification: string;
    provider: Observable<string>;
    providerType: string;
    deploymentPending: boolean = false;
    disableDeployButton = false;

    title: string;
    edition: string;
    clusterTypeDescriptor: string = '';

    stepDescription: Map<string, string> = new Map<string, string>();   // Field that fulfill WizardStepRegistrar
    stepData: FormDataForHTML[];    // needs to be public for step-wrapper-set to use
    private currentStep: string;
    private visitedLastStep: boolean;

    review = false;

    protected constructor(
        protected router: Router,
        protected el: ElementRef,
        protected titleService: Title,
        protected formBuilder: FormBuilder
    ) {
        super();
    }

    // supplyFileImportedEvent() allows the child class to give this class the event to broadcast on successful file import
    protected abstract supplyFileImportedEvent(): TanzuEventType;
    // supplyFileImportErrorEvent() allows the child class to give this class the event to broadcast on file import error
    protected abstract supplyFileImportErrorEvent(): TanzuEventType;
    // supplyStepData() allows the child class gives this class the data for the steps.
    protected abstract supplyStepData(): FormDataForHTML[];
    // supplyWizardName() allows the child class gives this class the wizard name; this is used to identify which wizard a step belongs to
    protected abstract supplyWizardName(): string;
    // supplyDisplayOrder() allows the child class to specify the order (and which steps) get displayed (on confirmation page).
    // By default, we take the order from the stepData (so stepData should be set before invoking this method)
    protected supplyDisplayOrder(): string[] {
        if (!this.stepData || this.stepData.length === 0) {
            console.warn('supplyDisplayOrder() called before step data was set');
            return [];
        }
        return this.defaultDisplayOrder(this.stepData);
    }

    ngOnInit() {
        this.form = this.formBuilder.group({});
        // loop through stepData definitions and add a new form control for each step and we'll have the step formGroup objects built
        // even before the step components are instantiated (and Clarity will be happy, since it wants to process formGroup directives
        // before the step components are instantiated)
        this.stepData = this.supplyStepData();
        if (!this.stepData || this.stepData.length === 0) {
            console.error('wizard did not supply step data to base class');
        } else {
            this.storeWizardDisplayOrder(this.supplyDisplayOrder());
            this.storeWizardTitles();   // since the titles don't change, store them once
            for (const daStepData of this.stepData) {
                this.form.controls[daStepData.name] = this.formBuilder.group({});
                this.stepDescription[daStepData.name] = daStepData.description;
            }
            this.currentStep = this.firstStep;
        }

        // set step description (if it's a step description for this wizard)
        AppServices.messenger.subscribe<StepDescriptionChangePayload>(TanzuEventType.STEP_DESCRIPTION_CHANGE, data => {
                const stepDescriptionPayload = data.payload as StepDescriptionChangePayload;
                if (this.supplyWizardName() === stepDescriptionPayload.wizard) {
                    // we use setTimeout to avoid a possible ExpressionChangedAfterItHasBeenCheckedError
                    setTimeout(() => { this.stepDescription[stepDescriptionPayload.step] = stepDescriptionPayload.description; }, 0);
                }
            }, this.unsubscribe);

        // set branding and cluster type on branding change for base wizard components
        AppServices.messenger.subscribe<EditionData>(TanzuEventType.BRANDING_CHANGED, data => {
                this.edition = data.payload.edition;
                this.clusterTypeDescriptor = data.payload.clusterTypeDescriptor;
                this.title = data.payload.branding.title;
            }, this.unsubscribe);

        setTimeout(() => this.broadcastStepStarted(this.firstStep), 0);
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

    displayError(errorMessage) {
        this.errorNotification = errorMessage;
    }

    getWizardValidity(): boolean {
        return this.visitedLastStep && this.form.status === 'VALID';
    }

    getClusterType(): ClusterType {
        return AppServices.appDataService.isModeClusterStandalone() ? ClusterType.Standalone : ClusterType.Management;
    }

    /**
     * @method method to trigger deployment
     */
    abstract createRegionalCluster(params: any): Observable<any>;
    abstract getPayload(): any;
    abstract setFromPayload(payload: any);

    isOnFirstStep() {
        return this.currentStep === this.firstStep;
    }

    resetToFirstStep() {
        if (!this.isOnFirstStep()) {
            this.currentStep = this.firstStep;
            this.visitedLastStep = false;
            this.stepWrapperSet.restartWizard();
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
    onNextStep(stepCompleted: string) {
        this.broadcastStepComplete(stepCompleted);
        if (stepCompleted === this.lastStep) {
            this.visitedLastStep = true;
        } else {
            const indexCompletedStep = this.stepData.findIndex(stepData => stepData.name === stepCompleted);
            this.setCurrentStep(this.stepData[indexCompletedStep + 1].name);
            this.broadcastStepStarted(this.currentStep);
        }
        // TODO: we need to know that the last step has actually completed its recording of the data
        // NOTE: we do onWizardComplete EVERY time the user completes ANY step, if they have previously completed the wizard
        if (this.visitedLastStep) {
            this.onWizardComplete();
        }
    }

    private setCurrentStep(step: string) {
        this.currentStep = step;
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
        AppServices.messenger.publish({
            type: TanzuEventType.CLI_CHANGED,
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
        if (obj) {
            Object.keys(obj).forEach(key => {
                result.set(key, obj[key]);
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
            networkName: this.getFieldValue(WizardForm.NETWORK, NetworkField.NETWORK_NAME),
            clusterDNSName: '',
            clusterNodeCIDR: '',
            clusterServiceCIDR: this.getFieldValue(WizardForm.NETWORK, NetworkField.CLUSTER_SERVICE_CIDR),
            clusterPodCIDR: this.getFieldValue(WizardForm.NETWORK, NetworkField.CLUSTER_POD_CIDR),
            cniType: this.getFieldValue(WizardForm.NETWORK, NetworkField.CNI_TYPE)
        };

        if (this.getFieldValue(WizardForm.NETWORK, NetworkField.PROXY_SETTINGS)) {
            let proxySettingsMap = null;
            proxySettingsMap = [
                ['HTTPProxyURL', WizardForm.NETWORK, NetworkField.HTTP_PROXY_URL],
                ['HTTPProxyUsername', WizardForm.NETWORK, NetworkField.HTTP_PROXY_USERNAME],
                ['HTTPProxyPassword', WizardForm.NETWORK, NetworkField.HTTP_PROXY_PASSWORD],
                ['noProxy', WizardForm.NETWORK, NetworkField.NO_PROXY]
            ];
            if (this.getFieldValue(WizardForm.NETWORK, NetworkField.HTTPS_IS_SAME_AS_HTTP)) {
                proxySettingsMap = [
                    ...proxySettingsMap,
                    ['HTTPSProxyURL', WizardForm.NETWORK, NetworkField.HTTP_PROXY_URL],
                    ['HTTPSProxyUsername', WizardForm.NETWORK, NetworkField.HTTP_PROXY_USERNAME],
                    ['HTTPSProxyPassword', WizardForm.NETWORK, NetworkField.HTTP_PROXY_PASSWORD]
                ];
            } else {
                proxySettingsMap = [
                    ...proxySettingsMap,
                    ['HTTPSProxyURL', WizardForm.NETWORK, NetworkField.HTTPS_PROXY_URL],
                    ['HTTPSProxyUsername', WizardForm.NETWORK, NetworkField.HTTPS_PROXY_USERNAME],
                    ['HTTPSProxyPassword', WizardForm.NETWORK, NetworkField.HTTPS_PROXY_PASSWORD]
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

        payload.ceipOptIn = this.getBooleanFieldValue(WizardForm.CEIP, CeipField.OPTIN);
        // TODO: for labels, we are reaching into storage to get the value, whereas all the other data come from fields
        // It would be better for ALL the fields to use a FieldMapping to retrieve the data
        const labelFieldMapping: FieldMapping = {name: MetadataField.CLUSTER_LABELS, isMap: true};
        let labelsMap = AppServices.userDataService.retrieveStoredValue(this.supplyWizardName(), WizardForm.METADATA, labelFieldMapping);
        payload.labels = this.strMapToObj(labelsMap);

        payload.os = this.getFieldValue(WizardForm.OSIMAGE, OsImageField.IMAGE);
        payload.annotations = {
            'description': this.getFieldValue(WizardForm.METADATA, MetadataField.CLUSTER_DESCRIPTION),
            'location': this.getFieldValue(WizardForm.METADATA, MetadataField.CLUSTER_LOCATION)
        };

        let ldap_url = '';
        if (this.getFieldValue(WizardForm.IDENTITY, IdentityField.ENDPOINT_IP)) {
            ldap_url = this.getFieldValue(WizardForm.IDENTITY, IdentityField.ENDPOINT_IP) +
                ':' + this.getFieldValue(WizardForm.IDENTITY, IdentityField.ENDPOINT_PORT);
        }

        payload.identityManagement = {
            'idm_type': this.getFieldValue(WizardForm.IDENTITY, IdentityField.IDENTITY_TYPE) || 'none'
        }

        if (this.getFieldValue(WizardForm.IDENTITY, IdentityField.IDENTITY_TYPE) === IdentityManagementType.OIDC) {
            payload.identityManagement = Object.assign({
                    'oidc_provider_name': '',
                    'oidc_provider_url': this.getFieldValue(WizardForm.IDENTITY, IdentityField.ISSUER_URL),
                    'oidc_client_id': this.getFieldValue(WizardForm.IDENTITY, IdentityField.CLIENT_ID),
                    'oidc_client_secret': this.getFieldValue(WizardForm.IDENTITY, IdentityField.CLIENT_SECRET),
                    'oidc_scope': this.getFieldValue(WizardForm.IDENTITY, IdentityField.SCOPES),
                    'oidc_claim_mappings': {
                        'username': this.getFieldValue(WizardForm.IDENTITY, IdentityField.OIDC_USERNAME_CLAIM),
                        'groups': this.getFieldValue(WizardForm.IDENTITY, IdentityField.OIDC_GROUPS_CLAIM)
                    }

                }
                , payload.identityManagement);
        } else if (this.getFieldValue(WizardForm.IDENTITY, IdentityField.IDENTITY_TYPE) === IdentityManagementType.LDAP) {
            payload.identityManagement = Object.assign({
                    'ldap_url': ldap_url,
                    'ldap_bind_dn': this.getFieldValue(WizardForm.IDENTITY, IdentityField.BIND_DN),
                    'ldap_bind_password': this.getFieldValue(WizardForm.IDENTITY, IdentityField.BIND_PW),
                    'ldap_user_search_base_dn': this.getFieldValue(WizardForm.IDENTITY, IdentityField.USER_SEARCH_BASE_DN),
                    'ldap_user_search_filter': this.getFieldValue(WizardForm.IDENTITY, IdentityField.USER_SEARCH_FILTER),
                    'ldap_user_search_username': this.getFieldValue(WizardForm.IDENTITY, IdentityField.USER_SEARCH_USERNAME),
                    'ldap_user_search_name_attr': this.getFieldValue(WizardForm.IDENTITY, IdentityField.USER_SEARCH_USERNAME),
                    'ldap_group_search_base_dn': this.getFieldValue(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_BASE_DN),
                    'ldap_group_search_filter': this.getFieldValue(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_FILTER),
                    'ldap_group_search_user_attr': this.getFieldValue(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_USER_ATTR),
                    'ldap_group_search_group_attr': this.getFieldValue(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_GROUP_ATTR),
                    'ldap_group_search_name_attr': this.getFieldValue(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_NAME_ATTR),
                    'ldap_root_ca': this.getFieldValue(WizardForm.IDENTITY, IdentityField.LDAP_ROOT_CA)
                }
                , payload.identityManagement);
        }

        // TODO: for clusterLabels, we are reaching into storage to get the value, whereas all the other data come from fields
        // It would be better for ALL the fields to use a FieldMapping to retrieve the data
        labelsMap = AppServices.userDataService.retrieveStoredValue(this.supplyWizardName(), 'loadBalancerForm', labelFieldMapping);
        payload.aviConfig = {
            'controller': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.CONTROLLER_HOST),
            'username': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.USERNAME),
            'password': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.PASSWORD),
            'cloud': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.CLOUD_NAME),
            'service_engine': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.SERVICE_ENGINE_GROUP_NAME),
            'ca_cert': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.CONTROLLER_CERT),
            'network': {
                'name': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.NETWORK_NAME),
                'cidr': this.getFieldValue(WizardForm.LOADBALANCER, LoadBalancerField.NETWORK_CIDR)
            },
            'labels': this.strMapToObj(labelsMap),
        }
        return payload;
    }

    // Methods that fulfill WizardStepRegistrar
    //
    registerStep(stepName: string, stepComponent: StepFormDirective) {
        // set the wizard name, stepName and formGroup (already created for this step) into the component
        const stepRegistrantData = {
            wizard: this.supplyWizardName(),
            step: stepName,
            formGroup: this.form.controls[stepName] as FormGroup,
            eventFileImported: this.supplyFileImportedEvent(),
            eventFileImportError: this.supplyFileImportErrorEvent(),
        }
        stepComponent.setStepRegistrantData(stepRegistrantData);
    }
    //
    // Methods that fulfill WizardStepRegistrar

    // storeFieldString() is a convenience method to avoid lengthy code lines
    storeFieldString(step, field, value: string, displayString?: string) {
        const identifier = { wizard: this.supplyWizardName(), step, field };
        const display = displayString ? displayString : value;
        AppServices.userDataService.store(identifier, { value, display });
    }

    storeFieldBoolean(step, field: string, booleanValue: boolean) {
        const identifier = { wizard: this.supplyWizardName(), step, field };
        const display = booleanValue ? 'yes' : 'no';
        const value = String(booleanValue);
        AppServices.userDataService.store(identifier, { value, display });
    }

    storeProxyFieldsFromPayload(payload: any) {
        if (payload.networking !== undefined && payload.networking.httpProxyConfiguration !== undefined) {
            const proxyConfig = payload.networking.httpProxyConfiguration;
            const hasProxySettings = proxyConfig.enabled;
            this.storeFieldString(WizardForm.NETWORK, NetworkField.PROXY_SETTINGS, hasProxySettings);
            if (hasProxySettings) {
                let proxySettingsMap = [
                    ['HTTPProxyURL', WizardForm.NETWORK, NetworkField.HTTP_PROXY_URL],
                    ['HTTPProxyUsername', WizardForm.NETWORK, NetworkField.HTTP_PROXY_USERNAME],
                    ['HTTPProxyPassword', WizardForm.NETWORK, NetworkField.HTTP_PROXY_PASSWORD],
                    ['noProxy', WizardForm.NETWORK, NetworkField.NO_PROXY]
                ];
                // when HTTP matches HTTPS, we check the "matches" UI box and clear the HTTPS fields
                const httpMatchesHttps = this.httpMatchesHttpsSettings(proxyConfig);
                this.storeFieldBoolean(WizardForm.NETWORK, NetworkField.HTTPS_IS_SAME_AS_HTTP, httpMatchesHttps);
                if (httpMatchesHttps) {
                    this.storeFieldString(WizardForm.NETWORK, NetworkField.HTTPS_PROXY_URL, '');
                    this.storeFieldString(WizardForm.NETWORK, NetworkField.HTTPS_PROXY_USERNAME, '');
                    this.storeFieldString(WizardForm.NETWORK, NetworkField.HTTPS_PROXY_PASSWORD, '');
                } else {
                    proxySettingsMap = [
                        ...proxySettingsMap,
                        ['HTTPSProxyURL', WizardForm.NETWORK, NetworkField.HTTPS_PROXY_URL],
                        ['HTTPSProxyUsername', WizardForm.NETWORK, NetworkField.HTTPS_PROXY_USERNAME],
                        ['HTTPSProxyPassword', WizardForm.NETWORK, NetworkField.HTTPS_PROXY_PASSWORD]
                    ];
                }
                proxySettingsMap.forEach(attr => {
                    this.storeFieldString(attr[1], attr[2], proxyConfig[attr[0]]);
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
            this.storeFieldString(WizardForm.NETWORK, NetworkField.NETWORK_NAME, payload.networking.networkName);
            this.storeFieldString(WizardForm.NETWORK, NetworkField.CLUSTER_SERVICE_CIDR, payload.networking.clusterServiceCIDR);
            this.storeFieldString(WizardForm.NETWORK, NetworkField.CLUSTER_POD_CIDR, payload.networking.clusterPodCIDR);
            this.storeFieldString(WizardForm.NETWORK, NetworkField.CNI_TYPE, payload.networking.cniType);
        }

        // Proxy settings
        this.storeProxyFieldsFromPayload(payload);

        // Other fields
        this.storeFieldString(WizardForm.CEIP, CeipField.OPTIN, payload.ceipOptIn);
        if (payload.labels !== undefined) {
            // we construct a label value that mimics how the meta-data step constructs the saved label value
            // when the user creates it label by label
            const labelArray: Array<string> = [];
            Object.keys(payload.labels).forEach(key => {
                const value = payload.labels[key];
                labelArray[labelArray.length] = key + ":" + value;
            });
            const labelValueToSave = labelArray.join(', ');
            this.storeFieldString(WizardForm.METADATA, MetadataField.CLUSTER_LABELS, labelValueToSave);
        }

        if (payload.annotations !== undefined) {
            this.storeFieldString(WizardForm.METADATA, MetadataField.CLUSTER_DESCRIPTION, payload.annotations.description);
            this.storeFieldString(WizardForm.METADATA, MetadataField.CLUSTER_LOCATION, payload.annotations.location);
        }

        // Identity Management form
        if (payload.identityManagement !== undefined) {
            const idmType = payload.identityManagement.idm_type === 'none' ? '' : payload.identityManagement.idm_type;
            this.storeFieldString(WizardForm.IDENTITY, IdentityField.IDENTITY_TYPE, idmType);
            if (idmType === IdentityManagementType.OIDC) {
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.ISSUER_URL, payload.identityManagement.oidc_provider_url);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.CLIENT_ID, payload.identityManagement.oidc_client_id);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.CLIENT_SECRET, payload.identityManagement.oidc_client_secret);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.SCOPES, payload.identityManagement.oidc_scope);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.OIDC_USERNAME_CLAIM,
                    payload.identityManagement.oidc_claim_mappings.username);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.OIDC_GROUPS_CLAIM,
                    payload.identityManagement.oidc_claim_mappings.groups);
            } else if (idmType === IdentityManagementType.LDAP) {
                if (payload.id.ldap_url !== undefined) {
                    // separate the IP address from the port in the LDAP URL
                    const ldapUrlPieces = payload.id.ldap_url.split(':');
                    this.storeFieldString(WizardForm.IDENTITY, IdentityField.ENDPOINT_IP, ldapUrlPieces[0]);
                    if (ldapUrlPieces.length() > 1) {
                        this.storeFieldString(WizardForm.IDENTITY, IdentityField.ENDPOINT_PORT, ldapUrlPieces[1]);
                    }
                }
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.BIND_DN, payload.identityManagement.ldap_bind_dn);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.BIND_PW, payload.identityManagement.ldap_bind_password);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.USER_SEARCH_BASE_DN,
                    payload.identityManagement.ldap_user_search_base_dn);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.USER_SEARCH_FILTER,
                    payload.identityManagement.ldap_user_search_filter);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.USER_SEARCH_USERNAME,
                    payload.identityManagement.ldap_user_search_username);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.USER_SEARCH_USERNAME,
                    payload.identityManagement.ldap_user_search_name_attr);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_BASE_DN,
                    payload.identityManagement.ldap_group_search_base_dn);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_FILTER,
                    payload.identityManagement.ldap_group_search_filter);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_USER_ATTR,
                    payload.identityManagement.ldap_group_search_user_attr);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_GROUP_ATTR,
                    payload.identityManagement.ldap_group_search_group_attr);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.GROUP_SEARCH_NAME_ATTR,
                    payload.identityManagement.ldap_group_search_name_attr);
                this.storeFieldString(WizardForm.IDENTITY, IdentityField.LDAP_ROOT_CA, payload.identityManagement.ldap_root_ca);
            }
        }

        if (payload.aviConfig !== undefined) {
            // Load Balancer form
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.CONTROLLER_HOST, payload.aviConfig.controller);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.USERNAME, payload.aviConfig.username);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.PASSWORD, payload.aviConfig.password);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.CLOUD_NAME, payload.aviConfig.cloud);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.SERVICE_ENGINE_GROUP_NAME, payload.aviConfig.service_engine);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.CONTROLLER_CERT, payload.aviConfig.ca_cert);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.NETWORK_NAME, payload.aviConfig.network.name);
            this.storeFieldString(WizardForm.LOADBALANCER, LoadBalancerField.NETWORK_CIDR, payload.aviConfig.network.cidr);
            this.storeMap(WizardForm.LOADBALANCER, LoadBalancerField.CLUSTER_LABELS, this.objToStrMap(payload.aviConfig.labels));
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
        return { name: WizardForm.IDENTITY, title: 'Identity Management', description: SharedIdentityStepComponent.description,
            i18n: { title: 'identity step title', description: 'identity step description' },
        clazz: SharedIdentityStepComponent };
    }
    get MetadataForm(): FormDataForHTML {
        return { name: WizardForm.METADATA, title: 'Metadata',
            description: 'Specify metadata for the ' + this.clusterTypeDescriptor + ' cluster',
            i18n: { title: 'metadata step name', description: 'metadata step description' },
        clazz: MetadataStepComponent };
    }
    get NetworkForm(): FormDataForHTML {
        return { name: WizardForm.NETWORK, title: 'Kubernetes Network',
            description: SharedNetworkStepComponent.description,
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
    get clusterTypeDescriptorTitleCase() {
        return FormUtility.titleCase(this.clusterTypeDescriptor);
    }
    get wizardName(): string {
        return this.supplyWizardName();
    }

    get isDataOld(): boolean {
        return AppServices.userDataService.isWizardDataOld(this.supplyWizardName());
    }
    //
    // HTML convenience methods

    // convenience methods to keep code clean
    get lastStep() {
        return this.numSteps > 0 ? this.stepData[this.numSteps - 1].name : '';
    }

    get firstStep() {
        return this.numSteps > 0 ? this.stepData[0].name : '';
    }

    get numSteps() {
        return this.stepData ? this.stepData.length : 0;
    }

    private broadcastStepComplete(stepCompletedName: string) {
        const payload: StepCompletedPayload = {
            wizard: this.supplyWizardName(),
            step: stepCompletedName,
        }
        AppServices.messenger.publish<StepCompletedPayload>( { type: TanzuEventType.STEP_COMPLETED, payload } );
    }

    private broadcastStepStarted(stepStartedName: string) {
        const payload: StepStartedPayload = {
            wizard: this.supplyWizardName(),
            step: stepStartedName,
        }
        AppServices.messenger.publish<StepStartedPayload>( { type: TanzuEventType.STEP_STARTED, payload } );
    }

    private defaultDisplayOrder(stepData: FormDataForHTML[]): string[] {
        // reduce the array of stepData items into an array of step name strings, which will be in the same order
        return stepData.reduce<string[]>((accumulator, daStep) => {
            accumulator.push(daStep.name); return accumulator;
        }, []);
    }
    private storeWizardDisplayOrder(displayOrder: string[]) {
        AppServices.userDataService.storeWizardDisplayOrder(this.supplyWizardName(), displayOrder);
    }
    private storeWizardStepDescriptions() {
        AppServices.userDataService.storeWizardDescriptions(this.wizardName, this.stepDescription);
    }
    private storeWizardTitles() {
        const titles = this.stepData.reduce<Map<string, string>>( (accumulator, stepData) => {
            accumulator[stepData.name] = stepData.title;
            return accumulator;
        }, new Map<string, string>());
        AppServices.userDataService.storeWizardTitles(this.wizardName, titles);
    }

    protected onWizardComplete() {
        this.storeWizardStepDescriptions();
    }

    protected storeMap(step: string, field: string, map: Map<string, string>) {
        const identifier = { wizard: this.supplyWizardName(), step, field };
        AppServices.userDataService.storeMap(identifier, map);
    }

    protected getStoredClusterPlan(step: string): string {
        return AppServices.userDataService.retrieveStoredValue(this.wizardName, step, {name: NodeSettingField.CLUSTER_PLAN});
    }

    protected setStoredClusterPlan(step, clusterPlan: string) {
        const identifier = {wizard: this.wizardName, step, field: NodeSettingField.CLUSTER_PLAN};
        AppServices.userDataService.store(identifier, {display: clusterPlan, value: clusterPlan});
    }
}
