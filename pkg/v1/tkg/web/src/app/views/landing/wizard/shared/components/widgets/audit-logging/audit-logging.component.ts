import { FormControl, FormGroup } from '@angular/forms';
import { Component, Input, OnInit } from '@angular/core';

@Component({
    selector: 'app-audit-logging',
    templateUrl: './audit-logging.component.html',
    styleUrls: ['./audit-logging.component.scss']
})
export class AuditLoggingComponent implements OnInit {

    @Input() formName: string;
    @Input() formGroup: FormGroup;

    constructor() { }

    ngOnInit(): void {
        this.formGroup.addControl(
            'enableAuditLogging',
            new FormControl(false, [])
        );
    }

}
