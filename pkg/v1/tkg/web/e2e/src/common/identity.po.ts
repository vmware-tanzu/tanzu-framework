import { by, element } from 'protractor';
import { Step } from '../step.po';

export class Identity extends Step {
    hasMovedToStep() {
        return this.getOIDCRadioButton().isPresent();
    }

    getOIDCRadioButton() {
        return element(by.cssContainingText("label", "OIDC"));
    }

    getLDAPRadioButton() {
        return element(by.cssContainingText("label", "LDAP"));
    }

    getPorviderNameInput() {
        return element(by.css('input[formcontrolname="providerName"]'));
    }

    getIssuerURLInput() {
        return element(by.css('input[formcontrolname="issuerURL"]'));
    }

    getClientIdInput() {
        return element(by.css('input[formcontrolname="clientId"]'));
    }

    getClientSecretInput() {
        return element(by.css('input[formcontrolname="clientSecret"]'));
    }

    getScopesInput() {
        return element(by.css('input[formcontrolname="scopes"]'));
    }

    getOidcUsernameClaim() {
        return element(by.css('input[formcontrolname="oidcUsernameClaim"]'));
    }

    getOidcGroupsClaim() {
        return element(by.css('input[formcontrolname="oidcGroupsClaim"]'));
    }

    getOIDCLabelsKey() {
        return element(by.css('input[formcontrolname="newOidcClaimKey"]'));
    }

    getOIDCLabelsValue() {
        return element(by.css('input[formcontrolname="newOidcClaimValue"]'));
    }

    getOIDCLabelsAddButton() {
        return element(by.buttonText('ADD'));
    }

    getOIDCLabelsDeleteButton(key: string) {
        const id = "claim-delete-" + key;
        return element(by.id(id));
    }

    getSkipVerifyCertificateCheckbox() {
        return element(by.cssContainingText("label", "Skip verify certificate"));
    }

    getLdapIP() {
        return element(by.css('input[formcontrolname="endpointIp"]'));
    }

    getLdapPort() {
        return element(by.css('input[formcontrolname="endpointPort"]'));
    }

    getRootCAData() {
        return element(by.css('textarea[formcontrolname="rootCAData"]'));
    }

}
