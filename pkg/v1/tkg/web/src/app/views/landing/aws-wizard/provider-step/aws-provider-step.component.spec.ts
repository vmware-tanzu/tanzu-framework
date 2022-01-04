// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';
import { By } from '@angular/platform-browser';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
// Third party imports
import { of, empty, throwError, Observable } from 'rxjs';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { AwsField } from '../aws-wizard.constants';
import { AwsProviderStepComponent } from './aws-provider-step.component';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

describe('AwsProviderStepComponent', () => {
    let component: AwsProviderStepComponent;
    let fixture: ComponentFixture<AwsProviderStepComponent>;
    let mockedApiService;

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
                {
                    provide: APIClient,
                    useValue: jasmine.createSpyObj(
                        'APIClient',
                        ['getAWSRegions', 'setAWSEndpoint', 'getAWSCredentialProfiles']
                    )
                }
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [AwsProviderStepComponent]
        })
            .compileComponents();

        mockedApiService = TestBed.get(APIClient);
        mockedApiService.getAWSRegions.and.returnValue(of(["US-WEST", "US-EAST"]));
        mockedApiService.getAWSCredentialProfiles.and.returnValue(of(['profile1', 'profile2']));
        mockedApiService.setAWSEndpoint.and.returnValue(empty());
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(AwsProviderStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({});

        fixture.detectChanges();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it("valid credentials fields should activate connect button", async(() => {

        component.ngOnInit();
        let connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
        expect(connectBtn.nativeElement.disabled).toBeTruthy();

        component.setFieldValue(AwsField.PROVIDER_PROFILE_NAME, 'profile1');
        component.setFieldValue(AwsField.PROVIDER_SESSION_TOKEN, 'token1');
        component.setFieldValue(AwsField.PROVIDER_ACCESS_KEY, 'accessKey1');
        component.setFieldValue(AwsField.PROVIDER_SECRET_ACCESS_KEY, 'secretAccessKey1');
        component.setFieldValue(AwsField.PROVIDER_REGION, 'region1');
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
                expect(connectBtn).toBeTruthy();
                expect(connectBtn.nativeElement.disabled).toBeFalsy();
            }
        );
    }));

    it("invalid (blank) credential fields should deactivate connect button", async(() => {
        component.ngOnInit();
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
                expect(connectBtn.nativeElement.disabled).toBeTruthy();
            }
        );
    }));

    it("should be successful when clicked", async(() => {
        mockedApiService.setAWSEndpoint.and.returnValue(empty());

        const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));

        connectBtn.nativeElement.click();
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                const globalError = fixture.debugElement.query(By.css("app-alert-notification"));
                expect(globalError.nativeElement.innerText).toBe("");
            }
        );
    }));

    it("should show an error when connect button clicked and endpoint throws error", async(() => {
        mockedApiService.setAWSEndpoint.and.returnValue(throwError({ status: 400,
            error: {message: 'failed to set aws endpoint' }}));

        component.ngOnInit();
        component.setFieldValue(AwsField.PROVIDER_PROFILE_NAME, 'profile2');
        component.setFieldValue(AwsField.PROVIDER_SESSION_TOKEN, 'token1');
        component.setFieldValue(AwsField.PROVIDER_ACCESS_KEY, 'accessKey1');
        component.setFieldValue(AwsField.PROVIDER_SECRET_ACCESS_KEY, 'secretAccessKey1');
        component.setFieldValue(AwsField.PROVIDER_REGION, 'region1');
        expect(component.errorNotification).toBeFalsy();

        fixture.whenStable().then( () => {
            fixture.detectChanges();    // this is necessary to pick up the field changes above and thereby activate the connect btn
            const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
            expect(connectBtn.nativeElement.disabled).toBeFalsy('connect button is unexpectedly deactivated');
            connectBtn.nativeElement.click();
            fixture.whenStable().then(() => {
                fixture.detectChanges();
                expect(component.validCredentials).toBeFalsy();
                expect(component.errorNotification).not.toBeFalsy();
            });
        });
    }));
})
