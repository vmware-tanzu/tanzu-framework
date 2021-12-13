import { WizardBaseDirective } from '../../wizard-base/wizard-base';
import { IdentityManagementType, WizardForm } from '../../constants/wizard.constants';
import { Component, Type } from '@angular/core';
import { AbstractControl, FormGroup } from '@angular/forms';
import { StepFormDirective } from '../../step-form/step-form';
import { WizardFormBase } from '../../../../../../shared/service/wizard-form-base';

interface I18nDataForHtml {
    title: string,
    description: string,
}
export interface FormDataForHTML {
    name: string,           // name of this step (not displayable)
    description: string,    // description of this step (displayed)
    title: string,          // title of this step (displayed)
    i18n: I18nDataForHtml,  // data used for navigating UI (displayed)
    // TODO: clazz should be required
    clazz?: Type<StepFormDirective>,    // the class of the step component
}

export class FormUtility {
    static IdentityFormDescription(wizard: WizardBaseDirective): string {
        const identityType = wizard.getFieldValue(WizardForm.IDENTITY, 'identityType');
        const ldapEndpointIp = wizard.getFieldValue(WizardForm.IDENTITY, 'endpointIp');
        const ldapEndpointPort = wizard.getFieldValue(WizardForm.IDENTITY, 'endpointPort');
        const oidcIssuer = wizard.getFieldValue(WizardForm.IDENTITY, 'issuerURL');

        if (identityType === IdentityManagementType.OIDC && oidcIssuer) {
            return 'OIDC configured: ' + oidcIssuer;
        } else if (identityType === IdentityManagementType.LDAP && ldapEndpointIp) {
            return 'LDAP configured: ' + ldapEndpointIp + ':' + ldapEndpointPort;
        }
        return 'Specify identity management';
    }

    static titleCase(target): string {
        if (target === undefined || target === null || target.length === 0) {
            return '';
        }
        return target.replace(/(^|\s)\S/g, function(t) { return t.toUpperCase() });
    }

    static formOverrideDescription(formData: FormDataForHTML, description: string): FormDataForHTML {
        formData.description = description;
        return formData;
    }

    static formOverrideClazz(formData: FormDataForHTML, clazz: Type<StepFormDirective>) {
        formData.clazz = clazz;
        return formData;
    }
}
