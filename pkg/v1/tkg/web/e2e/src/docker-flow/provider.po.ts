import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Provider extends Step {
    hasMovedToStep() {
        throw new Error("Method not implemented.");
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="dockerDaemonForm"] clr-step-description')).getText() as Promise<string>;
    }

}
