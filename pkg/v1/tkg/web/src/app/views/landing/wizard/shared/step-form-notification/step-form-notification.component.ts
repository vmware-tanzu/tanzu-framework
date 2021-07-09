import { Component, OnInit, Input } from '@angular/core';

@Component({
    selector: 'app-step-form-notification',
    templateUrl: './step-form-notification.component.html',
    styleUrls: ['./step-form-notification.component.scss']
})
export class StepFormNotificationComponent implements OnInit {
    @Input() errorNotification
    constructor() { }

    ngOnInit() {
    }

}
