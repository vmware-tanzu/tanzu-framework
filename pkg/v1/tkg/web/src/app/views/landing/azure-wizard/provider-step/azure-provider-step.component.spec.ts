// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";

// Library imports
import { of, throwError, Observable } from 'rxjs';
import { APIClient, AzureResourceGroup } from 'tanzu-ui-api-lib';

// App imports
import AppServices from '../../../../shared/service/appServices';
import { AzureProviderStepComponent } from './azure-provider-step.component';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { DataServiceRegistrarTestExtension } from '../../../../testing/data-service-registrar.testextension';
import { AzureForm } from '../azure-wizard.constants';

describe('AzureProviderStepComponent', () => {
    let component: AzureProviderStepComponent;
    let fixture: ComponentFixture<AzureProviderStepComponent>;
    let apiService: APIClient;

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
                FieldMapUtilities,
                APIClient
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [AzureProviderStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        AppServices.dataServiceRegistrar = new DataServiceRegistrarTestExtension();
        apiService = TestBed.inject(APIClient);

        fixture = TestBed.createComponent(AzureProviderStepComponent);
        component = fixture.componentInstance;
        component.setInputs('CarrotWizard', AzureForm.PROVIDER, new FormBuilder().group({}));
        component.ngOnInit();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should setup AZURE_GET_RESOURCE_GROUPS event handler', () => {
        const dataServiceRegistrar = AppServices.dataServiceRegistrar as DataServiceRegistrarTestExtension;
        // The wizard is expected to have registered this event
        dataServiceRegistrar.simulateRegistration<AzureResourceGroup>(TkgEventType.AZURE_GET_RESOURCE_GROUPS);

        component.ngOnInit();
        dataServiceRegistrar.simulateError(TkgEventType.AZURE_GET_RESOURCE_GROUPS, 'test error');
        expect(component.errorNotification).toBe('test error');

        const resourceGroup = [
            {id: 1, location: 'us-west', name: 'resource-group1'},
            {id: 2, location: 'us-east', name: 'resource-group2'},
            {id: 3, location: 'us-south', name: 'resource-group3'}
        ];
        dataServiceRegistrar.simulateData(TkgEventType.AZURE_GET_RESOURCE_GROUPS, resourceGroup);
        expect(component.resourceGroups).toEqual(resourceGroup);
    });

    it('should init azure credentials', () => {
        spyOn(apiService, 'getAzureEndpoint').and.returnValue(of({
            tenantId: "azure-tenant1",
            clientId: "azure-client-12343",
            clientSecret: "",
            subscriptionId: "azure-subscription-12342-asdf3"
        }));
        component.savedMetadata = null;
        component.initAzureCredentials();
        expect(component.formGroup.controls['tenantId'].value).toBe('azure-tenant1');
    });

    it('should show error message when the credential can not be retrieved', () => {
        spyOn(apiService, 'getAzureEndpoint').and.returnValue(throwError(new Error('oops!')));
        component.initAzureCredentials();
        expect(component.errorNotification).toBe('Unable to retrieve Azure credentials');
    });

    it('should fetch regions information', () => {
        const regions = [
            {name: 'westus', displayName: 'West US'},
            {name: 'northcentralus', displayName: 'North central US'},
            {name: 'southcentralus', displayName: 'South central US'}
        ];

        spyOn(apiService, 'getAzureRegions').and.returnValue(of(regions));
        component.getRegions();
        expect(component.loadingRegions).toBeFalsy();
        expect(component.regions).toEqual(regions);
    });

    it('should show error message when the region info can not be retrieved', () => {
        spyOn(apiService, 'getAzureRegions').and.returnValue(throwError(new Error('oops!')));
        component.getRegions();
        expect(component.errorNotification).toBe('Unable to retrieve Azure regions');
    })

    it('should verify credentials', () => {
        spyOn(apiService, 'setAzureEndpoint').and.returnValues(new Observable(subscriber => {
            subscriber.next();
        }));
        const regions = spyOn(component, 'getRegions').and.stub();
        component.verifyCredentials();
        expect(component.errorNotification).toBe('');
        expect(component.validCredentials).toBeTruthy();
        expect(regions).toHaveBeenCalled();
    });

    it('should show error message if credential can not be verified', () => {
        spyOn(apiService, 'setAzureEndpoint').and.returnValue(throwError({error: {message: 'oops!'}}));
        component.verifyCredentials();
        expect(component.errorNotification).toBe('oops!');
        expect(component.validCredentials).toBeFalsy();
        expect(component.regions).toEqual([]);
        expect(component.formGroup.get('region').value).toBe('');
    });

    it('should show different resource based on option', () => {
        component.showResourceGroupExisting();
        expect(component.formGroup.get('resourceGroupCustom').value).toBe('');
        component.showResourceGroupCustom();
        expect(component.formGroup.get('resourceGroupExisting').value).toBe('');
    });

    it('should handle resource group name change', () => {
        const messengerSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.onResourceGroupNameChange();
        expect(messengerSpy).toHaveBeenCalled();
        expect(messengerSpy).toHaveBeenCalledWith({
            type: TkgEventType.AZURE_RESOURCEGROUP_CHANGED,
            payload: ''
        });
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        const tenantControl = component.formGroup.get('tenantId');
        tenantControl.setValue('');

        const description = component.dynamicDescription();
        expect(description).toEqual('Validate the Azure provider credentials for Tanzu');

        tenantControl.setValue('RIDDLER');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'CarrotWizard',
                step: AzureForm.PROVIDER,
                description: 'Azure tenant: RIDDLER',
            }
        });
    })
});
