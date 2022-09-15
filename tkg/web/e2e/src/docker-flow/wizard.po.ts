import { browser, by, element } from 'protractor';
import { WizardBase } from '../wizard-base.po';

export class Wizard extends WizardBase {

    navigateTo() {
        browser.get(`http://${browser.params.SERVER_URL}/#/ui/docker/wizard`);
    }

    getTitleText() {
        return element(by.css('app-docker-wizard h2')).getText() as Promise<string>;
    }

}
