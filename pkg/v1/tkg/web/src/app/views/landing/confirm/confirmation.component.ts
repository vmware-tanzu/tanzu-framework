// Angular imports
import { Component, Input, OnInit } from '@angular/core';
// App imports
import AppServices from '../../../shared/service/appServices';
import { BasicSubscriber } from '../../../shared/abstracts/basic-subscriber';
import { TkgEventType } from '../../../shared/service/Messenger';
import { UserDataWizard } from '../../../shared/service/user-data.service';
import { EditionData } from '../../../shared/service/branding.service';

@Component({
    selector: 'app-confirmation',
    templateUrl: './confirmation.component.html',
    styleUrls: ['./confirmation.component.scss']
})
export class ConfirmationComponent extends BasicSubscriber implements OnInit {
    @Input() wizard: string;

    pageTitle: string = '';
    wizardEntry: UserDataWizard;

    ngOnInit() {
        AppServices.messenger.subscribe<EditionData>(TkgEventType.BRANDING_CHANGED, data => {
                this.pageTitle = data.payload.branding.title;
            }, this.unsubscribe);
        this.wizardEntry = AppServices.userDataService.retrieveWizardEntry(this.wizard);
    }
}
