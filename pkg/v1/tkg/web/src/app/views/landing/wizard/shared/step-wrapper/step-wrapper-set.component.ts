// The mission of this class is to cycle through a wizard's stepData array
// and create a step component for each of the data elements
import { Component, Input, OnInit, ViewChild } from '@angular/core';
import { WizardBaseDirective, WizardStepRegistrar } from '../wizard-base/wizard-base';
import { ClrStepper } from '@clr/angular';

@Component({
    selector: 'app-step-wrapper-set',
    templateUrl: './step-wrapper-set.component.html',
})
export class StepWrapperSetComponent implements OnInit {
    @Input() wizard: WizardBaseDirective;    // the wizard that holds the steps we create

    @ViewChild('clarityWizard', { read: ClrStepper, static: true })
    clarityWizard: ClrStepper;

    ngOnInit() {
        // work around an issue within StepperModel
        this.clarityWizard['stepperService']['accordion']['openFirstPanel'] = function () {
            const firstPanel = this.getFirstPanel();
            if (firstPanel) {
                this._panels[firstPanel.id].open = true;
                this._panels[firstPanel.id].disabled = true;
            }
        }
    }
}
