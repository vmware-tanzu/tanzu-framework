import { browser, by, element, protractor } from 'protractor';

export class AppPage {
    navigateTo() {
        return browser.get(browser.baseUrl) as Promise<any>;
    }

    getTitleText() {
        return element(by.css('tkg-kickstart-ui-start h2')).getText() as Promise<string>;
    }

    matchTitleText() {
        const EC = protractor.ExpectedConditions;
        return EC.or(
            EC.presenceOf(element(by.cssContainingText('tkg-kickstart-ui-start h2',
             'Welcome to the VMware Tanzu Kubernetes Grid Installer'))),
            EC.presenceOf(element(by.cssContainingText('tkg-kickstart-ui-start h2',
             'Welcome to the Tanzu Community Edition Installer')))
        );
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
