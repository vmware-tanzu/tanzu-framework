import { browser, by, element } from 'protractor';

export class AppPage {
    navigateTo() {
        return browser.get(browser.baseUrl) as Promise<any>;
    }

    getTitleText() {
        return element(by.css('tkg-kickstart-ui-start h2')).getText() as Promise<string>;
    }

    getDeployOnVsphere() {
        return element(by.id("btn-deploy-vsphere"));
    }

    getDeployOnAws() {
        return element(by.id("btn-deploy-aws"));
    }

    getDeployOnAzure() {
        return element(by.id("btn-deploy-azure"));
    }

    getDeployOnDocker() {
        return element(by.id("btn-deploy-docker"));
    }

}
