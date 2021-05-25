import { NsxSettings } from './nsx-settings.po';
import { NetworkProxy } from '../common/networkproxy.po';
import { TmcRegister } from '../common/tmc-settings.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { Identity } from '../common/identity.po'
import { browser } from 'protractor';
import { NodeSettings } from './node-settings.po';

export class EnableOptionsCommon extends WizardCommon {

    selectEndpointProvider(nodeSettings: NodeSettings) {
        nodeSettings.selectOptionByText(nodeSettings.getEndpointProviderSelect(), 'NSX Advanced Load Balancer');
    }

    getFlowTestingDescription() {
        return "Vsphere flow (enable Options)"
    }

    setNetworkProxy(step: NetworkProxy) {
        step.getProxyButton().click();
        browser.waitForAngular();
        step.getHttpProxyUrl().sendKeys(PARAMS.DEFAULT_PROXY_HTTP_URL);
        step.getIsSameAsHttp().click();
    }

    executeNsxStep() {
        describe("NSX for Vsphere step", () => {
            const nsxSettings = new NsxSettings();

            it('should have moved to NSX Settings for Vsphere step', () => {
                expect(nsxSettings.hasMovedToStep()).toBeTruthy();
            })

            it('"VERIFY" button should be enabled', () => {
                console.log(`${PARAMS.CONTROLLER_HOST}:${PARAMS.DEFAULT_NSX_USER}:${PARAMS.DEFAULT_NSX_PASSWORD}`);
                nsxSettings.getControllerHost().sendKeys(PARAMS.CONTROLLER_HOST);
                nsxSettings.getNsxUserName().sendKeys(PARAMS.DEFAULT_NSX_USER);
                nsxSettings.getNsxPassword().sendKeys(PARAMS.DEFAULT_NSX_PASSWORD + " ");
                expect(nsxSettings.getNsxVerifyButton().isEnabled()).toBeTruthy();
            })

            it('should NOT show "VERIFIED" with wrong password', () => {
                nsxSettings.getNsxVerifyButton().click();
                browser.waitForAngular();
                expect(nsxSettings.getNsxVerifyButton().isEnabled()).toBeTruthy();
                expect(nsxSettings.getNsxVerifyButton().getText()).not.toEqual('VERIFIED');
            })

            it('"VERIFY" button should display "VERIFIED"', () => {
                nsxSettings.getNsxPassword().clear();
                nsxSettings.getNsxPassword().sendKeys(PARAMS.DEFAULT_NSX_PASSWORD);
                nsxSettings.getControllerCert().sendKeys(PARAMS.DEFAULT_NSX_CA);
                nsxSettings.getNsxVerifyButton().click();
                browser.waitForAngular();
                expect(nsxSettings.getNsxVerifyButton().isEnabled()).toBeFalsy();
                expect(nsxSettings.getNsxVerifyButton().getText()).toEqual('VERIFIED');
            })

            it('Capture all user inputs', () => {
                nsxSettings.selectOptionByText(nsxSettings.getCloudName(), PARAMS.DEFAULT_NSX_CLOUD);
                nsxSettings.selectOptionByText(nsxSettings.getServiceEngine(), PARAMS.DEFAULT_NSX_SE);
                nsxSettings.selectOptionByText(nsxSettings.getNetworkName(), PARAMS.DEFAULT_NSX_NETWORK_NAME);
                nsxSettings.selectOptionByText(nsxSettings.getNetworkCIDR(), PARAMS.DEFAULT_NSX_NETWORK_CIDR);
                nsxSettings.getNsxLabelsKey().sendKeys("somekey");
                nsxSettings.getNsxLabelsValue().sendKeys("someval");
                nsxSettings.getNsxLabelsAddButton().click();
                nsxSettings.getNsxLabelsKey().sendKeys("delete-this-key");
                nsxSettings.getNsxLabelsValue().sendKeys("delete-this-value");
                nsxSettings.getNsxLabelsAddButton().click();
                nsxSettings.getNsxLabelsDeleteButton("delete-this-key").click();
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
