// Angular imports
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TestBed, async } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientModule } from '@angular/common/http';

// App imports
import { APIClient } from 'tanzu-mgmt-plugin-api-lib';
import { AppComponent } from './app.component';
import { ThemeToggleComponent } from './shared/components/theme-toggle/theme-toggle.component';
import { BrandingService } from './shared/service/branding.service';
import { BrandingServiceStub } from './testing/branding.service.stub';

describe('AppComponent', () => {
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                HttpClientModule,
                RouterTestingModule
            ],
            providers: [
                APIClient,
                { provide: BrandingService, useClass: BrandingServiceStub }
            ],
            declarations: [
                AppComponent,
                ThemeToggleComponent
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ]
        }).compileComponents();
    }));

    it('should create the app', () => {
        const fixture = TestBed.createComponent(AppComponent);
        const app = fixture.debugElement.componentInstance;
        expect(app).toBeTruthy();
    });
});
