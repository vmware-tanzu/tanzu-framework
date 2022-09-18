import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Provider extends Step {
    hasMovedToStep() {
        return this.getUsername().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="vsphereProviderForm"] clr-step-description')).getText() as Promise<string>;
    }

    getVC() {
        return element(by.css('input[formcontrolname="vcenterAddress"]'));
    }

    getUsername() {
        return element(by.css('input[formcontrolname="username"]'));
    }

    getPassword() {
        return element(by.css('input[formcontrolname="password"]'));
    }

    getDC() {
        return element(by.css('select[formcontrolname="datacenter"]'));
    }

    getSSHKey() {
        return element(by.css('textarea[formcontrolname="ssh_key"]'));
    }

    getConectButton() {
        return element(by.css('button.btn-connect'));
    }

    getThumbprintContinueButton() {
        return element(by.css('app-ssl-thumbprint-modal .continue-btn'));
    }
}
