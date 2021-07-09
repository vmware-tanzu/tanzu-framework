import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Vpc extends Step {
    hasMovedToStep() {
        return this.getSelectAnExistingVpc().isPresent();
    }

    getSelectAnExistingVpc() {
        return element(by.cssContainingText("label", "Select an existing VPC"));
    }

    getVpcId() {
        return element(by.css('select[name="existingVpcId"]'));
    }

    getVpcCidr() {
        return element(by.css('input[formcontrolname="existingVpcCidr"]'));
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="vpcForm"] clr-step-description')).getText() as Promise<string>;
    }

}
