import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';

import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';
import { APIClient } from '../../../../../../../swagger/api-client.service';
import { MetadataStepComponent } from './metadata-step.component';

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
