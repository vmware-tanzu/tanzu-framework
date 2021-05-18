import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';

@Component({
    selector: 'app-ssl-thumbprint-modal',
    templateUrl: './ssl-thumbprint-modal.component.html',
    styleUrls: ['./ssl-thumbprint-modal.component.scss']
})
export class SSLThumbprintModalComponent implements OnInit {
    @Input() thumbprint: string;
    @Input() vcenterHost: string;
    @Output() verifiedThumbprint: EventEmitter<boolean> = new EventEmitter();

    show: boolean;

    constructor() {}

    ngOnInit(): void {
        this.show = false;
    }

    open() {
        this.show = true;
    }

    close() {
        this.verifiedThumbprint.emit(false);
        this.show = false;
    }

    continueBtnHandler() {
        this.verifiedThumbprint.emit(true);
        this.show = false;
    }

}
