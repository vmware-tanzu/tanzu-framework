import {Component, Input, OnInit} from '@angular/core';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';

@Component({
    selector: 'app-import-file-success-notification',
    templateUrl: './import-file-success-notification.component.html'
})
export class ImportFileSuccessNotificationComponent extends BasicSubscriber implements OnInit {
    @Input() successImportFile: any;

    ngOnInit() {
    }
}
