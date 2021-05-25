import { NodeSettings } from './node-settings.po';
import { Provider } from './provider.po';
import { browser } from 'protractor';
import { Wizard } from './wizard.po';
import { OsImage } from '../common/osimage.po';
import { Network } from './network.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import { Resources } from './resources.po';
import { AppPage } from '../app.po';
import { Metadata } from '../common/metadata.po'
import { NetworkProxy } from '../common/networkproxy.po'
import { Identity } from '../common/identity.po'
import { NodeOpt } from '../common/node-setting-opt.po';

console.log(JSON.stringify(PARAMS, null, 4));
const title = 'Deploying Tanzu Community Edition on vSphere';

export default abstract class WizardCommon {

    abstract selectEndpointProvider(nodeSettings);
    abstract executeNsxStep();
    abstract setNetworkProxy(step: NetworkProxy);
    abstract executeIdentityStep();
    abstract getFlowTestingDescription();

    executeAll(isVtaas) {
        describe(this.getFlowTestingDescription(), () => {

            const page = new AppPage();

            it('should display welcome message', () => {
                page.navigateTo();
                expect(page.matchTitleText()).toBeTruthy();
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
                    console.log(`${PARAMS.VCIP}:${PARAMS.DEFAULT_VC_USER}:${PARAMS.DEFAULT_VC_PASSWORD}`);
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
                    provider.getSSHKey().click();
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
                const nodeOpt = new NodeOpt();

                it('should have moved to "Control Plane Settings" step', () => {
                    expect(nodeSettings.hasMovedToStep()).toBeTruthy();
                })

                if (PARAMS.CONTROL_PLANE_TYPE === 'dev') {
                    it('should display "Development cluster selected: 1 node control plane"', () => {
                        const devSelect = nodeSettings.getDevSelect();
                        nodeSettings.selectOptionByText(devSelect, PARAMS.DEFAULT_VC_MC_TYPE);
                        expect(nodeSettings.getTitleText()).toEqual('Development cluster selected: 1 node control plane');
                    })

                    it('captures all user inputs', () => {
                        nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
                        this.selectEndpointProvider(nodeSettings);
                        nodeSettings.getVirtualIpAddress().sendKeys(PARAMS.VSPHERE_ENDPOINT_IP);
                        nodeSettings.selectOptionByText(nodeSettings.getWorkerNodeType(), PARAMS.DEFAULT_VC_WC_TYPE);
                        nodeOpt.getEnableAudit().click();
                        expect(true).toBeTruthy();
                    });
                }
                else {
                    it('should display "Production cluster selected: 3 node control plane"', () => {
                        const prodSelect = nodeSettings.getProdSelect();
                        nodeSettings.selectOptionByText(prodSelect, PARAMS.DEFAULT_VC_MC_TYPE);
                        expect(nodeSettings.getTitleText()).toEqual('Production cluster selected: 3 node control plane');
                    })

                    it('captures all user inputs', () => {
                        nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
                        this.selectEndpointProvider(nodeSettings);
                        nodeSettings.getVirtualIpAddress().sendKeys(PARAMS.VSPHERE_ENDPOINT_IP);
                        nodeSettings.selectOptionByText(nodeSettings.getWorkerNodeType(), PARAMS.DEFAULT_VC_WC_TYPE);
                        nodeOpt.getEnableAudit().click();
                        expect(true).toBeTruthy();
                    });
                }

                afterAll(() => {
                    nodeSettings.getNextButton().click();
                    browser.sleep(SLEEP_TIME_AFTER_NEXT);
                })
            });

            this.executeNsxStep();

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
                    resources.getResource(PARAMS.DEFAULT_RESOURCE_POOL).click();
                    resources.selectDatalistByText("vmFolder", PARAMS.DEFAULT_VC_FOLDER);
                    resources.selectDatalistByText("datastore", PARAMS.DEFAULT_VC_DATASTORE);
                    expect(true).toBeTruthy();
                });

                afterAll(() => {
                    resources.getNextButton().click();
                    browser.sleep(SLEEP_TIME_AFTER_NEXT);
                })
            });

            describe("Kubernetes Network step", () => {
                const network = new Network();
                const networkproxy = new NetworkProxy();

                it('should have moved to "Kubernetes Network" step', () => {
                    expect(network.hasMovedToStep()).toBeTruthy();
                })

                it('captures all user inputs', () => {
                    network.selectDatalistByText("networkName", PARAMS.DEFAULT_VC_NETWORK);
                    this.setNetworkProxy(networkproxy);
                    expect(true).toBeTruthy();
                });

                afterAll(() => {
                    network.getNextButton().click();
                    browser.sleep(SLEEP_TIME_AFTER_NEXT);
                })
            });

            this.executeIdentityStep();

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
            if (isVtaas === false) {
                flow.executeDeployFlow(title);
            }
        });
    }
}
