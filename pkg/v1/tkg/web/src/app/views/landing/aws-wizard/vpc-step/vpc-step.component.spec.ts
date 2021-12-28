import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { VpcStepComponent } from './vpc-step.component';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { APIClient } from 'src/app/swagger';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import Broker from 'src/app/shared/service/broker';
import { Messenger } from 'src/app/shared/service/Messenger';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import ServiceBroker from '../../../../shared/service/service-broker';

describe('VpcComponent', () => {
    let component: VpcStepComponent;
    let fixture: ComponentFixture<VpcStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [VpcStepComponent],
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                APIClient,
                ValidationService,
                FormBuilder,
                FieldMapUtilities,
                ServiceBroker,
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
        fixture = TestBed.createComponent(VpcStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });
        fixture.detectChanges();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
    });

    it("should be invalid when VPC CIDR is 192.168.1.1", async(() => {
        component.formGroup.get('vpcType').setValue("new");
        fixture.detectChanges();
        component.setNewVpcValidators();
        component.formGroup.get('vpc').setValue("192.168.1.1");
        expect(component.formGroup.valid).toBeFalsy();
    }));

    it("should be invalid when VPC CIDR is 192.168.1.0/32", async(() => {
        component.formGroup.get('vpcType').setValue("new");
        fixture.detectChanges();
        component.setNewVpcValidators();
        component.formGroup.get('vpc').setValue("192.168.1.0/32");
        expect(component.formGroup.valid).toBeFalsy();
    }));

    it("selecting existing VPC should populate existing VPC CIDR", async(() => {
        component.existingVpcs = [{
            id: 'vpc-1',
            cidr: '100.64.0.0/13'
        }];
        component.formGroup.get('vpcType').setValue("existing");
        fixture.detectChanges();
        component.existingVpcOnChange('vpc-1');
        expect(component.formGroup.get('existingVpcCidr').value).toBe('100.64.0.0/13');
    }));
});
