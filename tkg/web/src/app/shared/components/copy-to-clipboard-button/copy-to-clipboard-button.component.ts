import {
    ElementRef,
    HostBinding,
    Output,
    EventEmitter,
    Component,
    OnInit,
    Input,
    ViewChild,
    OnChanges,
} from '@angular/core';

const SHOW_CHECKBOX_TIMEOUT = 1500;

@Component({
    selector: 'app-copy-to-clipboard-button',
    templateUrl: './copy-to-clipboard-button.component.html',
    styleUrls: ['./copy-to-clipboard-button.component.scss']
})
export class VmwCopyToClipboardButtonComponent implements OnInit, OnChanges {
    @ViewChild('area', { read: ElementRef }) area: ElementRef;

    // tslint:disable-next-line: no-input-rename
    @HostBinding('class') @Input('class') classList: string = '';

    @Input() value: string;
    @Input() ariaLabel: string = "";
    @Input() size = 16;
    @Input() tooltip = '';
    @Input() btnLabel = '';  // if no label specified, show the normal copy icon
    @Input() btnClasses = ['btn-outline'];  // if no label specified, show the normal copy icon
    @Input() disabled: boolean = false;
    @Input() tooltipDirection = 'top-left';

    @Output() copyClick = new EventEmitter<null>();

    private firstLoad = true;   // show correct icon first

    btnClassesToApply: string;
    animClasses = '';
    bounds: string;
    hasProjectedContent: boolean = false;
    isSafari: boolean = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);

    constructor(private el: ElementRef) {}

    ngOnInit() {
        this.bounds = (this.size + 6) + 'px';

        this.hasProjectedContent = this.el.nativeElement.innerText.trim();

        this.calculateClassesToApply();
    }

    calculateClassesToApply() {
        let classes: Array<string> = [];

        if (!this.btnLabel.length) {
            classes.push('icon-btn');
        }

        if (this.btnLabel.length) {
            classes = classes.concat(this.btnClasses);
        }

        if (this.disabled) {
            classes.push('disabled');
        }

        this.btnClassesToApply = classes.join(' ') + ' ' + this.classList;
    }

    ngOnChanges() {
        this.calculateClassesToApply();
    }

    copyToClipboard(val: string) {
        const myWindow: any = window;

        const onCopy = (e: ClipboardEvent) => {
            e.preventDefault();

            if (e.clipboardData) {
                e.clipboardData.setData('text/plain', val);
            } else if (myWindow.clipboardData) {
                myWindow.clipboardData.setData('Text', val);
            }

            myWindow.removeEventListener('copy', onCopy);
        };

        if (this.isSafari) {
            this.area.nativeElement.value = val;
            this.area.nativeElement.select();
        }

        myWindow.addEventListener('copy', onCopy);

        if (myWindow.clipboardData && myWindow.clipboardData.setData) {
            myWindow.clipboardData.setData('Text', val);
        } else {
            document.execCommand('copy');
        }
    }

    doCopy() {
        this.copyToClipboard(this.value);
        this.copyClick.emit();
        this.firstLoad = false;
        this.animClasses = 'flip-horizontal-bottom';
        this.ariaLabel = 'copied cli command';
        setTimeout(() => {
            this.animClasses = 'flip-horizontal-reverse';
            this.ariaLabel = '';
        }, SHOW_CHECKBOX_TIMEOUT);
    }
}
