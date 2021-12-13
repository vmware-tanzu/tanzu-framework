import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AzureWizardComponent } from './azure-wizard.component';
import { RouterTestingModule } from '@angular/router/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SharedModule } from 'src/app/shared/shared.module';
import { APIClient } from 'src/app/swagger';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AzureWizardFormService } from 'src/app/shared/service/azure-wizard-form.service';
import Broker from 'src/app/shared/service/broker';
import { Messenger } from 'src/app/shared/service/Messenger';
import { ClusterType, WizardForm } from "../wizard/shared/constants/wizard.constants";
import { AzureForm } from './azure-wizard.constants';

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
            metadataForm: fb.group({
                clusterLocation: ['']
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
        component.clusterTypeDescriptor = '' + ClusterType.Management;
        fixture.detectChanges();
    });

    describe('step description', () => {
        it('should create', () => {
            expect(component).toBeTruthy();
        });

        it('azure provider form', () => {
            expect(component.AzureProviderForm.description).toBe('Validate the Azure provider credentials for Tanzu');
            component.form.get(AzureForm.PROVIDER).get('tenantId').setValue('testId');
            expect(component.AzureProviderForm.description).toBe('Azure tenant: testId');
        });
        it('vnet form', () => {
            expect(component.AzureVnetFormDescription).toBe('Specify a Azure VNET CIDR');
            component.form.get(AzureForm.VNET).get('vnetCidrBlock').setValue('1.1.1.1/24');
            expect(component.AzureVnetFormDescription).toBe('Subnet: 1.1.1.1/24');
        });
        it('node setting form', () => {
            expect(component.AzureNodeSettingFormDescription).toBe('Specifying the resources backing the management cluster');
            component.form.get(AzureForm.NODESETTING).get('controlPlaneSetting').setValue('dev');
            expect(component.AzureNodeSettingFormDescription).toBe('Control plane type: dev');
        });
        it('meta data form', () => {
            expect(component.MetadataForm.description).toBe('Specify metadata for the management cluster');
            component.form.get(WizardForm.METADATA).get('clusterLocation').setValue('testLocation');
            expect(component.MetadataForm.description).toBe('Location: testLocation');
        });
        it('network form', () => {
            expect(component.NetworkForm.description).toBe('Specify how TKG networking is provided and global network settings');
            const networkForm = component.form.get(WizardForm.NETWORK);
            networkForm.get('clusterServiceCidr').setValue('1.1.1.1/23');
            networkForm.get('clusterPodCidr').setValue('2.2.2.2/23');
            expect(component.NetworkForm.description).toBe('Cluster service CIDR: 1.1.1.1/23 Cluster POD CIDR: 2.2.2.2/23');
        });
        it('ceip opt in form', () => {
            expect(component.CeipFormDescription)
                .toBe('Join the CEIP program for TKG');
        });
    });

    it('should generate cli', () => {
        const path = '/testPath/xyz.yaml';
        expect(component.getCli(path)).toBe(`tanzu management-cluster create --file ${path} -v 6`);
    });

    it('should call api to create azure regional cluster', () => {
        const apiSpy = spyOn(component['apiClient'], 'createAzureRegionalCluster').and.callThrough();
        component.createRegionalCluster({});
        expect(apiSpy).toHaveBeenCalled();
    });

    it('should apply TKG config for azure', () => {
        const apiSpy = spyOn(component['apiClient'], 'applyTKGConfigForAzure').and.callThrough();
        component.applyTkgConfig();
        expect(apiSpy).toHaveBeenCalled();
    });
});
