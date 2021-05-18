import { Provider } from './provider.po';
import { NodeSettings } from './node-settings.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { Vpc } from './vpc.po';
import { browser } from 'protractor';
import { NetworkProxy } from '../common/networkproxy.po';
import { Identity } from '../common/identity.po'

export class NewVpcCommon extends WizardCommon {

    setCredentials(step: Provider) {
        step.getProfileName().click();
        step.selectOptionByText(step.getProfileName(), PARAMS.AWS_PROFILE);
    }

    selectSubnets(step: NodeSettings) {
        // Do nothing as it is not applicable
    }

    setNetworkProxy(step: NetworkProxy) {
        step.getProxyButton().click();
        browser.waitForAngular();
        step.getHttpProxyUrl().sendKeys(PARAMS.DEFAULT_PROXY_HTTP_URL);
        step.getIsSameAsHttp().click();
    }

    getFlowTestingDescription() {
        return "AWS flow (new VPC)"
    }

    executeVpcStep() {
        describe("VPC for AWS step", () => {
            const vpc = new Vpc();

            it('should have moved to VPC for AWS step', () => {
                expect(vpc.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                vpc.getNextButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
            })
        });
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
