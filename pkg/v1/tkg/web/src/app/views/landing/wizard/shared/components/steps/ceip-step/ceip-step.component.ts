// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import Broker from '../../../../../../../shared/service/broker';
import { CeipStepMapping } from './ceip-step.fieldmapping';
import { StepFormDirective } from '../../../step-form/step-form';

@Component({
    selector: 'app-shared-ceip-step',
    templateUrl: './ceip-step.component.html',
    styleUrls: ['./ceip-step.component.scss']
})
export class SharedCeipStepComponent extends StepFormDirective implements OnInit {

    ngOnInit() {
        super.ngOnInit();
        Broker.fieldMapUtilities.buildForm(this.formGroup, this.formName, CeipStepMapping);
        this.htmlFieldLabels = Broker.fieldMapUtilities.getFieldLabelMap(CeipStepMapping);
        this.storeDefaultLabels(CeipStepMapping);

        this.initFormWithSavedData();
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(CeipStepMapping);
        this.storeDefaultDisplayOrder(CeipStepMapping);
    }
}
