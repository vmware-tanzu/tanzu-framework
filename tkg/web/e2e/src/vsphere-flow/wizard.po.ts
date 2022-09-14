import { browser, by, element, protractor } from 'protractor';
import { WizardBase } from '../wizard-base.po';

export class Wizard extends WizardBase {

    navigateTo() {
        browser.get(`http://${browser.params.SERVER_URL}/#/ui/wizard`);
        browser.sleep(10000);
    }

    getTitleText() {
        return element(by.css('app-wizard h2')).getText() as Promise<string>;
    }

}
