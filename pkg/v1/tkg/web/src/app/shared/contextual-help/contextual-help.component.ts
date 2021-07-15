import { Component, ElementRef, OnInit, ViewEncapsulation } from '@angular/core';
import { takeUntil } from 'rxjs/operators';
import index from '../../../contextualHelpDocs/index.json';
import { BasicSubscriber } from '../abstracts/basic-subscriber';
import Broker from '../service/broker';
import { TkgEventType } from '../service/Messenger';

declare let elasticlunr: any;

interface ContextualHelpIndex {
    htmlContent: string,
    tags: Array<string>,
    title: string
};
interface ContextualHelpOpenEvent {
    type: TkgEventType.OPEN_CONTEXTUAL_HELP,
    payload: {
        title: string,
        keywords: Array<string>
    }
};
@Component({
    selector: 'app-contextual-help',
    templateUrl: './contextual-help.component.html',
    styleUrls: ['./contextual-help.component.scss'],
    encapsulation: ViewEncapsulation.None
})
export class ContextualHelpComponent extends BasicSubscriber implements OnInit {

    visible: boolean = false;
    isTopicView: boolean = true;
    isPinned: boolean = false;
    title: string = '';
    lunrIndex: any;
    htmlContentIndexArray: Array<ContextualHelpIndex> = [];
    htmlContentIndex: ContextualHelpIndex = {
        htmlContent: '',
        title: '',
        tags: []
    };

    constructor(
        private elementRef: ElementRef
    ) {
        super();
        Broker.messenger.getSubject(TkgEventType.OPEN_CONTEXTUAL_HELP)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((event: ContextualHelpOpenEvent) => {
                this.show(event.payload.keywords);
                this.title = event.payload.title;
            });
    }

    ngOnInit(): void {
        this.lunrIndex = elasticlunr.Index.load(index);
    }

    getHTMLContent(htmlRefs) {
        this.htmlContentIndexArray = [];
        htmlRefs.forEach(ref => {
            this.htmlContentIndexArray.push(this.lunrIndex.documentStore.getDoc(ref.ref));
        });

        if (this.htmlContentIndexArray.length === 1) {
            this.isTopicView = false;
            this.htmlContentIndex = this.htmlContentIndexArray[0];
        }
    }

    show(keywords: Array<string>) {
        this.visible = true;
        this.getHTMLContent(this.lunrIndex.search(keywords, {bool: 'AND'}));
    }

    hide() {
        this.visible = false;
        this.isTopicView = true;
    }

    showContent(htmlContentIndex: ContextualHelpIndex) {
        this.isTopicView = false;
        this.htmlContentIndex = htmlContentIndex;
    }
    navigateBack() {
        this.isTopicView = true;
    }

    togglePin() {
        const prevEl: HTMLElement = this.elementRef.nativeElement.previousElementSibling;

        if (this.isPinned) {
            prevEl.style.marginRight = '';
            prevEl.style.display = '';
        } else {
            prevEl.style.marginRight = '380px';
            prevEl.style.display = 'block';
        }

        this.isPinned = !this.isPinned;
    }

}
