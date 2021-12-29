// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
// App imports
import { APIClient } from '../../../../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { SharedLoadBalancerStepComponent } from './load-balancer-step.component';
import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';

describe('SharedLoadBalancerStepComponent', () => {
    let component: SharedLoadBalancerStepComponent;
    let fixture: ComponentFixture<SharedLoadBalancerStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule,
                BrowserAnimationsModule
            ],
            providers: [
                ValidationService,
                FormBuilder,
                APIClient
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [SharedLoadBalancerStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();

        const fb = new FormBuilder();
        fixture = TestBed.createComponent(SharedLoadBalancerStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });

        fixture.detectChanges();
    });

    it('should call get clouds when controller credentials have been validated', () => {
        const apiSpy = spyOn(component['apiClient'], 'getAviClouds').and.callThrough();
        component.getClouds();
        expect(apiSpy).toHaveBeenCalled();
    });

    it('should call get service engine groups when controller credentials have been validated', () => {
        const apiSpy = spyOn(component['apiClient'], 'getAviServiceEngineGroups').and.callThrough();
        component.getServiceEngineGroups();
        expect(apiSpy).toHaveBeenCalled();
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
