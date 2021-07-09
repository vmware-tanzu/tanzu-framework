import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import Broker from 'src/app/shared/service/broker';
import { Messenger } from 'src/app/shared/service/Messenger';
import { SharedModule } from 'src/app/shared/shared.module';
import { APIClient } from 'src/app/swagger/api-client.service';
import { ValidationService } from '../../../validation/validation.service';
import { SharedNetworkStepComponent } from "./network-step.component";

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
                APIClient
            ],
            schemas: [
                // CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [SharedNetworkStepComponent]
        }).compileComponents();
    }));
    beforeEach(() => {
        Broker.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(SharedNetworkStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
            noProxy: ''
        });
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
});
