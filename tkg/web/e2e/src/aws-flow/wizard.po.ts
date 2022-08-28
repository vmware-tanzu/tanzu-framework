import { browser, by, element } from 'protractor';
import { WizardBase } from '../wizard-base.po';

export class Wizard extends WizardBase {

    navigateTo() {
        browser.get(`http://${browser.params.SERVER_URL}/#/ui/aws/wizard`);
    }

    getTitleText() {
        return element(by.css('aws-wizard h2')).getText() as Promise<string>;
    }

}
