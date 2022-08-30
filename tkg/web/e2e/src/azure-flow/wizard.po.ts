import { browser, by, element } from 'protractor';
import { WizardBase } from '../wizard-base.po';

export class Wizard extends WizardBase {

    navigateTo() {
        browser.get(`http://${browser.params.SERVER_URL}/#/ui/azure/wizard`);
    }

    getTitleText() {
        return element(by.id('azure-title')).getText() as Promise<string>;
    }

}
