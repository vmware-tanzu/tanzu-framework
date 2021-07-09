import { by, element } from 'protractor';
import { Step } from '../step.po';

export class Ceip extends Step {
    hasMovedToStep() {
        return this.getCeipCheckbox().isPresent();
    }

    getCeipCheckbox() {
        return element(by.css('input[formcontrolname="ceipOptIn"]'));
    }
}
