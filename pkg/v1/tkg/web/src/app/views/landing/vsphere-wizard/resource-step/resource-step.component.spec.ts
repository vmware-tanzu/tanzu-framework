import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

import { SharedModule } from '../../../../shared/shared.module';
import { ResourceStepComponent } from './resource-step.component';
import { APIClient } from '../../../../swagger/api-client.service';
import { Messenger } from '../../../../shared/service/Messenger';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

describe('ResourceStepComponent', () => {
    let component: ResourceStepComponent;
    let fixture: ComponentFixture<ResourceStepComponent>;
    const messenger = new Messenger();

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                APIClient,
                FormBuilder,
                ValidationService,
                {provide: Messenger, useValue: messenger}
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [ResourceStepComponent]
        }).compileComponents();
    }));

    beforeEach(() => {
        TestBed.inject(ValidationService);
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(ResourceStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should retrieve resources when load resoruce', () => {
        const retrieveRrcSpy = spyOn(component, 'retrieveResourcePools').and.callThrough();
        const retrieveDsSpy = spyOn(component, 'retrieveDatastores').and.callThrough();
        const retrieveVmSpy = spyOn(component, 'retrieveVMFolders').and.callThrough();
        component.loadResourceOptions();
        expect(retrieveRrcSpy).toHaveBeenCalled();
        expect(retrieveDsSpy).toHaveBeenCalled();
        expect(retrieveVmSpy).toHaveBeenCalled();
    });

    it('should retrieve resources when load resoruce', () => {
        component.resetFieldsUponDCChange();
        expect(component.formGroup.get('resourcePool').value).toBeFalsy();
        expect(component.formGroup.get('datastore').value).toBeFalsy();
        expect(component.formGroup.get('vmFolder').value).toBeFalsy();
    });

    it('should retrieve resources when load resoruce', () => {
        const msgSpy = spyOn(messenger, 'publish').and.callThrough();
        component.retrieveResourcePools();
        expect(msgSpy).toHaveBeenCalled();
    });

    it('should retrieve ds when load resoruce', () => {
        const msgSpy = spyOn(messenger, 'publish').and.callThrough();
        component.retrieveDatastores();
        expect(msgSpy).toHaveBeenCalled();
    });

    it('should retrieve vm folders when load resoruce', () => {
        const msgSpy = spyOn(messenger, 'publish').and.callThrough();
        component.retrieveVMFolders();
        expect(msgSpy).toHaveBeenCalled();
    });
});
