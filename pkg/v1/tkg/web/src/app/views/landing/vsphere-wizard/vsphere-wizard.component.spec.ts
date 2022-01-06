// Angular imports
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormControl } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
// Third party imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
// App imports
import { APIClient } from '../../../swagger/api-client.service';
import { ClusterType } from "../wizard/shared/constants/wizard.constants";
import { FieldMapUtilities } from '../wizard/shared/field-mapping/FieldMapUtilities';
import { FormMetaDataStore } from '../wizard/shared/FormMetaDataStore';
import { Messenger } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../shared/shared.module';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { VSphereProviderStepComponent } from './provider-step/vsphere-provider-step.component';
import { VSphereWizardComponent } from './vsphere-wizard.component';
import AppServices from '../../../shared/service/appServices';

describe('VSphereWizardComponent', () => {
    let component: VSphereWizardComponent;
    let fixture: ComponentFixture<VSphereWizardComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                ReactiveFormsModule,
                BrowserAnimationsModule,
                RouterTestingModule.withRoutes([
                    { path: 'ui', component: VSphereWizardComponent }
                ]),
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
            declarations: [
                VSphereWizardComponent
            ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(VSphereWizardComponent);
        component = fixture.componentInstance;
        component.form = fb.group({
            vsphereProviderForm: fb.group({}),
            vsphereNodeSettingForm: fb.group({}),
            metadataForm: fb.group({}),
            storageForm: fb.group({}),
            resourceForm: fb.group({}),
            loadBalancerForm: fb.group({}),
            networkForm: fb.group({}),
            identityForm: fb.group({}),
            osImageForm: fb.group({}),
            ceipOptInForm: fb.group({})
        });
        component.clusterTypeDescriptor = '' + ClusterType.Management;
        fixture.detectChanges();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should call getStepList in reviewConfiguration', () => {
        const getStepListSpy = spyOn(FormMetaDataStore, 'getStepList').and.callThrough();
        component.getWizardValidity();
        expect(getStepListSpy).toHaveBeenCalled();
    });

    it('getWizardValidity should return false when getStepList is empty', () => {
        expect(component['getWizardValidity']()).toBeFalsy();
    });

    it('VsphereProviderFormDescription should return correct static description when wizard is not filled', () => {
        const description = component.describeStep('vsphereProviderForm', component.VsphereProviderForm.description)
        expect(description).toBe('Validate the vSphere provider account for Tanzu');
    });

    it('VsphereProviderFormDescription should return correct dynamic summary for wizard input', () => {
        const fb = new FormBuilder();
        component.form.controls['vsphereProviderForm'] = fb.group({
            vcenterAddress: new FormControl('vcAddr'),
            datacenter: new FormControl('dc'),
        });
        component.clusterTypeDescriptor = 'management';
        const providerStep = TestBed.createComponent(VSphereProviderStepComponent).componentInstance;
        providerStep.clusterTypeDescriptor = 'management';
        component.registerStep('vsphereProviderForm', providerStep);

        const description = component.describeStep('vsphereProviderForm', component.VsphereProviderForm.description)
        expect(description).toBe('vCenter vcAddr connected');
    });

    it('should call create vsphere api when deploying', () => {
        const apiSpy = spyOn(component['apiClient'], 'createVSphereRegionalCluster').and.callThrough();
        component.providerType = 'vsphere';
        component.deploy();
        expect(apiSpy).toHaveBeenCalled();
    });
});
