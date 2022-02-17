// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// Third party imports
import { distinctUntilChanged, takeUntil, tap } from 'rxjs/operators';
// App imports
import { APIClient } from 'src/app/swagger';
import AppServices from '../../../../../../../shared/service/appServices';
import { IdentityField, IdentityStepMapping } from './identity-step.fieldmapping';
import { IdentityManagementType } from '../../../constants/wizard.constants';
import { IpFamilyEnum } from 'src/app/shared/constants/app.constants';
import { LdapParams } from './../../../../../../../swagger/models/ldap-params.model';
import { LdapTestResult } from 'src/app/swagger/models';
import { StepFormDirective } from '../../../step-form/step-form';
import { ValidationService } from '../../../validation/validation.service';

const CONNECT = "CONNECT";
const BIND = "BIND";
const USER_SEARCH = "USER_SEARCH";
const GROUP_SEARCH = "GROUP_SEARCH";
const DISCONNECT = "DISCONNECT";

const TEST_SUCCESS = 1;

const LDAP_TESTS = [CONNECT, BIND, USER_SEARCH, GROUP_SEARCH, DISCONNECT];

const NOT_STARTED = "not-started";
const SUCCESS = "success";
const ERROR = "error";
const PROCESSING = "processing";

const oidcFields: Array<string> = [
    IdentityField.ISSUER_URL,
    IdentityField.CLIENT_ID,
    IdentityField.CLIENT_SECRET,
    IdentityField.SCOPES,
    IdentityField.OIDC_GROUPS_CLAIM,
    IdentityField.OIDC_USERNAME_CLAIM
];

const ldapValidatedFields: Array<string> = [
    IdentityField.ENDPOINT_IP,
    IdentityField.ENDPOINT_PORT,
    IdentityField.BIND_PW,
    IdentityField.GROUP_SEARCH_FILTER,
    IdentityField.USER_SEARCH_FILTER,
    IdentityField.USER_SEARCH_USERNAME
];

const ldapNonValidatedFields: Array<string> = [
    IdentityField.BIND_DN,
    IdentityField.USER_SEARCH_BASE_DN,
    IdentityField.GROUP_SEARCH_BASE_DN,
    IdentityField.GROUP_SEARCH_USER_ATTR,
    IdentityField.GROUP_SEARCH_GROUP_ATTR,
    IdentityField.GROUP_SEARCH_NAME_ATTR,
    IdentityField.LDAP_ROOT_CA,
    IdentityField.TEST_USER_NAME,
    IdentityField.TEST_GROUP_NAME
];

const LDAP_PARAMS = {
    ldap_bind_dn: "bindDN",
    ldap_bind_password: "bindPW",
    ldap_group_search_base_dn: "groupSearchBaseDN",
    ldap_group_search_filter: "groupSearchFilter",
    ldap_group_search_group_attr: "groupSearchGroupAttr",
    ldap_group_search_name_attr: "groupSearchNameAttr",
    ldap_group_search_user_attr: "groupSearchUserAttr",
    ldap_root_ca: "ldapRootCAData",
    ldap_user_search_base_dn: "userSearchBaseDN",
    ldap_user_search_filter: "userSearchFilter",
    ldap_user_search_name_attr: "userSearchUsername",
    ldap_user_search_username: "userSearchUsername",
    ldap_test_group: "testGroupName",
    ldap_test_user: "testUserName"
}

@Component({
    selector: 'app-shared-identity-step',
    templateUrl: './identity-step.component.html',
    styleUrls: ['./identity-step.component.scss']
})
export class SharedIdentityStepComponent extends StepFormDirective implements OnInit {
    static description = 'Optionally specify identity management';

    _verifyLdapConfig = false;

    fields: Array<string> = [...oidcFields, ...ldapValidatedFields, ...ldapNonValidatedFields];

    timelineState = {};
    timelineError = {};

    private usingIdmSettings = false;
    private idmSettingType;

    constructor(private apiClient: APIClient,
                private validationService: ValidationService) {
        super();
        this.resetTimelineState();
    }

    private customizeForm() {
        this.registerOnIpFamilyChange(IdentityField.ISSUER_URL, [], [], () => {
            if (this.isUsingIdentityManagement) {
                if (this.isIdentityManagementOidc) {
                    this.setOIDCValidators();
                } else if (this.isIdentityManagementLdap) {
                    this.setLDAPValidators();
                }
            }
        });
        this.formGroup.get(IdentityField.IDENTITY_TYPE).valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(newIdentityManagementType => {
            this.idmSettingType = newIdentityManagementType;
            this.formGroup.markAsPending();
            if (this.isIdentityManagementOidc) {
                this.unsetValidators(ldapValidatedFields);
                this.setOIDCValidators();
                this.setControlValueSafely(IdentityField.CLIENT_SECRET, '');
            } else if (this.isIdentityManagementLdap) {
                this.unsetValidators(oidcFields);
                this.setLDAPValidators();
            } else {
                this.unsetValidators(this.fields);
                this.disarmField(IdentityField.IDENTITY_TYPE);
            }
            this.triggerStepDescriptionChange();
        });
        this.registerStepDescriptionTriggers({fields: [IdentityField.ENDPOINT_IP, IdentityField.ENDPOINT_PORT,  IdentityField.ISSUER_URL]});
    }

    ngOnInit(): void {
        super.ngOnInit();
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, IdentityStepMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(IdentityStepMapping);
        this.storeDefaultLabels(IdentityStepMapping);
        this.registerDefaultFileImportedHandler(this.eventFileImported, IdentityStepMapping);
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);

        this.customizeForm();
    }

    protected onStepStarted() {
        // these fields should have been restored during ngOnInit()
        this.idmSettingType = this.getFieldValue(IdentityField.IDENTITY_TYPE);
        this.usingIdmSettings = this.getFieldValue(IdentityField.IDM_SETTINGS);
    }

    setOIDCValidators() {
        this.resurrectField(IdentityField.ISSUER_URL, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidIpOrFqdnWithHttpsProtocol() : this.validationService.isValidIpv6OrFqdnWithHttpsProtocol(),
            this.validationService.isStringWithoutUrlFragment(),
            this.validationService.isStringWithoutQueryParams(),
        ], this.getStoredValue(IdentityField.ISSUER_URL, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.CLIENT_ID, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds()
        ], this.getStoredValue(IdentityField.CLIENT_ID, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.CLIENT_SECRET, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds()
        ], '');

        this.resurrectField(IdentityField.SCOPES, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isCommaSeperatedList()
        ], this.getStoredValue(IdentityField.SCOPES, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.OIDC_GROUPS_CLAIM, [
            Validators.required
        ], this.getStoredValue(IdentityField.OIDC_GROUPS_CLAIM, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.OIDC_USERNAME_CLAIM, [
            Validators.required
        ], this.getStoredValue(IdentityField.OIDC_USERNAME_CLAIM, IdentityStepMapping, ''));
    }

    setLDAPValidators() {
        this.resurrectField(IdentityField.ENDPOINT_IP, [
            Validators.required
        ], this.getStoredValue(IdentityField.ENDPOINT_IP, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.ENDPOINT_PORT, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidLdap(this.formGroup.get(IdentityField.ENDPOINT_IP)) :
                this.validationService.isValidIpv6Ldap(this.formGroup.get(IdentityField.ENDPOINT_IP))
        ], this.getStoredValue(IdentityField.ENDPOINT_PORT, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.BIND_PW, [], '');

        this.resurrectField(IdentityField.USER_SEARCH_FILTER, [
            Validators.required
        ], this.getStoredValue(IdentityField.USER_SEARCH_FILTER, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.USER_SEARCH_USERNAME, [
            Validators.required
        ], this.getStoredValue(IdentityField.USER_SEARCH_USERNAME, IdentityStepMapping, ''));

        this.resurrectField(IdentityField.GROUP_SEARCH_FILTER, [
            Validators.required
        ], this.getStoredValue(IdentityField.GROUP_SEARCH_FILTER, IdentityStepMapping, ''));

        ldapNonValidatedFields.forEach(field => this.resurrectField(
            field, [], this.getStoredValue(field, IdentityStepMapping, '')));
    }

    unsetValidators(fields: string[]) {
        fields.forEach(field => this.disarmField(field));
    }

    toggleIdmSetting() {
        this.usingIdmSettings = !this.usingIdmSettings;
        if (this.usingIdmSettings) {
            if (this.isIdentityManagementOidc) {
                this.setOIDCValidators();
                this.setControlValueSafely(IdentityField.CLIENT_SECRET, '');
            } else if (this.isIdentityManagementLdap) {
                this.setLDAPValidators();
            } else {
                // The type in use by default when user activated identity management is OIDC;
                this.idmSettingType = IdentityManagementType.OIDC;
                // setting the idm setting type should trigger the handler to set the oidc validators, etc.
                this.setControlValueSafely(IdentityField.IDENTITY_TYPE, this.idmSettingType);
            }
        } else {
            this.unsetValidators(this.fields);
        }
    }

    /**
     * @method ldapEndpointInputValidity return true if ldap endpoint inputs are valid
     */
    ldapEndpointInputValidity(): boolean {
        return this.formGroup.get(IdentityField.ENDPOINT_IP).valid &&
            this.formGroup.get(IdentityField.ENDPOINT_PORT).valid;
    }

    resetTimelineState() {
        LDAP_TESTS.forEach(t => {
            this.timelineState[t] = NOT_STARTED;
            this.timelineError[t] = null;
        })
    }

    cropLdapConfig(): LdapParams {
        const ldapParams: LdapParams = {};

        Object.entries(LDAP_PARAMS).forEach(([k, v]) => {
            if (this.formGroup.get(v)) {
                ldapParams[k] = this.formGroup.get(v).value || "";
            } else {
                console.log("Unable to find field: " + v);
            }
        });
        ldapParams.ldap_url = "ldaps://" + this.formGroup.get(IdentityField.ENDPOINT_IP).value + ':' +
            this.formGroup.get(IdentityField.ENDPOINT_PORT).value;

        return ldapParams;
    }

    formatError(err) {
        if (err) {
            const errMsg = err.error ? err.error.message : null;
            return errMsg || err.message || JSON.stringify(err, null, 4);
        }
        return "";
    }

    async startVerifyLdapConfig() {
        this.resetTimelineState();
        const params = this.cropLdapConfig();

        console.log(JSON.stringify(params, null, 8));
        let result: LdapTestResult;
        try {
            this.timelineState[CONNECT] = PROCESSING;
            result = await this.apiClient.verifyLdapConnect({ credentials: params }).toPromise();
            this.timelineState[CONNECT] = result && (result.code === TEST_SUCCESS ? SUCCESS : NOT_STARTED);
        } catch (err) {
            console.log(JSON.stringify(err, null, 8));
            this.timelineState[CONNECT] = ERROR;
            this.timelineError[CONNECT] = this.formatError(err);
        }

        try {
            this.timelineState[BIND] = PROCESSING;
            result = await this.apiClient.verifyLdapBind().toPromise();
            this.timelineState[BIND] = result && (result.code === TEST_SUCCESS ? SUCCESS : NOT_STARTED); ;
        } catch (err) {
            console.log(JSON.stringify(err, null, 8));
            this.timelineState[BIND] = ERROR;
            this.timelineError[BIND] = this.formatError(err);
        }

        try {
            this.timelineState[USER_SEARCH] = PROCESSING;
            result = await this.apiClient.verifyLdapUserSearch().toPromise();
            this.timelineState[USER_SEARCH] = result && (result.code === TEST_SUCCESS ? SUCCESS : NOT_STARTED); ;
        } catch (err) {
            console.log(JSON.stringify(err, null, 8));
            this.timelineState[USER_SEARCH] = ERROR;
            this.timelineError[USER_SEARCH] = this.formatError(err);
        }

        try {
            this.timelineState[GROUP_SEARCH] = PROCESSING;
            result = await this.apiClient.verifyLdapGroupSearch().toPromise();
            this.timelineState[GROUP_SEARCH] = result && (result.code === TEST_SUCCESS ? SUCCESS : NOT_STARTED); ;
        } catch (err) {
            console.log(JSON.stringify(err, null, 8));
            this.timelineState[GROUP_SEARCH] = ERROR;
            this.timelineError[GROUP_SEARCH] = this.formatError(err);
        }

        try {
            this.timelineState[DISCONNECT] = PROCESSING;
            await this.apiClient.verifyLdapCloseConnection().toPromise();
            this.timelineState[DISCONNECT] = SUCCESS;
        } catch (err) {
            console.log(JSON.stringify(err, null, 8));
            this.timelineState[DISCONNECT] = ERROR;
            this.timelineError[DISCONNECT] = this.formatError(err);
        }
    }

    get verifyLdapConfig(): boolean {
        return this._verifyLdapConfig;
    }

    set verifyLdapConfig(vlc: boolean) {
        this._verifyLdapConfig = vlc;
        this.resetTimelineState();
    }

    dynamicDescription(): string {
        const identityType = this.getFieldValue(IdentityField.IDENTITY_TYPE, true);
        const ldapEndpointIp = this.getFieldValue(IdentityField.ENDPOINT_IP, true);
        const ldapEndpointPort = this.getFieldValue(IdentityField.ENDPOINT_PORT, true);
        const oidcIssuer = this.getFieldValue(IdentityField.ISSUER_URL, true);

        if (identityType === IdentityManagementType.OIDC && oidcIssuer) {
            return 'OIDC configured: ' + oidcIssuer;
        } else if (identityType === IdentityManagementType.LDAP && ldapEndpointIp) {
            return 'LDAP configured: ' + ldapEndpointIp + ':' + (ldapEndpointPort ? ldapEndpointPort : '');
        }
        return SharedIdentityStepComponent.description;
    }

    protected storeUserData() {
        this.clearUnusedData();
        this.storeUserDataFromMapping(IdentityStepMapping);
        this.storeDefaultDisplayOrder(IdentityStepMapping);
    }

    // At the END of the step, we clear from memory all the unused values. This will allow us to present a "clean" confirmation page,
    // and also remove any lingering values in local storage that are no longer relevant
    private clearUnusedData() {
        if (!this.isUsingIdentityManagement) {
            this.clearControlValue(IdentityField.IDENTITY_TYPE, true);
            // NOTE: by clearing this setting the tests below will be true, and all the fields will be cleared
            this.idmSettingType = null;
        }
        if (!this.isIdentityManagementOidc) {
            oidcFields.forEach(field => this.clearControlValue(field, true));
        }
        if (!this.isIdentityManagementLdap) {
            [...ldapValidatedFields, ...ldapNonValidatedFields].forEach(field => this.clearControlValue(field, true));
        }
    }

    get isIdentityManagementOidc(): boolean {
        return this.idmSettingType === IdentityManagementType.OIDC;
    }

    get isIdentityManagementLdap(): boolean {
        return this.idmSettingType === IdentityManagementType.LDAP;
    }

    get isUsingIdentityManagement(): boolean {
        return this.usingIdmSettings;
    }
}
