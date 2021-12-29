// Angular imports
import { Component, OnInit } from '@angular/core';
import { Validators } from '@angular/forms';
// Third party imports
import { distinctUntilChanged, takeUntil, tap } from 'rxjs/operators';
// App imports
import { APIClient } from 'src/app/swagger';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
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
const TEST_SKIPPED = 2;

const LDAP_TESTS = [CONNECT, BIND, USER_SEARCH, GROUP_SEARCH, DISCONNECT];

const NOT_STARTED = "not-started";
const CURRENT = "current";
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

    identityTypeValue: string = IdentityManagementType.OIDC;
    _verifyLdapConfig = false;

    fields: Array<string> = [...oidcFields, ...ldapValidatedFields, ...ldapNonValidatedFields];

    timelineState = {};
    timelineError = {};

    constructor(private apiClient: APIClient,
                private validationService: ValidationService) {
        super();
        this.resetTimelineState();
    }

    private customizeForm() {
        this.registerOnIpFamilyChange(IdentityField.ISSUER_URL, [], [], () => {
            if (this.identityTypeValue === IdentityManagementType.OIDC) {
                this.setOIDCValidators();
            } else if (this.identityTypeValue === IdentityManagementType.LDAP) {
                this.setLDAPValidators();
            }
        });
        this.formGroup.get(IdentityField.IDENTITY_TYPE).valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(data => {
            this.identityTypeValue = data;
            this.unsetAllValidators();
            this.formGroup.markAsPending();
            if (this.identityTypeValue === IdentityManagementType.OIDC) {
                this.setOIDCValidators();
                this.setControlValueSafely(IdentityField.CLIENT_SECRET, '');
            } else if (this.identityTypeValue === IdentityManagementType.LDAP) {
                this.setLDAPValidators();
            } else {
                this.disarmField(IdentityField.IDENTITY_TYPE, true);
            }
            this.triggerStepDescriptionChange();
        });
        this.registerStepDescriptionTriggers({fields: [IdentityField.ENDPOINT_IP, IdentityField.ENDPOINT_PORT,  IdentityField.ISSUER_URL]});
    }

    ngOnInit(): void {
        super.ngOnInit();
        AppServices.fieldMapUtilities.buildForm(this.formGroup, this.formName, IdentityStepMapping);
        this.htmlFieldLabels = Broker.fieldMapUtilities.getFieldLabelMap(IdentityStepMapping);
        this.storeDefaultLabels(IdentityStepMapping);

        this.customizeForm();

        this.initFormWithSavedData();
        this.identityTypeValue = this.getSavedValue(IdentityField.IDENTITY_TYPE, IdentityManagementType.OIDC);
        this.setControlValueSafely(IdentityField.IDENTITY_TYPE, this.identityTypeValue, { emitEvent: false });
    }

    setOIDCValidators() {
        this.resurrectField(IdentityField.ISSUER_URL, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidIpOrFqdnWithHttpsProtocol() : this.validationService.isValidIpv6OrFqdnWithHttpsProtocol(),
            this.validationService.isStringWithoutUrlFragment(),
            this.validationService.isStringWithoutQueryParams(),
        ], this.getSavedValue(IdentityField.ISSUER_URL, ''));

        this.resurrectField(IdentityField.CLIENT_ID, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds()
        ], this.getSavedValue(IdentityField.CLIENT_ID, ''));

        this.resurrectField(IdentityField.CLIENT_SECRET, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds()
        ], '');

        this.resurrectField(IdentityField.SCOPES, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isCommaSeperatedList()
        ], this.getSavedValue(IdentityField.SCOPES, ''));

        this.resurrectField(IdentityField.OIDC_GROUPS_CLAIM, [
            Validators.required
        ], this.getSavedValue(IdentityField.OIDC_GROUPS_CLAIM, ''));

        this.resurrectField(IdentityField.OIDC_USERNAME_CLAIM, [
            Validators.required
        ], this.getSavedValue(IdentityField.OIDC_USERNAME_CLAIM, ''));
    }

    setLDAPValidators() {
        this.resurrectField(IdentityField.ENDPOINT_IP, [
            Validators.required
        ], this.getSavedValue(IdentityField.ENDPOINT_IP, ''));

        this.resurrectField(IdentityField.ENDPOINT_PORT, [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.ipFamily === IpFamilyEnum.IPv4 ?
                this.validationService.isValidLdap(this.formGroup.get(IdentityField.ENDPOINT_IP)) :
                this.validationService.isValidIpv6Ldap(this.formGroup.get(IdentityField.ENDPOINT_IP))
        ], this.getSavedValue(IdentityField.ENDPOINT_PORT, ''));

        this.resurrectField(IdentityField.BIND_PW, [], '');

        this.resurrectField(IdentityField.USER_SEARCH_FILTER, [
            Validators.required
        ], this.getSavedValue(IdentityField.USER_SEARCH_FILTER, ''));

        this.resurrectField(IdentityField.USER_SEARCH_USERNAME, [
            Validators.required
        ], this.getSavedValue(IdentityField.USER_SEARCH_USERNAME, ''));

        this.resurrectField(IdentityField.GROUP_SEARCH_FILTER, [
            Validators.required
        ], this.getSavedValue(IdentityField.GROUP_SEARCH_FILTER, ''));

        ldapNonValidatedFields.forEach(field => this.resurrectField(
            field, [], this.getSavedValue(field, '')));
    }

    unsetAllValidators() {
        this.fields.forEach(field => this.disarmField(field, true));
    }

    toggleIdmSetting() {
        const identityType = this.formGroup.value[IdentityField.IDM_SETTINGS] ? IdentityManagementType.OIDC : IdentityManagementType.NONE;
        // onlySelf option will update the changes for the current control only
        this.setControlValueSafely(IdentityField.IDENTITY_TYPE, identityType, { onlySelf: true });
    }

    initFormWithSavedData() {
        super.initFormWithSavedData();
        this.scrubPasswordField(IdentityField.CLIENT_SECRET);

        if (!this.formGroup.value[IdentityField.IDM_SETTINGS]) {
            this.setControlValueSafely(IdentityField.IDENTITY_TYPE, IdentityManagementType.NONE);
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
        this.storeUserDataFromMapping(IdentityStepMapping);
        this.storeDefaultDisplayOrder(IdentityStepMapping);
    }
}
