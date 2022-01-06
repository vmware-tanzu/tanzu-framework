// Angular imports
import { async, ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
// Third party imports
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
// App imports
import { APIClient } from 'src/app/swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { SharedModule } from 'src/app/shared/shared.module';
import { SSLThumbprintModalComponent } from '../../wizard/shared/components/modals/ssl-thumbprint-modal/ssl-thumbprint-modal.component';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereProviderStepComponent } from './vsphere-provider-step.component';

describe('VSphereProviderStepComponent', () => {
    let component: VSphereProviderStepComponent;
    let fixture: ComponentFixture<VSphereProviderStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes([
                    { path: 'ui', component: VSphereProviderStepComponent }
                ]),
                ReactiveFormsModule,
                SharedModule,
                BrowserAnimationsModule
            ],
            providers: [
                ValidationService,
                FormBuilder,
                FieldMapUtilities,
                APIClient
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [
                VSphereProviderStepComponent,
                SSLThumbprintModalComponent
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();

        const fb = new FormBuilder();
        fixture = TestBed.createComponent(VSphereProviderStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open SSL thumbprint modal when connect vc', fakeAsync(() => {
        const apiSpy = spyOn(component['apiClient'], 'getVsphereThumbprint').and.returnValue(of({insecure: false}).pipe(delay(1)));
        spyOn(component.sslThumbprintModal, 'open');
        component.connectVC();
        tick(1);
        expect(apiSpy).toHaveBeenCalled();
        expect(component.sslThumbprintModal.open).toHaveBeenCalled();
    }));

    it('should call get datacenter when retrieve trigger datacenter', () => {
        const apiSpy = spyOn(component['apiClient'], 'getVSphereDatacenters').and.callThrough();
        component.retrieveDatacenters();
        expect(apiSpy).toHaveBeenCalled();
    });

    it('should return disabled when username is not valid', () => {
        expect(component.getDisabled()).toBeTruthy();
    });

    it('should set vsphere modal open when show method is triggered', () => {
        component.showVSphereWithK8Modal();
        expect(component.vSphereWithK8ModalOpen).toBeTruthy();
    });
});
