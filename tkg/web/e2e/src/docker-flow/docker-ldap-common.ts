import { Provider } from './provider.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { browser } from 'protractor';
import { NetworkProxy } from '../common/networkproxy.po';
import { Identity } from '../common/identity.po'

export class DockerLdapCommon extends WizardCommon {

    setNetworkProxy(step: NetworkProxy) {
        step.getProxyButton().click();
        browser.waitForAngular();
        step.getHttpProxyUrl().sendKeys(PARAMS.DEFAULT_PROXY_HTTP_URL);
        step.getIsSameAsHttp().click();
    }

    getFlowTestingDescription() {
        return "Docker flow (Ldap)"
    }

    executeIdentityStep() {
        describe("Identity step", () => {
            const identity = new Identity();

            it('should have moved to this step', () => {
                expect(identity.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                identity.getLDAPRadioButton().click();
                identity.getLdapIP().sendKeys(PARAMS.LDAP_IP);
                identity.getLdapPort().sendKeys(PARAMS.LDAP_PORT);
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                identity.getNextButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
            })
        });
    }

}
