import { NsxSettings } from './nsx-settings.po';
import { Network } from './network.po';
import { TmcRegister } from '../common/tmc-settings.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { Identity } from '../common/identity.po'
import { browser } from 'protractor';
import { NetworkProxy } from '../common/networkproxy.po';
import { NodeSettings } from './node-settings.po';

export class DisableOptionsCommon extends WizardCommon {

    selectEndpointProvider(nodeSettings: NodeSettings) {
        // Do nothing as it is not enabled
    }

    getFlowTestingDescription() {
        return "Vsphere flow (disable Options)"
    }

    setNetworkProxy(step: NetworkProxy) {
        // Do nothing as it is not enabled
    }

    executeNsxStep() {
        describe("NSX for Vsphere step", () => {
            const nsxSettings = new NsxSettings();

            it('should have moved to NSX Settings for Vsphere step', () => {
                expect(nsxSettings.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                nsxSettings.getNextButton().click();
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
