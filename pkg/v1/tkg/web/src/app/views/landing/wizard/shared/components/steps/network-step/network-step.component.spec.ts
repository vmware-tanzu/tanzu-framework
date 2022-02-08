// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from 'src/app/swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { NetworkField } from './network-step.fieldmapping';
import { SharedModule } from 'src/app/shared/shared.module';
import { SharedNetworkStepComponent } from "./network-step.component";
import { ValidationService } from '../../../validation/validation.service';
import { WizardForm } from '../../../constants/wizard.constants';

describe('networkStepComponent', () => {
    let component: SharedNetworkStepComponent;
    let fixture: ComponentFixture<SharedNetworkStepComponent>;

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
                // CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [SharedNetworkStepComponent]
        }).compileComponents();
    }));
    beforeEach(() => {
        AppServices.messenger = new Messenger();
        fixture = TestBed.createComponent(SharedNetworkStepComponent);
        component = fixture.componentInstance;
        component.setInputs('BozoWizard', WizardForm.NETWORK, new FormBuilder().group({}));
        component.ngOnInit();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    describe('should generate a full no proxy list', () => {
        it ('should return empty string', () => {
            component.generateFullNoProxy();
            expect(component.fullNoProxy).toBe('');
        });

        it('should have a complete no proxy list', () => {
            component.additionalNoProxyInfo = '10.0.0.0/16,169.254.0.0/16';
            component.formGroup.setValue({
                proxySettings: true,
                httpProxyUrl: 'http://myproxy.com',
                httpProxyUsername: 'username1',
                httpProxyPassword: 'password1',
                isSameAsHttp: true,
                httpsProxyUrl: 'http://myproxy.com',
                httpsProxyUsername: 'username1',
                httpsProxyPassword: 'password1',
                cniType: 'Antrea',
                clusterServiceCidr: '100.64.0.0/13',
                clusterPodCidr: '100.96.0.0/11',
                noProxy: 'noproxy.yourdomain.com,192.168.0.0/24'
            });
            expect(component.fullNoProxy).toBe('noproxy.yourdomain.com,192.168.0.0/24,10.0.0.0/16,169.254.0.0/16,' +
            '100.64.0.0/13,100.96.0.0/11,localhost,127.0.0.1,.svc,.svc.cluster.local');
        });

        it('should generate complete no proxy list correctly if there are more commas in the noProxy field', () => {
            component.additionalNoProxyInfo = '10.0.0.0/16,169.254.0.0/16';
            component.formGroup.setValue({
                proxySettings: true,
                httpProxyUrl: 'http://myproxy.com',
                httpProxyUsername: 'username1',
                httpProxyPassword: 'password1',
                isSameAsHttp: true,
                httpsProxyUrl: 'http://myproxy.com',
                httpsProxyUsername: 'username1',
                httpsProxyPassword: 'password1',
                cniType: 'Antrea',
                clusterServiceCidr: '100.64.0.0/13',
                clusterPodCidr: '100.96.0.0/11',
                noProxy: 'noproxy.yourdomain.com,192.168.0.0/24,,,,,',
            });
            expect(component.fullNoProxy).toBe('noproxy.yourdomain.com,192.168.0.0/24,10.0.0.0/16,169.254.0.0/16,' +
            '100.64.0.0/13,100.96.0.0/11,localhost,127.0.0.1,.svc,.svc.cluster.local');
        });
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        const serviceCidrControl = component.formGroup.controls[NetworkField.CLUSTER_SERVICE_CIDR];
        const podCidrControl = component.formGroup.controls[NetworkField.CLUSTER_POD_CIDR];

        serviceCidrControl.setValue('');
        podCidrControl.setValue('');
        expect(component.dynamicDescription()).toEqual(SharedNetworkStepComponent.description);

        podCidrControl.setValue('1.2.3.4/12');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.NETWORK,
                description: 'Cluster Pod CIDR: 1.2.3.4/12'
            }
        });

        serviceCidrControl.setValue('5.6.7.8/16');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.NETWORK,
                description: 'Cluster Service CIDR: 5.6.7.8/16 Cluster Pod CIDR: 1.2.3.4/12'
            }
        });
    });
});
