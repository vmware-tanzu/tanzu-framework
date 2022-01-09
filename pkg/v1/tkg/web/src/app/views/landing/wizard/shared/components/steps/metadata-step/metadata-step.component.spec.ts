// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { MetadataStepComponent } from './metadata-step.component';
import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';

describe('MetadataStepComponent', () => {
    let component: MetadataStepComponent;
    let fixture: ComponentFixture<MetadataStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule
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
            declarations: [MetadataStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(MetadataStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });

        fixture.detectChanges();
    });

    it('should add new label', () => {
        component.addLabel("somekey", "someval");
        component.addLabel("somekey2", "someval2");
        expect(component.clusterLabelsValue).toEqual("somekey:someval, somekey2:someval2");
    });

    it('should delete existing label', () => {
        component.addLabel("akey", "avalue");
        expect(component.clusterLabelsValue).toEqual('akey:avalue');
        component.deleteLabel("akey");
        expect(component.clusterLabelsValue).toEqual('');
    });
});
