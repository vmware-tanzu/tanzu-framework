// Angular modules
import { TestBed, async } from '@angular/core/testing';

// App imports
import { IncompatibleComponent } from './incompatible.component';

describe('IncompatibleComponent', () => {
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [],
            declarations: [
                IncompatibleComponent
            ]
        }).compileComponents();
    }));

    it('should exist', () => {
        const fixture = TestBed.createComponent(IncompatibleComponent);
        const incompatibleComponent = fixture.debugElement.componentInstance;
        expect(incompatibleComponent).toBeTruthy();
    });
});
