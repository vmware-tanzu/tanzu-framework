import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class NodeSettings extends Step {
    hasMovedToStep() {
        return this.getDevSelect().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="vsphereNodeSettingForm"] clr-step-description'))
            .getText() as Promise<string>;
    }

    getDevSelect() {
        return element(by.css('select[name="devInstanceType"]'));
    }

    getMCName() {
        return element(by.css('input[formcontrolname="clusterName"]'));
    }

    getVirtualIpAddress() {
        return element(by.css('input[formcontrolname="controlPlaneEndpointIP"]'));
    }

    getWorkerNodeType() {
        return element(by.css('select[name="workerNodeInstanceType"]'));
    }

    getLoadBalancerInstanceType() {
        return element(by.css('select[name="haProxyInstanceType"]'));
    }

    getAdvancedSettings() {
        return element(by.css('clr-icon[data-e2e-id="advancedSettings"]'));
    }
}
