// Angular imports
import { Component, OnInit, Input } from '@angular/core';
import { Router } from '@angular/router';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';

// App imports
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import { FormMetaDataStore, FormMetaData, StepMetaData } from './../wizard/shared/FormMetaDataStore';

@Component({
    selector: 'tkg-kickstart-ui-confirm',
    templateUrl: './confirm.component.html',
    styleUrls: ['./confirm.component.scss']
})
export class ConfirmComponent implements OnInit {
    @Input() errorNotification: any;

    APP_ROUTES: Routes = APP_ROUTES;
    reviewData;

    steps: string[];
    stepMetaDataList: StepMetaData[];
    formMetaDataList: any[];

    constructor(
        private wizardFormService: VSphereWizardFormService,
        private router: Router) {
    }

    ngOnInit() {
        this.stepMetaDataList = FormMetaDataStore.getStepList();
        this.steps = FormMetaDataStore.getFormList();
        this.formMetaDataList = this.steps.map(formName => FormMetaDataStore.getMetaData(formName));
    }

    /**
     * Get all the entries of the 'data' object.
     * @param data the data whose entries to return
     */
    entries(data: Object) {
        if (data) {
            return Object.values(data);
        }
        return null;
    }
}
