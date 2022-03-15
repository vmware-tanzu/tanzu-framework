import { Provider } from './provider.po';
import { NodeSettings } from './node-settings.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { Vnet } from './vnet.po';
import { browser } from 'protractor';
import { NetworkProxy } from '../common/networkproxy.po';
import { Identity } from '../common/identity.po'

export class ExistingVnetCommon extends WizardCommon {
    getFlowTestingDescription() {
        return "Azure flow (existing VNet)"
    }

    setProvideResourceGroup(step: Provider) {
        step.selectExistingResourceGroup().click();
        browser.waitForAngular();
        step.selectOptionByText(step.selectExistingResourceGroup(), PARAMS.AZURE_RESOURCE_GROUP)
    }

    setNetworkProxy(step: NetworkProxy) {
        // Do nothing as it is not enabled
    }

    executeVnetStep() {
        describe("VNet for Azure step", () => {
            const vnet = new Vnet();

            it('should have moved to VNet for Azure step', () => {
                expect(vnet.hasMovedToStep()).toBeTruthy();
            })

            it('should be able to select an existing VNet', () => {
                vnet.getSelectAnExistingVnet().click();
                browser.waitForAngular();
                vnet.selectOptionByText(vnet.getResourceGroup(), PARAMS.AZURE_RESOURCE_GROUP);
                vnet.selectOptionByText(vnet.getVnetNameExisting(), PARAMS.AZURE_VNET);
                browser.waitForAngular();
            });

            it('Capture all user inputs', () => {
                vnet.selectOptionByText(vnet.getControlPlaneSubnet(), PARAMS.AZURE_MASTER_SUBNET);
                vnet.selectOptionByText(vnet.getWorkerNodeSubnet(), PARAMS.AZURE_WORKER_SUBNET);
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                vnet.getNextButton().click();
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
