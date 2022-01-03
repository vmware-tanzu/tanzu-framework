// Angular imports
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
// App imports
import { APIClient } from 'src/app/swagger/api-client.service';
import Broker from 'src/app/shared/service/broker';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';
import { VsphereOsImageStepComponent } from '../../../../../vsphere-wizard/vsphere-os-image-step/vsphere-os-image-step.component';
import { VSphereWizardFormService } from 'src/app/shared/service/vsphere-wizard-form.service';

describe('VsphereOsImageStepComponent', () => {
    let component: VsphereOsImageStepComponent;
    let fixture: ComponentFixture<VsphereOsImageStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [VsphereOsImageStepComponent],
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                ValidationService,
                VSphereWizardFormService,
                FormBuilder,
                FieldMapUtilities,
                APIClient,
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        Broker.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(VsphereOsImageStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({});

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should retrieve os image when function invoked', () => {
        const msgSpy = spyOn(Broker.messenger, 'publish').and.callThrough();
        component.retrieveOsImages();
        expect(component.formGroup.get('osImage').value).toBeFalsy();
        expect(msgSpy).toHaveBeenCalled();
    });
});
