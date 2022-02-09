// Angular imports
import { Component, OnInit } from '@angular/core';
// Third party imports
import { takeUntil } from 'rxjs/operators';
// App imports
import AppServices from '../../service/appServices';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { TanzuEventType } from '../../service/Messenger';

@Component({
    selector: 'app-preview-config',
    templateUrl: './preview-config.component.html',
    styleUrls: ['./preview-config.component.scss']
})
export class PreviewConfigComponent extends BasicSubscriber implements OnInit {

    cli = "";

    constructor() {
        super();
        AppServices.messenger.getSubject(TanzuEventType.CLI_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.cli = event.payload;
            });
    }

    ngOnInit() {
    }
}
