// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import AppServices from '../../../../../../../shared/service/appServices';
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
        AppServices.userDataFormService.buildForm(this.formGroup, this.wizardName, this.formName, CeipStepMapping);
        this.htmlFieldLabels = AppServices.fieldMapUtilities.getFieldLabelMap(CeipStepMapping);
        this.storeDefaultLabels(CeipStepMapping);
        this.registerDefaultFileImportedHandler(this.eventFileImported, CeipStepMapping);
        this.registerDefaultFileImportErrorHandler(this.eventFileImportError);
    }

    protected storeUserData() {
        this.storeUserDataFromMapping(CeipStepMapping);
        this.storeDefaultDisplayOrder(CeipStepMapping);
    }
}
