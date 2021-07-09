import { Provider } from './provider.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { browser } from 'protractor';
import { NetworkProxy } from '../common/networkproxy.po';
import { Identity } from '../common/identity.po'

export class DockerOidcCommon extends WizardCommon {

    setNetworkProxy(step: NetworkProxy) {
        // Do nothing as it is not enabled
    }

    getFlowTestingDescription() {
        return "Docker flow (Oidc)"
    }

    executeIdentityStep() {
        describe("Identity step", () => {
            const identity = new Identity();

            it('should have moved to this step', () => {
                expect(identity.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                identity.getIssuerURLInput().sendKeys("https://my-url.com");
                identity.getClientIdInput().sendKeys("some-client-id");
                identity.getClientSecretInput().sendKeys("some-client-secret");
                identity.getScopesInput().sendKeys("openid, offline_access");
                identity.getOidcUsernameClaim().sendKeys("some-username");
                identity.getOidcGroupsClaim().sendKeys("some-group");
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                identity.getNextButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
            })
        });
    }

}
