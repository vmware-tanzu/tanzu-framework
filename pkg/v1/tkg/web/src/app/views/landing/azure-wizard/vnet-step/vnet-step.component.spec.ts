// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { AzureField, AzureForm } from '../azure-wizard.constants';
import { DataServiceRegistrarTestExtension } from '../../../../testing/data-service-registrar.testextension';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VnetStepComponent } from './vnet-step.component';
import { VSphereResourcePool } from '../../../../swagger/models';

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
        component.setStepRegistrantData({ wizard: 'ZuchiniWizard', step: AzureForm.VNET, formGroup: new FormBuilder().group({}),
            eventFileImported: TkgEventType.AZURE_CONFIG_FILE_IMPORTED, eventFileImportError: TkgEventType.AZURE_CONFIG_FILE_IMPORT_ERROR});

        const dataServiceRegistrar = new DataServiceRegistrarTestExtension();
        AppServices.dataServiceRegistrar = dataServiceRegistrar;
        // we expect the wizard to have registered for these events:
        dataServiceRegistrar.simulateRegistration<VSphereResourcePool>(TkgEventType.AZURE_GET_RESOURCE_GROUPS);
        dataServiceRegistrar.simulateRegistration<VSphereResourcePool>(TkgEventType.AZURE_GET_VNETS);

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
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'ZuchiniWizard',
                step: AzureForm.VNET,
                description: 'Subnet: 4.3.2.1/12',
            }
        });
    });
});
