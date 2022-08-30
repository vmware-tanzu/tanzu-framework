import { browser, by, element, protractor } from 'protractor';
import { Step } from '../step.po';

export class Stepper extends Step {
    hasMovedToStep() {
        return this.isClickable(element(by.buttonText("REVIEW CONFIGURATION")));
    }

    isClickable(target) {
        return protractor.ExpectedConditions.elementToBeClickable(target);
    }

    getProviderStep() {
        return element(by.cssContainingText('clr-step-title', 'IaaS Provider'));
    }

    getVpcStep() {
        return element(by.cssContainingText('clr-step-title', 'VPC for AWS'));
    }

    getVnetStep() {
        return element(by.cssContainingText('clr-step-title', 'Azure VNet Settings'));
    }

    getNodeSettingsStep() {
        return element(by.cssContainingText('clr-step-title', 'Management Cluster Settings'));
    }

    getNsxLbStep() {
        return element(by.cssContainingText('clr-step-title', 'VMware NSX Advanced Load Balancer'));
    }

    getResourceStep() {
        return element(by.cssContainingText('clr-step-title', 'Resources'));
    }

    getMetadataStep() {
        return element(by.cssContainingText('clr-step-title', 'Metadata'));
    }

    getNetworkStep() {
        return element(by.cssContainingText('clr-step-title', 'Kubernetes Network'));
    }

    getIdentityStep() {
        return element(by.cssContainingText('clr-step-title', 'Identity Management'));
    }

    getOsImageStep() {
        return element(by.cssContainingText('clr-step-title', 'OS Image'));
    }

    getTmcStep() {
        return element(by.cssContainingText('clr-step-title', 'Register TMC'));
    }

    getCeipStep() {
        return element(by.cssContainingText('clr-step-title', 'CEIP Agreement'));
    }

}
