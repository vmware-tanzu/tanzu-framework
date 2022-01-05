/**
 * Angular Modules
 */
import { Component, Input, OnInit } from '@angular/core';
import Broker from '../../../../../shared/service/broker';

@Component({
    selector: 'app-shared-delete-data-popup',
    templateUrl: './delete-data-popup.component.html',
    styleUrls: ['./delete-data-popup.component.scss']
})
export class DeleteDataPopupComponent implements OnInit {
    open: boolean;
    @Input() wizard: string;

    ngOnInit() {
        this.open = Broker.userDataService.isWizardDataOld(this.wizard);
    }

    clearDataClick() {
        Broker.userDataService.deleteWizardData(this.wizard);
        this.open = false;
    }

    useSavedDataClick() {
        Broker.userDataService.updateWizardTimestamp(this.wizard);
        this.open = false;
    }
}
