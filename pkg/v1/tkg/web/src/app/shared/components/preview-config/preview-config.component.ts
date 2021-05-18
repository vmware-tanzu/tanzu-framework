import { Messenger, TkgEventType } from '../../service/Messenger';
import { Component, OnInit } from '@angular/core';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { takeUntil } from 'rxjs/operators';

@Component({
    selector: 'app-preview-config',
    templateUrl: './preview-config.component.html',
    styleUrls: ['./preview-config.component.scss']
})
export class PreviewConfigComponent extends BasicSubscriber implements OnInit {

    cli = "";

    constructor(private messenger: Messenger) {
        super();
        messenger.getSubject(TkgEventType.CLI_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(event => {
                this.cli = event.payload;
            });
    }

    ngOnInit() {
    }
}
