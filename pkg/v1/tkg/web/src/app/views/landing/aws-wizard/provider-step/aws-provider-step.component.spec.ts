import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { FormBuilder } from '@angular/forms';
import { By } from '@angular/platform-browser';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from '../../../../shared/shared.module';
import { AwsProviderStepComponent } from './aws-provider-step.component';
import { APIClient } from '../../../../swagger/api-client.service';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { of, empty, throwError, Observable } from 'rxjs';
import Broker from 'src/app/shared/service/broker';
import { Messenger } from 'src/app/shared/service/Messenger';

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
                {
                    provide: APIClient,
                    useValue: jasmine.createSpyObj(
                        'APIClient',
                        ['getAWSEndpoint', 'getAWSRegions', 'setAWSEndpoint', 'getAWSCredentialProfiles']
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
        mockedApiService.getAWSEndpoint.and.returnValue(of({
            accessKeyID: "mykeyId",
            region: "US-WEST",
            secretAccessKey: "myKey",
            sshKeyName: "myKeyName"
        }));
        mockedApiService.getAWSCredentialProfiles.and.returnValue(of(['profile1', 'profile2']));
        mockedApiService.setAWSEndpoint.and.returnValue(empty());
    }));

    beforeEach(() => {
        Broker.messenger = new Messenger();
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

    it("should be enabled", async(() => {
        mockedApiService.getAWSEndpoint.and.returnValue(of({
            accessKeyID: "mykeyId",
            region: "US-WEST",
            secretAccessKey: "myKey",
            sshKeyName: "myKeyName"
        }));

        component.ngOnInit();
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
                expect(connectBtn).toBeTruthy();
                expect(connectBtn.nativeElement.disabled).toBeFalsy();
            }
        );
    }));

    it("should be disabled", async(() => {
        mockedApiService.getAWSEndpoint.and.returnValue(of({
            accessKeyID: "mykeyId",
            region: "US-WEST",
            secretAccessKey: "myKey",
        }));

        component.ngOnInit();
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
                expect(connectBtn).toBeTruthy();
                expect(connectBtn.nativeElement.disabled).toBeFalsy();
            }
        );
    }));

    it("should be successful when clicked", async(() => {
        mockedApiService.getAWSEndpoint.and.returnValue(of({
            accessKeyID: "mykeyId",
            region: "US-WEST",
            secretAccessKey: "myKey",
            sshKeyName: "myKeyName"
        }));
        mockedApiService.setAWSEndpoint.and.returnValue(empty());

        const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));

        connectBtn.nativeElement.click();
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                const globalError = fixture.debugElement.query(By.css("app-step-form-alert-notification"));
                expect(globalError.nativeElement.innerText).toBe("");
            }
        );
    }));

    it("should show an error when clicked", async(() => {
        mockedApiService.getAWSEndpoint.and.returnValue(of({
            accessKeyID: "mykeyId",
            region: "US-WEST",
            secretAccessKey: "myKey",
            sshKeyName: "myKeyName"
        }));
        mockedApiService.setAWSEndpoint.and.returnValue(throwError({ status: 400,
            error: {message: 'failed to set aws endpoint' }}));

        component.ngOnInit();
        expect(component.errorNotification).toBeFalsy();

        const connectBtn = fixture.debugElement.query(By.css("button.btn-primary"));
        connectBtn.nativeElement.click();
        fixture.whenStable().then(
            () => {
                fixture.detectChanges();
                expect(component.errorNotification).not.toBeFalsy();
                expect(component.validCredentials).toBeFalsy();
            }
        );
    }));

})
