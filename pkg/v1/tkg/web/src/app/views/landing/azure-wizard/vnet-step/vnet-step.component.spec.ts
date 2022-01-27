// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-management-cluster-api';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VnetStepComponent } from './vnet-step.component';
import { AzureField, AzureForm } from '../azure-wizard.constants';

describe('VnetStepComponent', () => {
    let component: VnetStepComponent;
    let fixture: ComponentFixture<VnetStepComponent>;

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
            declarations: [VnetStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();

        fixture = TestBed.createComponent(VnetStepComponent);
        component = fixture.componentInstance;
        component.setInputs('ZuchiniWizard', AzureForm.VNET,  new FormBuilder().group({}));

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();

        const staticDescription = component.dynamicDescription();
        expect(staticDescription).toEqual('Specify an Azure VNET CIDR')

        const customCidrControl = component.formGroup.get(AzureField.VNET_CUSTOM_CIDR);
        customCidrControl.setValue('4.3.2.1/12');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'ZuchiniWizard',
                step: AzureForm.VNET,
                description: 'Subnet: 4.3.2.1/12',
            }
        });
    });
});
