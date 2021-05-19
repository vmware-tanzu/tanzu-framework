import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class NodeSettings extends Step {
    hasMovedToStep() {
        return this.getDevSelect().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="awsNodeSettingForm"] clr-step-description')).getText() as Promise<string>;
    }

    getDevSelect() {
        return element(by.css('input[name="devInstanceType"]'));
    }

    getMCName() {
        return element(by.css('input[formcontrolname="clusterName"]'));
    }

    getWorkNodeInstanceType() {
        return element(by.css('select[name="workerNodeInstanceType"]'));
    }

    getSshKeyName() {
        return element(by.css('input[formcontrolname="sshKeyName"]'));
    }

    getAvailabilityZone() {
        return element(by.css('select[name="awsNodeAz1"]'));
    }

    getVpcPublicSubset() {
        return element(by.css('select[name="vpcPublicSubnet1"]'));
    }

    getVpcPrivateSubset() {
        return element(by.css('select[name="vpcPrivateSubnet1"]'));
    }
}
