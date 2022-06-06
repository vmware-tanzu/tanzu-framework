// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormArray, FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
// App imports
import { APIClient } from '../../../../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
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
        component.formGroup = fb.group({});

        fixture.detectChanges();
    });

    it('should initialize tkgLabelsConfig', () => {
        component.ngOnInit();
        const config = component.tkgLabelsConfig;

        expect(config.label.title).toEqual('CLUSTER LABELS (OPTIONAL)');
        expect(config.forms.parent.get('clusterLabels')).toBeInstanceOf(FormArray);
        expect(config.fields.clusterTypeDescriptor).toEqual('Workload');
    })

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
});
