// Angular imports
import { Component, OnInit } from '@angular/core';
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
        const step = this;
        AppServices.messenger.subscribe<string>(TanzuEventType.CLI_CHANGED, event => { step.cli = event.payload; }, this.unsubscribe);
    }

    ngOnInit() {
    }
}
