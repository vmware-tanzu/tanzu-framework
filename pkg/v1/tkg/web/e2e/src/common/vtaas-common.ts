import { protractor, browser, ElementArrayFinder, ElementFinder, WebElement } from 'protractor';
import { Stepper } from './step-expand.po';
import { SLEEP_TIME_AFTER_NEXT, PARAMS } from '../wizard-base.po';

const accessibility = require('vtaas-accessibility').accessibilityProtractor;

console.log(JSON.stringify(PARAMS, null, 4));

export class VtaasCommon {

    stepper: Stepper;
    step_list: ElementFinder[];
    testname: string;
    constructor(stepper, step_list, testname) {
        this.stepper = stepper;
        this.step_list = step_list;
        this.testname = testname;
    }

    executeVtaas() {
        describe("Vtaas check step", () => {
            beforeAll(async () => {
                await accessibility.begin(browser, {
                    user: PARAMS.VTAAS_USER,
                    product: 'Tanzu Kubernetes Grid (TKG)',
                    testName: this.testname
                });
            });

            it('review button should be clickable', () => {
                expect(this.stepper.hasMovedToStep()).toBeTruthy();
            })

            it('expand all steps', () => {
                for (let i = 0; i < this.step_list.length; i++) {
                    const cur_step = this.step_list[i];
                    browser.wait(this.stepper.isClickable(cur_step), 5000).then(() => {
                        cur_step.click();
                    });
                    browser.sleep(200);
                }
                expect(true).toBeTruthy();
                accessibility.check();
            });

            afterAll(() => {
                accessibility.end((result) => {
                    console.log(result);
                    expect(result.result_summary.failed).toBe(0);
                });
            })
        });
    }

}
