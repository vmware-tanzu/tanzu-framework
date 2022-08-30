import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class TmcRegister extends Step {
    hasMovedToStep() {
        return this.getRegUrl().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="registerTmcForm"] clr-step-description')).getText() as Promise<string>;
    }

    getRegUrl() {
        return element(by.css('input[formcontrolname="tmcRegUrl"]'));
    }

    getTmcButton() {
        return element(by.css('button.btn btn-sm'));
    }

}
