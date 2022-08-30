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
        // TODO: change the process of using stored data so that instead of using stored data to BUILD the steps' forms,
        // we re-purpose the CONFIG_FILE_IMPORTED events to be RESTORE_FROM_STORED_DATA events and only restore data in
        // response to that event (which can be used from here as well as from file import).
        // Once that is in place, then this component can broadcast that event (a) in ngOnInit() if the data is "fresh", or
        // (b) on useSavedDataClick() when data is "old" but the user says we should apply it anyway.
        // That way we would NOT apply the data if the user tells us not to.
        // In the meantime, we are disabling this functionality because we're using the data right away, even before the user
        // has a chance to respond to the question of whether we should!
        //
        // this.open = AppServices.userDataService.isWizardDataOld(this.wizard);
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
