import { browser, by, element } from 'protractor';
import { Step } from '../step.po';

export class NetworkProxy extends Step {
    hasMovedToStep() {
        return this.getHttpProxyUrl().isPresent();
    }

    getProxyButton() {
        return element(by.cssContainingText('label', 'Enable Proxy Settings'));
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
        return element(by.cssContainingText('label', 'Use same configuration for https proxy'));
    }

    getNoProxy() {
        return element(by.css('input[formcontrolname="noProxy"]'));
    }

}
