import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class Network extends Step {
    hasMovedToStep() {
        return this.getNetworkName().isPresent();
    }

    getTitleText() {
        return element(by.css('clr-stepper-panel[formgroupname="networkForm"] clr-step-description')).getText() as Promise<string>;
    }

    getNetworkName() {
        return element(by.css('input[formcontrolname="networkName"]'));
    }

    getClusterServiceCidr() {
        return element(by.css('input[formcontrolname="clusterServiceCidr"]'));
    }

    getClusterPodCidr() {
        return element(by.css('input[formcontrolname="clusterPodCidr"]'));
    }

/*
    getProxyButton() {
        return element(by.cssContainingText('.clr-control-label', 'Enable Proxy Settings'));
    }

    getHttpProxyUrl() {
        return element(by.css('input[formcontrolname="httpProxyUrl"]'));
    }

    getHttpProxyUsername() {
        return element(by.css('input[formcontrolname="httpProxyUsername"]'));
    }

    getHttpProxyPassword() {
        return element(by.css('input[formcontrolname="httpProxyPassword"]'));
    }

    getIsSameAsHttp() {
        return element(by.cssContainingText('.clr-control-label', 'Use same configuration for https proxy'));
    }

    getNoProxy() {
        return element(by.css('input[formcontrolname="noProxy"]'));
    }
*/

}
