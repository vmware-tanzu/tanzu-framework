import { NodeSettings } from './node-settings.po';
import { Provider } from './provider.po';
import { browser, Key } from 'protractor';
import { Wizard } from './wizard.po';
import { NodeOpt } from '../common/node-setting-opt.po';
import { Network } from './network.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import { AppPage } from '../app.po';
import { OsImage } from '../common/osimage.po';
import { Metadata } from '../common/metadata.po'
import { NetworkProxy } from '../common/networkproxy.po'
import { Identity } from '../common/identity.po';

const title = 'Deploying Tanzu Community Edition on Azure';

export default abstract class WizardCommon {

    abstract setProvideResourceGroup(step: Provider);
    abstract executeVnetStep();
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

            it('should navigate to Azure flow', () => {
                page.getDeployOnAzure().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
                expect(browser.getCurrentUrl()).toEqual(`http://${PARAMS.SERVER_URL}/#/ui/azure/wizard`);
            })

            const flow = new Wizard();
            flow.navigateTo();

            it('should display "Deploy Management Cluster on Azure"', () => {
                expect(flow.getTitleText()).toEqual('Deploy Management Cluster on Azure');
            });

            describe("provider step", () => {
                const provider = new Provider();

                it('should display "Validate the Azure provider credentials for Tanzu Kubernetes Grid"', () => {
                    provider.getTenantId().sendKeys(Key.chord(Key.CONTROL, "a"));
                    provider.getTenantId().sendKeys(Key.BACK_SPACE);
                    provider.getTenantId().clear();
                    provider.getClientId().clear();
                    provider.getClientSecret().clear();
                    provider.getSubscriptionId().clear();
                    expect(provider.getTitleText()).toEqual('Validate the Azure provider credentials for Tanzu Kubernetes Grid');
                })

                it('"CONNECT" button should be enabled', () => {
                    provider.getTenantId().sendKeys(PARAMS.AZURE_TENANT_ID);
                    provider.getClientId().sendKeys(PARAMS.AZURE_CLIENT_ID);
                    provider.getClientSecret().sendKeys(PARAMS.AZURE_CLIENT_SECRET);
                    provider.getSubscriptionId().sendKeys(PARAMS.AZURE_SUBSCRIPTION_ID);
                    provider.selectOptionByText(provider.getAzureCloud(), PARAMS.AZURE_CLOUD);
                    expect(provider.getConectButton().isEnabled()).toBeTruthy();
                    expect(provider.getTitleText()).toEqual('Azure tenant: ' + PARAMS.AZURE_TENANT_ID);
                })

                it('"CONNECT" button should display "CONNECTED"', () => {
                    provider.getConectButton().click();
                    browser.waitForAngular();
                    expect(provider.getConectButton().isEnabled()).toBeFalsy();
                    expect(provider.getConectButton().getText()).toEqual('CONNECTED');
                })

                it('Capture all user inputs', () => {
                    provider.selectOptionByText(provider.getRegion(), PARAMS.AZURE_REGION);
                    provider.getSshPublicKey().clear();
                    provider.getSshPublicKey().sendKeys(PARAMS.AZURE_SSH_PUBLIC_KEY_B64);
                    this.setProvideResourceGroup(provider);
                    expect(true).toBeTruthy();
                })

                afterAll(() => {
                    provider.getNextButton().click();
                    browser.sleep(SLEEP_TIME_AFTER_NEXT);
                })
            });

            this.executeVnetStep();

            describe("Control Plane Settings step", () => {
                const nodeSettings = new NodeSettings();
                const nodeOpt = new NodeOpt();

                it('should have moved to this step', () => {
                    expect(nodeSettings.hasMovedToStep()).toBeTruthy();
                })

                if (PARAMS.CONTROL_PLANE_TYPE === 'dev') {
                    it('should display "Control plane type: dev"', () => {
                        const devSelect = nodeSettings.getDevSelect();
                        nodeSettings.selectOptionByText(devSelect, PARAMS.AZURE_MC_TYPE);
                        expect(nodeSettings.getTitleText()).toEqual('Control plane type: dev');
                    })

                    it('should be able to select instance type', () => {
                        nodeSettings.getMCName().clear();
                        nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
                        nodeSettings.selectOptionByText(nodeSettings.getWorkNodeInstanceType(), PARAMS.AZURE_WC_TYPE);
                        nodeOpt.getEnableAudit().click();
                        expect(true).toBeTruthy();
                    })
                } else {
                    it('should display "Control plane type: prod"', () => {
                        const prodSelect = nodeSettings.getProdSelect();
                        nodeSettings.selectOptionByText(prodSelect, PARAMS.AZURE_MC_TYPE);
                        expect(nodeSettings.getTitleText()).toEqual('Control plane type: prod');
                    })

                    it('should be able to select instance type', () => {
                        nodeSettings.getMCName().clear();
                        nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
                        nodeSettings.selectOptionByText(nodeSettings.getWorkNodeInstanceType(), PARAMS.AZURE_WC_TYPE);
                        nodeOpt.getEnableAudit().click();
                        expect(true).toBeTruthy();
                    })
                }
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

            describe("Kubernetes Network step", () => {
                const network = new Network();
                const networkproxy = new NetworkProxy();

                it('should have moved to "Kubernetes Network" step', () => {
                    expect(network.hasMovedToStep()).toBeTruthy();
                })

                it('captures all user inputs', () => {
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
