// Angular modules
import { TestBed, async } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';

// App imports
import { SharedModule } from '../../../shared/shared.module';
import { WcpRedirectComponent } from './wcp-redirect.component';

describe('WcpRedirectComponent', () => {
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                SharedModule
            ],
            declarations: [
                WcpRedirectComponent
            ]
        }).compileComponents();
    }));

    it('should exist', () => {
        const fixture = TestBed.createComponent(WcpRedirectComponent);
        const exitComponent = fixture.debugElement.componentInstance;
        expect(exitComponent).toBeTruthy();
    });
});
