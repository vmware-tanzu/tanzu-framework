// Angular modules
import { TestBed, async } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

// App imports
import { LandingComponent } from './landing.component';

describe('LandingComponent', () => {
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule
            ],
            declarations: [
                LandingComponent
            ]
        }).compileComponents();
    }));

    it('should exist', () => {
        const fixture = TestBed.createComponent(LandingComponent);
        const landingComponent = fixture.debugElement.componentInstance;
        expect(landingComponent).toBeTruthy();
    });
});
