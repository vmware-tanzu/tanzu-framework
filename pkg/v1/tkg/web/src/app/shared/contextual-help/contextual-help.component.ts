import { Component, Input, OnDestroy, OnInit, ViewEncapsulation } from '@angular/core';
import index from '../../../contextualHelpDocs/index.json';
import { BasicSubscriber } from '../abstracts/basic-subscriber';
import { ContexutalHelpService } from './contexutal-help.service';

declare let elasticlunr: any;

interface ContextualHelpIndex {
    htmlContent: string,
    tags: Array<string>,
    title: string
};
@Component({
    selector: 'app-contextual-help',
    templateUrl: './contextual-help.component.html',
    styleUrls: ['./contextual-help.component.scss'],
    encapsulation: ViewEncapsulation.None
})
export class ContextualHelpComponent extends BasicSubscriber implements OnInit, OnDestroy {

    @Input() id: string;
    @Input() keywords: Array<string>;
    @Input() title: string;

    isTopicView: boolean = true;
    isPinned: boolean = false;
    isVisible: boolean = false;

    lunrIndex: any;
    htmlContentIndexArray: Array<ContextualHelpIndex> = [];
    htmlContentIndex: ContextualHelpIndex = {
        htmlContent: '',
        title: '',
        tags: []
    };

    constructor(
        private service: ContexutalHelpService
    ) {
        super();
        this.service.add(this);
    }

    ngOnInit(): void {
        this.lunrIndex = elasticlunr.Index.load(index);
        this.getHTMLContent(this.lunrIndex.search(this.keywords, {bool: 'AND'}));
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

    open() {
        this.isVisible = true;
    }

    hide() {
        if (this.htmlContentIndexArray.length > 1) {
            this.isTopicView = true;
        }
        this.isVisible = false;
    }

    showContent(htmlContentIndex: ContextualHelpIndex) {
        this.isTopicView = false;
        this.htmlContentIndex = htmlContentIndex;
    }
    navigateBack() {
        this.isTopicView = true;
    }

    ngOnDestroy() {
        this.service.remove(this.id);
    }
}
