import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class NsxSettings extends Step {
    hasMovedToStep() {
        return this.getControllerHost().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="loadBalancerForm"] clr-step-description')).getText() as Promise<string>;
    }

    getControllerHost() {
        return element(by.css('input[formcontrolname="controllerHost"]'));
    }

    getNsxUserName() {
        return element(by.css('input[formcontrolname="username"]'));
    }

    getNsxPassword() {
        return element(by.css('input[formcontrolname="password"]'));
    }

    getNsxVerifyButton() {
        return element(by.css('button.btn-connect'));
    }

    getCloudName() {
        return element(by.css('select[formcontrolname="cloudName"]'));
    }

    getServiceEngine() {
        return element(by.css('select[formcontrolname="serviceEngineGroupName"]'));
    }

    getNetworkName() {
        return element(by.css('select[formcontrolname="networkName"]'));
    }

    getNetworkCIDR() {
        return element(by.css('select[formcontrolname="networkCIDR"]'));
    }

    getControllerCert() {
        return element(by.css('textarea[formcontrolname="controllerCert"]'));
    }

    getNsxLabelsKey() {
        return element(by.css('input[formcontrolname="newLabelKey"]'));
    }

    getNsxLabelsValue() {
        return element(by.css('input[formcontrolname="newLabelValue"]'));
    }

    getNsxLabelsAddButton() {
        return element(by.buttonText('ADD'));
    }

    getNsxLabelsDeleteButton(key: string) {
        const id = "label-delete-" + key;
        return element(by.id(id));
    }

}
