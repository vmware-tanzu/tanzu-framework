import { Provider } from './provider.po';
import { NodeSettings } from './node-settings.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';
import WizardCommon from "./wizard-common";
import { Vpc } from './vpc.po';
import { browser } from 'protractor';
import { NetworkProxy } from '../common/networkproxy.po';
import { Identity } from '../common/identity.po'

export class ExistingVpcCommon extends WizardCommon {
    setCredentials(step: Provider) {
        step.getStaticCredential().click();
        step.getAccessKeyId().clear();
        step.getAccessKeyId().sendKeys(PARAMS.AWS_ACCESS_KEY_ID);
        step.getSecretAccessKey().clear();
        step.getSecretAccessKey().sendKeys(PARAMS.AWS_SECRET_ACCESS_KEY);
    }

    selectSubnets(step: NodeSettings) {
        step.selectOptionByIndex(step.getVpcPublicSubset(), 2);
        step.selectOptionByIndex(step.getVpcPrivateSubset(), 2);
    }

    setNetworkProxy(step: NetworkProxy) {
        // Do nothing as it is not enabled
    }

    getFlowTestingDescription() {
        return "AWS flow (existing VPC)"
    }

    executeProviderStep() {
        describe("provider step", () => {
            const provider = new Provider();

            it('should display "Validate the AWS provider credentials for Tanzu Kubernetes Grid"', () => {
                expect(provider.getTitleText()).toEqual('Validate the AWS provider credentials for Tanzu Kubernetes Grid');
            })

            it('"CONNECT" button should be enabled', () => {
                provider.getAccessKeyId().clear();
                provider.getAccessKeyId().sendKeys(PARAMS.AWS_ACCESS_KEY_ID);
                provider.getSecretAccessKey().clear();
                provider.getSecretAccessKey().sendKeys(PARAMS.AWS_SECRET_ACCESS_KEY);
                provider.getSshKeyName().clear();
                provider.getSshKeyName().sendKeys(PARAMS.AWS_SSH_KEY_NAME);
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
    }

    executeVpcStep() {
        describe("VPC for AWS step", () => {
            const vpc = new Vpc();

            it('should have moved to VPC for AWS step', () => {
                expect(vpc.hasMovedToStep()).toBeTruthy();
            })

            it('should be able to select an existing VPC', () => {
                vpc.getSelectAnExistingVpc().click();
                browser.waitForAngular();
                vpc.selectOptionByText(vpc.getVpcId(), PARAMS.AWS_EXISTING_VPC_ID);
                browser.waitForAngular();
                expect(vpc.getVpcCidr().value).not.toBeNull();
            });

            it('Capture all user inputs', () => {
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                vpc.getNextButton().click();
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
