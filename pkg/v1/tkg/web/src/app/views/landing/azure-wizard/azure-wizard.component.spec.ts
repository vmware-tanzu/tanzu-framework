import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AzureWizardComponent } from './azure-wizard.component';
import { RouterTestingModule } from '@angular/router/testing';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SharedModule } from 'src/app/shared/shared.module';
import { APIClient } from 'src/app/swagger';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import Broker from 'src/app/shared/service/broker';
import { Messenger } from 'src/app/shared/service/Messenger';

describe('AzureWizardComponent', () => {
    let component: AzureWizardComponent;
    let fixture: ComponentFixture<AzureWizardComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                ReactiveFormsModule,
                BrowserAnimationsModule,
                RouterTestingModule,
                SharedModule
            ],
            providers: [
                APIClient,
                FormBuilder,
                AzureWizardFormService
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],

            declarations: [AzureWizardComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        Broker.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(AzureWizardComponent);
        component = fixture.componentInstance;
        component.form = fb.group({
            azureProviderForm: fb.group({
                tenantId: ['']
            }),
            vnetForm: fb.group({
                vnetCidrBlock: ['']
            }),
            azureNodeSettingForm: fb.group({
                controlPlaneSetting: ['']
            }),
            workerazureNodeSettingForm: fb.group({
            }),
            metadataForm: fb.group({
                clusterLocation: ['']
            }),
            registerTmcForm: fb.group({
                tmcRegUrl: ['']
            }),
            networkForm: fb.group({
                clusterServiceCidr: [''],
                clusterPodCidr: [''],
                cniType: ['']
            }),
            identityForm: fb.group({
                identityType: [''],
                issuerURL: ['']
            }),
            ceipOptInForm: fb.group({
                ceipOptIn: ['']
            }),
            osImageForm: fb.group({
            })
        });
        component.clusterType = 'management';
        fixture.detectChanges();
    });

    describe('step description', () => {
        it('should create', () => {
            expect(component).toBeTruthy();
        });

        it('azure provider form', () => {
            const formName = 'azureProviderForm';
            expect(component.getStepDescription(formName))
                .toBe('Validate the Azure provider credentials for Tanzu');
            component.form.get(formName).get('tenantId').setValue('testId');
            expect(component.getStepDescription(formName))
                .toBe('Azure tenant: testId');
        });
        it('vnet form', () => {
            const formName = 'vnetForm';
            expect(component.getStepDescription(formName))
                .toBe('Specify a Azure VNET CIDR');
            component.form.get(formName).get('vnetCidrBlock').setValue('1.1.1.1/24');
            expect(component.getStepDescription(formName))
                .toBe('Subnet: 1.1.1.1/24');
        });
        it('node setting form', () => {
            const formName = 'azureNodeSettingForm';
            expect(component.getStepDescription(formName))
                .toBe('Specifying the resources backing the management cluster');
            component.form.get(formName).get('controlPlaneSetting').setValue('dev');
            expect(component.getStepDescription(formName))
                .toBe('Control plane type: dev');
        });
        it('meta data form', () => {
            const formName = 'metadataForm';
            expect(component.getStepDescription(formName))
                .toBe('Specify metadata for the management cluster');
            component.form.get(formName).get('clusterLocation').setValue('testLocation');
            expect(component.getStepDescription(formName))
                .toBe('Location: testLocation');
        });
        it('network form', () => {
            const formName = 'networkForm';
            expect(component.getStepDescription(formName))
                .toBe('Specify how TKG networking is provided and global network settings');
            component.form.get(formName).get('clusterServiceCidr').setValue('1.1.1.1/23');
            component.form.get(formName).get('clusterPodCidr').setValue('2.2.2.2/23');
            expect(component.getStepDescription(formName))
                .toBe('Cluster service CIDR: 1.1.1.1/23 Cluster POD CIDR: 2.2.2.2/23');
        });
        it('register tmc form', () => {
            const formName = 'registerTmcFrom';
            expect(component.getStepDescription(formName))
                .toBe('Optional: register Tanzu Mission Control');
        });
        it('ceip opt in form', () => {
            const formName = 'ceipOptInForm';
            expect(component.getStepDescription(formName))
                .toBe('Join the CEIP program for TKG');
        });
        it('invalid form', () => {
            const formName = 'invalidForm';
            expect(component.getStepDescription(formName))
                .toBe(`Step ${formName} is not supported yet`);
        })
    });

    it('should generate cli', () => {
        const path = '/testPath/xyz.yaml';
        expect(component.getCli(path)).toBe(`tanzu management-cluster create --file ${path} -v 6`);
    });

    it('should call api to create aws regional cluster', () => {
        const apiSpy = spyOn(component['apiClient'], 'createAzureRegionalCluster').and.callThrough();
        component.createRegionalCluster({});
        expect(apiSpy).toHaveBeenCalled();
    });

    it('should apply TKG config for aws', () => {
        const apiSpy = spyOn(component['apiClient'], 'applyTKGConfigForAzure').and.callThrough();
        component.applyTkgConfig();
        expect(apiSpy).toHaveBeenCalled();
    });
});
