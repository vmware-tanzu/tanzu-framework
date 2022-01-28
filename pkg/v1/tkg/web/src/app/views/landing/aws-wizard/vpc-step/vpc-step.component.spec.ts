// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
// App imports
import { APIClient } from 'src/app/swagger';
import AppServices from '../../../../shared/service/appServices';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VpcStepComponent } from './vpc-step.component';
import { AwsField, AwsForm, VpcType } from '../aws-wizard.constants';

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
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        fixture = TestBed.createComponent(VpcStepComponent);
        component = fixture.componentInstance;
        component.setStepRegistrantData({ wizard: 'PickleWizard', step: AwsForm.VPC, formGroup: new FormBuilder().group({}),
            eventFileImported: TkgEventType.AWS_CONFIG_FILE_IMPORTED, eventFileImportError: TkgEventType.AWS_CONFIG_FILE_IMPORT_ERROR});
        fixture.detectChanges();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
    });

    it("should be invalid when VPC CIDR is 192.168.1.1", async(() => {
        component.formGroup.get(AwsField.VPC_TYPE).setValue(VpcType.NEW);
        fixture.detectChanges();
        component.setNewVpcValidators();
        component.formGroup.get(AwsField.VPC_NEW_CIDR).setValue("192.168.1.1");
        expect(component.formGroup.valid).toBeFalsy();
    }));

    it("should be invalid when VPC CIDR is 192.168.1.0/32", async(() => {
        component.formGroup.get(AwsField.VPC_TYPE).setValue(VpcType.NEW);
        fixture.detectChanges();
        component.setNewVpcValidators();
        component.formGroup.get(AwsField.VPC_NEW_CIDR).setValue("192.168.1.0/32");
        expect(component.formGroup.valid).toBeFalsy();
    }));

    it("selecting existing VPC should populate existing VPC CIDR", async(() => {
        component.existingVpcs = [{
            id: 'vpc-1',
            cidr: '100.64.0.0/13'
        }];
        component.formGroup.get(AwsField.VPC_TYPE).setValue(VpcType.EXISTING);
        fixture.detectChanges();
        component.onChangeExistingVpc('vpc-1');
        expect(component.formGroup.get('existingVpcCidr').value).toBe('100.64.0.0/13');
    }));

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();

        const description = component.dynamicDescription();
        expect(description).toEqual('Specify VPC settings for AWS');

        const vpcTypeControl = component.formGroup.controls[AwsField.VPC_TYPE];
        const vpcExistingCidrControl = component.formGroup.controls[AwsField.VPC_EXISTING_CIDR];
        const vpcExistingIdControl = component.formGroup.controls[AwsField.VPC_EXISTING_ID];
        const vpcNewCidrControl = component.formGroup.controls[AwsField.VPC_NEW_CIDR];

        vpcTypeControl.setValue(VpcType.NEW);
        vpcNewCidrControl.setValue('1.2.1.2/12');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'PickleWizard',
                step: AwsForm.VPC,
                description: 'VPC: (new) CIDR: 1.2.1.2/12',
            }
        });

        // NOTE: setting the existing VPC id causes a search of existingVpcs to find the corresponding CIDR,
        // so we need to set up existingVpcs to have the VPC id we're using in the test
        component.existingVpcs = [{
            id: 'someVpc',
            cidr: '3.4.3.4/24',
        }]
        vpcTypeControl.setValue(VpcType.EXISTING);
        vpcExistingIdControl.setValue('someVpc');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'PickleWizard',
                step: AwsForm.VPC,
                description: 'VPC: someVpc CIDR: 3.4.3.4/24',
            }
        });
    });
});
