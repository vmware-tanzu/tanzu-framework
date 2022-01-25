// Angular imports
import { async, ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';

// Library imports
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { APIClient } from 'tanzu-mgmt-plugin-api-lib';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { SharedModule } from 'src/app/shared/shared.module';
import { SSLThumbprintModalComponent } from '../../wizard/shared/components/modals/ssl-thumbprint-modal/ssl-thumbprint-modal.component';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereProviderStepComponent } from './vsphere-provider-step.component';
import { VsphereField } from '../vsphere-wizard.constants';

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

        fixture = TestBed.createComponent(VSphereProviderStepComponent);
        component = fixture.componentInstance;
        component.setInputs('BozoWizard', 'vsphereProviderForm', new FormBuilder().group({}));

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

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const vcenterIPControl = component.formGroup.get(VsphereField.PROVIDER_VCENTER_ADDRESS);
        const datacenterControl = component.formGroup.get(VsphereField.PROVIDER_DATA_CENTER);

        expect(component.dynamicDescription()).toEqual('Validate the vSphere provider account for Tanzu');

        component.vsphereVersion = 'CLOWNVERSION';
        expect(component.dynamicDescription()).toEqual('Validate the vSphere CLOWNVERSION provider account for Tanzu');

        vcenterIPControl.setValue('1.2.1.2');
        datacenterControl.setValue('DATACENTER');
        component.onLoginSuccess({});
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereProviderForm',
                description: 'vCenter 1.2.1.2 connected',
            }
        });
    });
});
