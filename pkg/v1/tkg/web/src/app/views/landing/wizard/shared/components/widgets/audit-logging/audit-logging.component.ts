import { FormControl, FormGroup } from '@angular/forms';
import { Component, Input, OnInit } from '@angular/core';
import { FormUtils } from '../../../utils/form-utils';

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
        FormUtils.addControl(
            this.formGroup,
            'enableAuditLogging',
            new FormControl(false, [])
        );
    }

}
