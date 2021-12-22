// Angular modules
import { Component, OnInit } from '@angular/core';

// App imports
import { FormMetaDataStore } from '../FormMetaDataStore';

@Component({
    selector: 'app-shared-delete-data-popup',
    templateUrl: './delete-data-popup.component.html',
    styleUrls: ['./delete-data-popup.component.scss']
})
export class DeleteDataPopupComponent implements OnInit {
    open: boolean;

    constructor() {}

    ngOnInit() {
        if (FormMetaDataStore.shouldPromptClearLocalStorage()) {
            this.open = true;
        } else {
            this.open = false;
        }
    }

    clearDataClick() {
        FormMetaDataStore.deleteAllSavedData();
        this.open = false;
    }

    useSavedDataClick() {
        FormMetaDataStore.updateLastSavedTimestamp();
        this.open = false;
    }
}
