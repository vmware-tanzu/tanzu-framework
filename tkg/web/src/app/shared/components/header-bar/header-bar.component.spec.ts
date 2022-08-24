// Angular modules
import { TestBed, async, fakeAsync } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

// App imports
import { HeaderBarComponent } from './header-bar.component';

describe('HeaderBarComponent', () => {
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule
            ],
            declarations: [
                HeaderBarComponent
            ]
        }).compileComponents();
    }));

    it('should exist', () => {
        const fixture = TestBed.createComponent(HeaderBarComponent);
        const landingComponent = fixture.debugElement.componentInstance;
        expect(landingComponent).toBeTruthy();
    });

    it('should call navigateHome() method if user clicks on TKG logo', fakeAsync(() => {
        const fixture = TestBed.createComponent(HeaderBarComponent);
        const comp = fixture.debugElement.componentInstance;
        spyOn(comp, 'navigateHome');
        const elem = fixture.nativeElement.querySelector('.branding');
        elem.click();
        expect(comp.navigateHome).toHaveBeenCalled();
    }));

    it('should call navigateToDocs() method if user clicks on Documentation link', fakeAsync(() => {
        const fixture = TestBed.createComponent(HeaderBarComponent);
        const comp = fixture.debugElement.componentInstance;
        spyOn(comp, 'navigateToDocs');
        const elem = fixture.nativeElement.querySelector('.btn-header-action');
        elem.click();
        expect(comp.navigateToDocs).toHaveBeenCalled();
    }));
});
