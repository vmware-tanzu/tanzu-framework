import { Component, OnInit } from '@angular/core';
import { FormControl, Validators } from '@angular/forms';
import { distinctUntilChanged, takeUntil } from 'rxjs/operators';
import { StepFormDirective } from '../../../step-form/step-form';
import { ValidationService } from '../../../validation/validation.service';

const oidcFields: Array<string> = [
    'issuerURL',
    'clientId',
    'clientSecret',
    'scopes',
    'oidcUsernameClaim',
    'oidcGroupsClaim'
];

const ldapValidatedFields: Array<string> = [
    'endpointIp',
    'endpointPort',
    'bindPW'
];

const ldapNonValidatedFields: Array<string> = [
    'bindDN',
    'userSearchBaseDN',
    'userSearchFilter',
    'userSearchUsername',
    'groupSearchBaseDN',
    'groupSearchFilter',
    'groupSearchUserAttr',
    'groupSearchGroupAttr',
    'groupSearchNameAttr',
    'ldapRootCAData'
];

@Component({
    selector: 'app-shared-identity-step',
    templateUrl: './identity-step.component.html',
    styleUrls: ['./identity-step.component.scss']
})
export class SharedIdentityStepComponent extends StepFormDirective implements OnInit {
    identityTypeValue: string = 'oidc'

    fields: Array<string> = [...oidcFields, ...ldapValidatedFields, ...ldapNonValidatedFields];

    constructor(private validationService: ValidationService) {
        super();
    }

    ngOnInit(): void {
        super.ngOnInit();

        this.formGroup.addControl('identityType', new FormControl('oidc', []));
        this.formGroup.addControl('idmSettings', new FormControl(true, []));

        this.fields.forEach(field => this.formGroup.addControl(field, new FormControl('', [])));

        this.formGroup.get('identityType').valueChanges.pipe(
            distinctUntilChanged((prev, curr) => JSON.stringify(prev) === JSON.stringify(curr)),
            takeUntil(this.unsubscribe)
        ).subscribe(data => {
            this.identityTypeValue = data;
            this.unsetAllValidators();
            if (this.identityTypeValue === 'oidc') {
                this.setOIDCValidators();
                this.formGroup.get('clientSecret').setValue('');
            } else if (this.identityTypeValue === 'ldap') {
                this.setLDAPValidators();
            } else {
                this.disarmField('identityType', true);
            }
        });
        this.identityTypeValue = this.getSavedValue('identityType', 'oidc');

        this.formGroup.get('identityType').setValue(this.identityTypeValue);
    }

    setOIDCValidators() {
        this.resurrectField('issuerURL', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidIpOrFqdnWithHttpsProtocol(),
            this.validationService.isStringWithoutUrlFragment(),
            this.validationService.isStringWithoutQueryParams(),
        ], this.getSavedValue('issuerURL', ''));

        this.resurrectField('clientId', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds()
        ], this.getSavedValue('clientId', ''));

        this.resurrectField('clientSecret', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds()
        ], '');

        this.resurrectField('scopes', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isCommaSeperatedList()
        ], this.getSavedValue('scopes', ''));

        this.resurrectField('oidcUsernameClaim', [
           Validators.required
        ], this.getSavedValue('oidcUsernameClaim', ''));

        this.resurrectField('oidcGroupsClaim', [
            Validators.required
        ], this.getSavedValue('oidcGroupsClaim', ''));
    }

    setLDAPValidators() {
        this.resurrectField('endpointIp', [
            Validators.required
        ], this.getSavedValue('endpointIp', ''));

        this.resurrectField('endpointPort', [
            Validators.required,
            this.validationService.noWhitespaceOnEnds(),
            this.validationService.isValidLdap(this.formGroup.get('endpointIp'))
        ], this.getSavedValue('endpointPort', ''));

        this.resurrectField('bindPW', [], '');

        ldapNonValidatedFields.forEach(field => this.resurrectField(
            field, [], this.getSavedValue(field, '')));
    }

    unsetAllValidators() {
        this.fields.forEach(field => this.disarmField(field, true));
    }

    toggleIdmSetting() {
        if (this.formGroup.value['idmSettings']) {
            this.formGroup.controls['identityType'].setValue('oidc');
        } else {
            this.formGroup.controls['identityType'].setValue('none');
        }
    }

    setSavedDataAfterLoad() {
        super.setSavedDataAfterLoad();
        this.formGroup.get('clientSecret').setValue('');
        if (!this.formGroup.value['idmSettings']) {
            this.formGroup.get('identityType').setValue('none');
        }
    }

    /**
     * @method ldapEndpointInputValidity return true if ldap endpoint inputs are valid
     */
    ldapEndpointInputValidity(): boolean {
        return this.formGroup.get('endpointIp').valid &&
            this.formGroup.get('endpointPort').valid;
    }
}
