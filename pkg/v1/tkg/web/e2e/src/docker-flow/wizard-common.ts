import { Provider } from './provider.po';
import { browser } from 'protractor';
import { Wizard } from './wizard.po';
import { Network } from './network.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import { AppPage } from '../app.po';
import { Identity } from '../common/identity.po'
import { NetworkProxy } from '../common/networkproxy.po'

export default abstract class WizardCommon {

    abstract setNetworkProxy(step: NetworkProxy);
    abstract executeIdentityStep();
    abstract getFlowTestingDescription();

    executeAll(isVtaas) {
        describe(this.getFlowTestingDescription(), () => {
            const page = new AppPage();

            it('should display welcome message', () => {
                page.navigateTo();
                expect(page.getTitleText()).toEqual('Welcome to the VMware Tanzu Kubernetes Grid Installer');
            });

            it('should navigate to Docker flow', () => {
                page.getDeployOnDocker().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
                expect(browser.getCurrentUrl()).toEqual(`http://${PARAMS.SERVER_URL}/#/ui/docker/wizard`);
            })

            const flow = new Wizard();
            flow.navigateTo();

            it('should display "Deploy Management Cluster on Docker"', () => {
                expect(flow.getTitleText()).toEqual('Deploy Management Cluster on Docker');
            });

            describe("provider step", () => {
                const provider = new Provider();

                it('should display "Validate the local Docker daemon"', () => {
                    expect(provider.getTitleText()).toEqual('Validate the local Docker daemon');
                })

                it('Capture all user inputs', () => {
                    expect(true).toBeTruthy();
                })

                afterAll(() => {
                    provider.getNextButton().click();
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

            if (isVtaas === false) {
                flow.executeDeployFlow();
            }
        });
    }

}
