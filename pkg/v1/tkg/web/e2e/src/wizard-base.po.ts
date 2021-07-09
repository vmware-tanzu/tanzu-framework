import { config } from 'process';
import { by, element, browser, protractor } from 'protractor';
import { Ceip } from './common/ceip.po';
import { TmcRegister } from './common/tmc-settings.po';

export const SLEEP_TIME_AFTER_NEXT = 10000; // ms
export const DEPLOY_TIMEOUT = 60; // minutes
export const PARAMS = browser.params;

export abstract class WizardBase {

    getReviewConfigurationButton() {
        return element(by.buttonText("REVIEW CONFIGURATION"));
    }

    getConfirmSettingsText() {
        return element(by.css('tkg-kickstart-ui-confirm h2')).getText() as Promise<string>;
    }

    displayIaaSProvider() {
        return element(by.cssContainingText('th', 'IaaS Provider')).isPresent();
    }

    getDeployButton() {
        return element(by.buttonText("Deploy Management Cluster"));
    }

    getEditButton() {
        return element(by.buttonText("Edit Configuration"));
    }

    displayDeploymentPage() {
        return element(by.cssContainingText('h2', 'Deploying Tanzu Kubernetes Grid')).isPresent();
    }

    matchDepoymentPage(title) {
        const EC = protractor.ExpectedConditions;
        return EC.or(
            EC.presenceOf(element(by.cssContainingText('h2', 'Deploying Tanzu Kubernetes Grid'))),
            EC.presenceOf(element(by.cssContainingText('h2', title)))
        );
    }

    matchConfirSettingsText() {
        const EC = protractor.ExpectedConditions;
        return EC.or(
            EC.presenceOf(element(by.cssContainingText('h2', 'Tanzu Kubernetes Grid - Confirm Settings'))),
            EC.presenceOf(element(by.cssContainingText('h2', 'Tanzu Community Edition - Confirm Settings')))
        );
    }

    untilSucceedOrFail() {
        const expectedCondition = protractor.ExpectedConditions;
        return expectedCondition.or(
            expectedCondition.presenceOf(element(by.cssContainingText('span', 'Installation complete'))),
            expectedCondition.presenceOf(element(by.cssContainingText('div', 'has failed')))
        );
    }

    isDeploymentSuccessful() {
        return element(by.cssContainingText('span', 'Installation complete, you can now close the browser....')).isPresent();
    }

    executeCommonFlow() {

        describe("Register TMC step", () => {
            const tmc = new TmcRegister();

            it('should have moved to this step', () => {
                expect(tmc.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                tmc.getNextButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
            })
        });

        describe("CEIP Agreement step", () => {
            const ceip = new Ceip();

            it('should have moved to this step', () => {
                expect(ceip.hasMovedToStep()).toBeTruthy();
            })

            it('Capture all user inputs', () => {
                expect(ceip.getCeipCheckbox().getAttribute("value")).toBeTruthy()
                expect(true).toBeTruthy();
            });

            afterAll(() => {
                ceip.getNextButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
            })
        });
    }

    executeDeployFlow(title) {
        describe("Review Configuration", () => {

            it('"Review Configuration" button should be enabled', () => {
                expect(this.getReviewConfigurationButton().isEnabled).toBeTruthy();
            })

            it('should display "Tanzu Kubernetes Grid - Confirm Settings"', () => {
                this.getReviewConfigurationButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
                browser.wait(this.matchConfirSettingsText(), 5000);
            })

            it('should navigate to deployment status page', () => {
                this.getDeployButton().click();
                browser.sleep(SLEEP_TIME_AFTER_NEXT);
                expect(this.matchDepoymentPage(title)).toBeTruthy();
            })

            it('should finish deployment', async () => {
                try {
                    await browser.wait(this.untilSucceedOrFail(),
                        DEPLOY_TIMEOUT * 60 * 1000, `Deployment timeout of ${DEPLOY_TIMEOUT} minutes has reached.`);
                } catch (e) {
                    expect(false).toBeTruthy();
                }
            });
        });
    }

}
