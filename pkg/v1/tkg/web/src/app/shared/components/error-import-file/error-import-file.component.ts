import { Component, Input, OnInit } from '@angular/core';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';

@Component({
    selector: 'app-error-import-file',
    templateUrl: './error-import-file.component.html'
})
export class ErrorImportFileComponent extends BasicSubscriber implements OnInit {
    @Input() errorImportFile: any;

    ngOnInit() {
    }
}
