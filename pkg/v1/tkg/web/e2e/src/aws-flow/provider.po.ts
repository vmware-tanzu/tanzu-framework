import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Provider extends Step {
    hasMovedToStep() {
        return this.getCredentialProfile().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="awsProviderForm"] clr-step-description')).getText() as Promise<string>;
    }

    getCredentialProfile() {
        return element(by.cssContainingText("label", "Credential Profile"));
    }

    getStaticCredential() {
        return element(by.cssContainingText("label", "Static Credentials"));
    }

    getAccessKeyId() {
        return element(by.css('input[formcontrolname="accessKeyID"]'));
    }

    getSecretAccessKey() {
        return element(by.css('input[formcontrolname="secretAccessKey"]'));
    }

    getSshKeyName() {
        return element(by.css('input[formcontrolname="sshKeyName"]'));
    }

    getRegion() {
        return element(by.css('select[name="region"]'));
    }

    getConectButton() {
        return element(by.css('button.btn-connect'));
    }

    getProfileName() {
        return element(by.css('select[name="profileName"]'));
    }

}
