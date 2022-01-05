/**
 * Angular Modules
 */
import AppServices from '../../../../../shared/service/appServices';
import { Component, Input, OnInit } from '@angular/core';

@Component({
    selector: 'app-shared-delete-data-popup',
    templateUrl: './delete-data-popup.component.html',
    styleUrls: ['./delete-data-popup.component.scss']
})
export class DeleteDataPopupComponent implements OnInit {
    open: boolean;
    @Input() wizard: string;

    ngOnInit() {
        this.open = AppServices.userDataService.isWizardDataOld(this.wizard);
    }

    clearDataClick() {
        AppServices.userDataService.deleteWizardData(this.wizard);
        this.open = false;
    }

    useSavedDataClick() {
        AppServices.userDataService.updateWizardTimestamp(this.wizard);
        this.open = false;
    }
}
