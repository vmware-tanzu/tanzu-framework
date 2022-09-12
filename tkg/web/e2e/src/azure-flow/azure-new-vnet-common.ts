import { Provider } from './provider.po';
import { NodeSettings } from './node-settings.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { Vnet } from './vnet.po';
import { NetworkProxy } from '../common/networkproxy.po';
import { browser } from 'protractor';
import { Identity } from '../common/identity.po'

export class NewVnetCommon extends WizardCommon {

    getFlowTestingDescription() {
        return "Azure flow (new VNet)"
    }

    setProvideResourceGroup(step: Provider) {
        step.getCreateNewResourceGroup().click();
        browser.waitForAngular();
        step.getCustomResourceGroup().clear();
        step.getCustomResourceGroup().sendKeys("my-resource-group")
    }

    setNetworkProxy(step: NetworkProxy) {
        step.getProxyButton().click();
        browser.waitForAngular();
        step.getHttpProxyUrl().sendKeys(PARAMS.DEFAULT_PROXY_HTTP_URL);
        step.getIsSameAsHttp().click();
    }

    executeVnetStep() {
        describe("VNet for Azure step", () => {
            const vnet = new Vnet();

            it('should have moved to VNet for Azure step', () => {
                expect(vnet.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                vnet.getCreateNewVnet().click();
                vnet.selectOptionByText(vnet.getResourceGroup(), "my-resource-group");
                vnet.getVnetName().clear();
                vnet.getVnetName().sendKeys("my-vnet");
                expect(vnet.getVnetCidrText()).toEqual("10.0.0.0/16");
                vnet.getControlPlaneSubnetNew().sendKeys("control-plane-subnet");
                expect(vnet.getControlPlaneSubnetCidrNewText()).toEqual("10.0.0.0/24");
                vnet.getControlPlaneSubnetCidrNew().clear();
                vnet.getControlPlaneSubnetCidrNew().sendKeys("10.0.0.0/24");
                vnet.getWorkerNodeSubnetNew().sendKeys("worker-node-subnet");
                expect(vnet.getWorkerNodeSubnetCidrNewText()).toEqual("10.0.1.0/24");
                vnet.getWorkerNodeSubnetCidrNew().clear();
                vnet.getWorkerNodeSubnetCidrNew().sendKeys("10.0.1.0/24");
                vnet.getPrivateCluster().click();
                vnet.getPrivateIP().sendKeys(PARAMS.AZURE_PRIVATE_IP);
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
