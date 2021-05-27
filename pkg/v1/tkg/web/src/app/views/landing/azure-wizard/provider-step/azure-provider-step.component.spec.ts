import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from '../../../../shared/shared.module';
import { AzureProviderStepComponent } from './azure-provider-step.component';
import { APIClient } from '../../../../swagger/api-client.service';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { of, throwError, Observable } from 'rxjs';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';

describe('AzureProviderStepComponent', () => {
    let component: AzureProviderStepComponent;
    let fixture: ComponentFixture<AzureProviderStepComponent>;
    let apiService: APIClient;
    let wizardFormService: AzureWizardFormService;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule,
                BrowserAnimationsModule
            ],
            providers: [
                ValidationService,
                AzureWizardFormService,
                FormBuilder,
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
        Broker.messenger = new Messenger();
        wizardFormService = TestBed.inject(AzureWizardFormService);
        apiService = TestBed.inject(APIClient);

        const fb = new FormBuilder();
        fixture = TestBed.createComponent(AzureProviderStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
            tenantId: ['']
        });
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should setup AZURE_GET_RESOURCE_GROUPS event handler', () => {
        component.ngOnInit();
        wizardFormService.publishError(TkgEventType.AZURE_GET_RESOURCE_GROUPS, 'test error');
        expect(component.errorNotification).toBe('Failed to retrieve resource groups for the particular region. test error');

        const resourceGrouop = [
            {id: 1, location: 'us-west', name: 'resource-group1'},
            {id: 2, location: 'us-east', name: 'resource-group2'},
            {id: 3, location: 'us-south', name: 'resource-group3'}
        ];
        wizardFormService.publishData(TkgEventType.AZURE_GET_RESOURCE_GROUPS, resourceGrouop);
        expect(component.resourceGroupSelection).toBeNull();
        expect(component.resourceGroups).toEqual(resourceGrouop);
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
        component.verifyCredentails();
        expect(component.errorNotification).toBe('');
        expect(component.validCredentials).toBeTruthy();
        expect(regions).toHaveBeenCalled();
    });

    it('should show error message if credential can not be verified', () => {
        spyOn(apiService, 'setAzureEndpoint').and.returnValue(throwError({error : {message: 'oops!'}}));
        component.verifyCredentails();
        expect(component.errorNotification).toBe('oops!');
        expect(component.validCredentials).toBeFalsy();
        expect(component.regions).toEqual([]);
        expect(component.formGroup.get('region').value).toBe('');
    });

    it('should show different resource based on option', () => {
        component.show('existing');
        expect(component.formGroup.get('resourceGroupCustom').value).toBe('');
        component.show('custom');
        expect(component.formGroup.get('resourceGroupExisting').value).toBe('');
    });

    it('should handle resource group name change', () => {
        const messengerSpy = spyOn(Broker.messenger, 'publish').and.callThrough();
        component.onResourceGroupNameChange();
        expect(messengerSpy).toHaveBeenCalled();
        expect(messengerSpy).toHaveBeenCalledWith({
            type: TkgEventType.AZURE_RESOURCEGROUP_CHANGED,
            payload: ''
        });
    });

    // it("should be enabled", async(() => {
    //     mockedApiService.getAWSEndpoint.and.returnValue(of({
    //         accessKeyID: "mykeyId",
    //         region: "US-WEST",
    //         secretAccessKey: "myKey",
    //         sshKeyName: "myKeyName"
    //     }));

    //     component.ngOnInit();
    //     fixture.whenStable().then(
    //         () => {
    //             fixture.detectChanges();
    //             const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
    //             expect(connectBtn.nativeElement.disabled).toBeFalsy();
    //         }
    //     );
    // }));

    // it("should be disabled", async(() => {
    //     mockedApiService.getAWSEndpoint.and.returnValue(of({
    //         accessKeyID: "mykeyId",
    //         region: "US-WEST",
    //         secretAccessKey: "myKey",
    //     }));

    //     component.ngOnInit();
    //     fixture.whenStable().then(
    //         () => {
    //             fixture.detectChanges();
    //             const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
    //             expect(connectBtn.nativeElement.disabled).toBeTruthy();
    //         }
    //     );
    // }));

    // it("should be successful when clicked", async(() => {
    //     mockedApiService.getAWSEndpoint.and.returnValue(of({
    //         accessKeyID: "mykeyId",
    //         region: "US-WEST",
    //         secretAccessKey: "myKey",
    //         sshKeyName: "myKeyName"
    //     }));
    //     mockedApiService.setAWSEndpoint.and.returnValue(empty());

    //     const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));

    //     connectBtn.nativeElement.click();
    //     fixture.whenStable().then(
    //         () => {
    //             fixture.detectChanges();
    //             const globalError = fixture.debugElement.query(By.css("app-step-form-notification"));
    //             expect(globalError.nativeElement.innerText).toBe("");
    //         }
    //     );
    // }));

    // it("should show an error when clicked", async(() => {
    //     mockedApiService.getAWSEndpoint.and.returnValue(of({
    //         accessKeyID: "mykeyId",
    //         region: "US-WEST",
    //         secretAccessKey: "myKey",
    //         sshKeyName: "myKeyName"
    //     }));
    //     mockedApiService.setAWSEndpoint.and.returnValue(throwError({ status: 400,
    //         error: {message: 'failed to set aws endpoint' }}));

    //     component.ngOnInit();
    //     expect(component.errorNotification).toBeFalsy();

    //     const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
    //     connectBtn.nativeElement.click();
    //     fixture.whenStable().then(
    //         () => {
    //             fixture.detectChanges();
    //             expect(component.errorNotification).not.toBeFalsy();
    //             expect(component.validCredentials).toBeFalsy();
    //         }
    //     );
    // }));

})
