// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';

// Library imports
import { APIClient } from 'tanzu-mgmt-plugin-api-lib';

// App imports
import AppServices from '../../../shared/service/appServices';
import { AzureForm } from './azure-wizard.constants';
import { AzureProviderStepComponent } from './provider-step/azure-provider-step.component';
import { AzureWizardComponent } from './azure-wizard.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ClusterType, WizardForm } from '../wizard/shared/constants/wizard.constants';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FieldMapUtilities } from '../wizard/shared/field-mapping/FieldMapUtilities';
import { FormBuilder, FormControl, ReactiveFormsModule } from '@angular/forms';
import { Messenger } from 'src/app/shared/service/Messenger';
import { MetadataStepComponent } from '../wizard/shared/components/steps/metadata-step/metadata-step.component';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { RouterTestingModule } from '@angular/router/testing';
import { SharedModule } from 'src/app/shared/shared.module';
import { SharedNetworkStepComponent } from '../wizard/shared/components/steps/network-step/network-step.component';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { VnetStepComponent } from './vnet-step/vnet-step.component';

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
                FieldMapUtilities,
                ValidationService
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],

            declarations: [AzureWizardComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
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
