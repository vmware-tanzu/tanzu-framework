// Angular imports
import { Component, Input, OnInit } from '@angular/core';
// Third party imports
import { takeUntil } from 'rxjs/operators';
// App imports
import AppServices from '../../../shared/service/appServices';
import { BasicSubscriber } from '../../../shared/abstracts/basic-subscriber';
import { TkgEvent, TkgEventType } from '../../../shared/service/Messenger';
import { UserDataWizard } from '../../../shared/service/user-data.service';

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
        AppServices.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                this.pageTitle = data.payload.branding.title;
            });
        this.wizardEntry = AppServices.userDataService.retrieveWizardEntry(this.wizard);
    }
}
