import { Component, Input, OnInit } from '@angular/core';
import { BasicSubscriber } from '../../../shared/abstracts/basic-subscriber';
import Broker from '../../../shared/service/broker';
import { TkgEvent, TkgEventType } from '../../../shared/service/Messenger';
import { takeUntil } from 'rxjs/operators';
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
        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                this.pageTitle = data.payload.branding.title;
            });
        this.wizardEntry = Broker.userDataService.retrieveWizardEntry(this.wizard);
    }
}
