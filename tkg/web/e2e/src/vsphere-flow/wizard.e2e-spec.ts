import { NodeSettings } from './node-settings.po';
import { Provider } from './provider.po';
import { browser } from 'protractor';
import { Wizard } from './wizard.po';
import { OsImage } from './osimage.po';
import { Network } from './network.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import { Resources } from './resources.po';
import { AppPage } from '../app.po';
import { Metadata } from '../common/metadata.po'
import { Identity } from '../common/identity.po'

console.log(JSON.stringify(PARAMS, null, 4));

describe('vSphere flow', () => {

    const page = new AppPage();

    it('should display welcome message', () => {
        page.navigateTo();
        expect(page.getTitleText()).toEqual('Welcome to the VMware Tanzu Kubernetes Grid Installer');
    });

    it('should navigate to vSphere flow', () => {
        page.getDeployOnVsphere().click();
        browser.sleep(SLEEP_TIME_AFTER_NEXT);
        expect(browser.getCurrentUrl()).toEqual(`http://${PARAMS.SERVER_URL}/#/ui/wizard`);
    })

    const flow = new Wizard();
    it('should display "Deploy Management Cluster on vSphere"', () => {
        expect(flow.getTitleText()).toEqual('Deploy Management Cluster on vSphere');
    });

    describe("provider step", () => {
        const provider = new Provider();

        it('should navigate to provider step', () => {
            expect(provider.hasMovedToStep()).toBeTruthy();
        })

        it('"CONNECT" button should be enabled', () => {
            provider.getVC().sendKeys(PARAMS.VCIP);
            provider.getUsername().sendKeys(PARAMS.DEFAULT_VC_USER);
            provider.getPassword().sendKeys(PARAMS.DEFAULT_VC_PASSWORD + " ");
            expect(provider.getConectButton().isEnabled()).toBeTruthy();
        })

        it('should NOT show "CONNECTED" with wrong password', () => {
            provider.getConectButton().click();
            provider.getThumbprintContinueButton().click();
            browser.waitForAngular();
            expect(provider.getConectButton().isEnabled()).toBeTruthy();
            expect(provider.getConectButton().getText()).not.toEqual('CONNECTED');
        })

        it('"CONNECT" button should display "CONNECTED"', () => {
            provider.getPassword().clear();
            provider.getPassword().sendKeys(PARAMS.DEFAULT_VC_PASSWORD);
            provider.getConectButton().click();
            provider.getThumbprintContinueButton().click();
            browser.waitForAngular();
            expect(provider.getConectButton().isEnabled()).toBeFalsy();
            expect(provider.getConectButton().getText()).toEqual('CONNECTED');
        })

        it('captures all user inputs', () => {
            provider.selectOptionByIndex(provider.getDC(), 2);
            provider.getSSHKey().sendKeys(PARAMS.DEFAULT_SSH_KEY);
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            provider.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    describe("Management Cluster Settings step", () => {
        const nodeSettings = new NodeSettings();

        it('should have moved to "Control Plane Settings" step', () => {
            expect(nodeSettings.hasMovedToStep()).toBeTruthy();
        })

        it('should display "Development cluster selected: 1 node control plane"', () => {
            const devSelect = nodeSettings.getDevSelect();
            nodeSettings.selectOptionByIndex(devSelect, 2);
            expect(nodeSettings.getTitleText()).toEqual('Development cluster selected: 1 node control plane');
        })

        it('captures all user inputs', () => {
            nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
            nodeSettings.getVirtualIpAddress().sendKeys("10.10.10.10")
            nodeSettings.selectOptionByIndex(nodeSettings.getWorkerNodeType(), 2);
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            nodeSettings.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    describe("Metadata step", () => {
        const metadata = new Metadata();

        it('should have moved to this step', () => {
            expect(metadata.hasMovedToStep()).toBeTruthy();
        })

        it('Capture all user inputs', () => {
            metadata.getMCDescription().sendKeys("some-description");
            metadata.getMCLocation().sendKeys("some-location");
            metadata.getMCLabelsKey().sendKeys("somekey");
            metadata.getMCLabelsValue().sendKeys("someval");
            metadata.getMCLabelsAddButton().click();
            metadata.getMCLabelsKey().sendKeys("delete-this-key");
            metadata.getMCLabelsValue().sendKeys("delete-this-value");
            metadata.getMCLabelsAddButton().click();
            metadata.getMCLabelsDeleteButton("delete-this-key").click();
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            metadata.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    describe("Resources step", () => {
        const resources = new Resources();

        it('should have moved to "Resources" step', () => {
            expect(resources.hasMovedToStep()).toBeTruthy();
        })

        it('"refresh" button should work', () => {
            resources.getRefreshButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
            expect(resources.getResourcePoolSize()).toBeGreaterThan(0);
        })

        it('captures all user inputs', () => {
            resources.getResources().get(1).click();     // Check the last option
            resources.selectDatalistByIndex("vmFolder", 2);
            resources.selectDatalistByIndex("datastore", 2);
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            resources.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    describe("Kubenetes Network step", () => {
        const network = new Network();

        it('should have moved to "Kubenetes Network" step', () => {
            expect(network.hasMovedToStep()).toBeTruthy();
        })

        it('captures all user inputs', () => {
            network.selectDatalistByIndex("networkName", 2);
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            network.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    describe("Identity step", () => {
        const identity = new Identity();

        it('should have moved to this step', () => {
            expect(identity.hasMovedToStep()).toBeTruthy();
        })

        it('Capture all user inputs', () => {
            identity.getPorviderNameInput().sendKeys("some-provider");
            identity.getIssuerURLInput().sendKeys("my-url.com");
            identity.getClientIdInput().sendKeys("some-client-id");
            identity.getClientSecretInput().sendKeys("some-client-secret");
            identity.getScopesInput().sendKeys("openid, offline_access");
            identity.getSkipVerifyCertificateCheckbox().click();
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            identity.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    describe("OS Image step", () => {
        const osImage = new OsImage();

        it('should have moved to "OS Image" step', () => {
            expect(osImage.hasMovedToStep()).toBeTruthy();
        })

        it('OS image refresh should work', () => {
            osImage.getRefreshButton().click();
            browser.waitForAngular();
            expect(osImage.getOsImageCount()).toBeGreaterThan(1);
            osImage.selectOptionByIndex(osImage.getOsImages(), 2);
        })

        it('captures all user inputs', () => {
            expect(true).toBeTruthy();
        });

        afterAll(() => {
            osImage.getNextButton().click();
            browser.sleep(SLEEP_TIME_AFTER_NEXT);
        })
    });

    flow.executeCommonFlow();
});
