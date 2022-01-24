// The mission of this class is to cycle through a wizard's stepData array
// and create a step component for each of the data elements
import { AfterViewInit, Component, Input, OnInit, ViewChild } from '@angular/core';
import { WizardBaseDirective, WizardStepRegistrar } from '../wizard-base/wizard-base';
import { ClrStepper } from '@clr/angular';

@Component({
    selector: 'app-step-wrapper-set',
    templateUrl: './step-wrapper-set.component.html',
})
export class StepWrapperSetComponent implements AfterViewInit {
    @Input() wizard: WizardBaseDirective;    // the wizard that holds the steps we create

    @ViewChild(ClrStepper)
    clarityStepper: ClrStepper;

    ngAfterViewInit() {
        this.validateClarityWizard();
        // work around an issue within StepperModel
        this.clarityStepper['stepperService']['accordion']['openFirstPanel'] = function () {
            const firstPanel = this.getFirstPanel();
            if (firstPanel) {
                this._panels[firstPanel.id].open = true;
                this._panels[firstPanel.id].disabled = true;
            }
        }
    }

    restartWizard() {
        this.validateClarityWizard();
        this.clarityStepper['stepperService'].resetPanels();
        this.clarityStepper['stepperService']['accordion'].openFirstPanel();
    }

    private validateClarityWizard() {
        if (!this.clarityStepper) {
            console.error('StepWrapperSetComponent does not have correct ClrStepper injected. Was the HTML changed but not the component?');
        } else if (!this.clarityStepper['stepperService']) {
            console.error('StepWrapperSetComponent\'s ClrStepper does not have a stepperService. Has the Clarity implementation changed?)');
        } else if (!this.clarityStepper['stepperService']['accordion']) {
            console.error('StepWrapperSetComponent.ClrStepper.stepperService.accordion is null. Has the Clarity implementation changed?)');
        }
    }
}
