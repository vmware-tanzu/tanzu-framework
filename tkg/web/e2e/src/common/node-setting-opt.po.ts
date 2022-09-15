import { by, element } from 'protractor';
import { Step } from '../step.po';

export class NodeOpt extends Step {
    hasMovedToStep() {
        return this.getEnableMHC().isPresent();
    }

    getEnableAudit() {
        return element(by.css('app-audit-logging clr-checkbox-wrapper label'));
    }

    getEnableMHC() {
        return element(by.css('input[formcontrolname="machineHealthChecksEnabled"]'));
    }

    getEnableBH() {
        return element(by.css('input[formcontrolname="bastionHostEnabled"]'));
    }
}
