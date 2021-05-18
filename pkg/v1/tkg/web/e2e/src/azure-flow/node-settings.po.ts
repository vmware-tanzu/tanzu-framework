import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class NodeSettings extends Step {
    hasMovedToStep() {
        return this.getDevSelect().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="azureNodeSettingForm"] clr-step-description')).getText() as Promise<string>;
    }

    getDevSelect() {
        return element(by.css('select[name="devInstanceType"]'));
    }

    getMCName() {
        return element(by.css('input[formcontrolname="managementClusterName"]'));
    }

    getWorkNodeInstanceType() {
        return element(by.css('select[name="workerNodeInstanceType"]'));
    }
}
