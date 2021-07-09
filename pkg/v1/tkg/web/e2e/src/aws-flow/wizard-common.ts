import { NodeSettings } from './node-settings.po';
import { Provider } from './provider.po';
import { browser } from 'protractor';
import { Wizard } from './wizard.po';
import { Network } from './network.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import { AppPage } from '../app.po';
import { OsImage } from '../common/osimage.po';
import { Metadata } from '../common/metadata.po'
import { Identity } from '../common/identity.po'
import { NetworkProxy } from '../common/networkproxy.po'
import { NodeOpt } from '../common/node-setting-opt.po';

const title = 'Deploying Tanzu Community Edition on AWS';

export default abstract class WizardCommon {

    abstract executeVpcStep();
    abstract setCredentials(step: Provider);
    abstract selectSubnets(step: NodeSettings);
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

            it('should navigate to AWS flow', () => {
                page.getDeployOnAws().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
                expect(browser.getCurrentUrl()).toEqual(`http://${PARAMS.SERVER_URL}/#/ui/aws/wizard`);
            })

            const flow = new Wizard();
            flow.navigateTo();

            it('should display "Deploy Management Cluster on Amazon EC2"', () => {
                expect(flow.getTitleText()).toEqual('Deploy Management Cluster on Amazon EC2');
            });

            describe("provider step", () => {
                const provider = new Provider();

                it('should navigate to provider step', () => {
                    expect(provider.hasMovedToStep()).toBeTruthy();
                })

                it('"CONNECT" button should be enabled', () => {
                    this.setCredentials(provider);
                    provider.selectOptionByText(provider.getRegion(), PARAMS.AWS_REGION);
                    expect(provider.getConectButton().isEnabled()).toBeTruthy();
                })

                it('"CONNECT" button should display "CONNECTED"', () => {
                    provider.getConectButton().click();
                    browser.waitForAngular();
                    expect(provider.getConectButton().isEnabled()).toBeFalsy();
                    expect(provider.getConectButton().getText()).toEqual('CONNECTED');
                })

                it('Capture all user inputs', () => {
                    expect(true).toBeTruthy();
                })

                afterAll(() => {
                    provider.getNextButton().click();
                    browser.sleep(SLEEP_TIME_AFTER_NEXT);
                })
            });

            this.executeVpcStep();

            describe("Control Plane Settings step", () => {
                const nodeSettings = new NodeSettings();
                const nodeOpt = new NodeOpt();

                it('should have moved to this step', () => {
                    expect(nodeSettings.hasMovedToStep()).toBeTruthy();
                })

                if (PARAMS.CONTROL_PLANE_TYPE === 'dev') {
                    it('should display "Development cluster selected: 1 node control plane"', () => {
                        nodeSettings.selectDatalistByText("devInstanceType", PARAMS.AWS_MC_TYPE);
                        expect(nodeSettings.getTitleText()).toEqual('Development cluster selected: 1 node control plane');
                    })

                    it('should be able to select instance type and availability zone', () => {
                        nodeSettings.getMCName().clear();
                        nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
                        nodeSettings.selectDatalistByText("workerNodeInstanceType", PARAMS.AWS_WC_TYPE);
                        nodeSettings.getSshKeyName().clear();
                        nodeSettings.getSshKeyName().sendKeys(PARAMS.AWS_SSH_KEY_NAME);
                        nodeSettings.selectOptionByText(nodeSettings.getAvailabilityZone1(), PARAMS.AWS_AZ1);
                        this.selectSubnets(nodeSettings);
                        expect(true).toBeTruthy();
                    })
                } else {
                    it('should display "Production cluster selected: 3 node control plane"', () => {
                        nodeSettings.selectDatalistByText("prodInstanceType", PARAMS.AWS_MC_TYPE);
                        expect(nodeSettings.getTitleText()).toEqual('Production cluster selected: 3 node control plane');
                    })

                    it('should be able to select instance type and availability zone', () => {
                        nodeSettings.getMCName().clear();
                        nodeSettings.getMCName().sendKeys(PARAMS.MC_NAME);
                        nodeSettings.selectDatalistByText("workerNodeInstanceType", PARAMS.AWS_WC_TYPE);
                        nodeSettings.getSshKeyName().clear();
                        nodeSettings.getSshKeyName().sendKeys(PARAMS.AWS_SSH_KEY_NAME);
                        nodeOpt.getEnableAudit().click();
                        nodeSettings.selectOptionByText(nodeSettings.getAvailabilityZone1(), PARAMS.AWS_AZ1);
                        nodeSettings.selectOptionByText(nodeSettings.getAvailabilityZone2(), PARAMS.AWS_AZ2);
                        nodeSettings.selectOptionByText(nodeSettings.getAvailabilityZone3(), PARAMS.AWS_AZ3);
                        this.selectSubnets(nodeSettings);
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

                it('Capture all user inputs', () => {
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
