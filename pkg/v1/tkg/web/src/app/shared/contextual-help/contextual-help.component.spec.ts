import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ContextualHelpComponent } from './contextual-help.component';
import mockIndex from '../../../contextualHelpDocs/mockIndex.json';

declare let elasticlunr: any;

describe('ContextualHelpComponent', () => {
    let component: ContextualHelpComponent;
    let fixture: ComponentFixture<ContextualHelpComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
        declarations: [ ContextualHelpComponent ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ContextualHelpComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
        component.lunrIndex = elasticlunr.Index.load(mockIndex);
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show contextual help component', () => {
        component.getHTMLContent(['step1']);
        component.open();
        expect(component.isVisible).toBeTrue();
        expect(component.htmlContentIndexArray.length).toEqual(1);
    });

    it('should hide contextual help component', () => {
        component.hide();
        expect(component.isVisible).toBeFalse();
    });

    it('should show detail for a topic', () => {
        const mockData = {
            htmlContent: '<p>hello world</p>',
            topicIds: ['step1'],
            topicTitle: 'docker step 1'
        };
        component.showContent(mockData);
        expect(component.isTopicView).toBeFalse();
        expect(component.htmlContentIndex).toEqual(mockData);
    });

    it('should navigate back to topic', () => {
        component.navigateBack();
        expect(component.isTopicView).toBeTrue();
    });
});
