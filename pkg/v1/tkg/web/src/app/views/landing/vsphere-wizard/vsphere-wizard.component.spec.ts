// Angular imports
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormControl } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
// Third party imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
// App imports
import { APIClient } from '../../../swagger/api-client.service';
import AppServices from '../../../shared/service/appServices';
import { ClusterType } from "../wizard/shared/constants/wizard.constants";
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { SharedModule } from '../../../shared/shared.module';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { VSphereWizardComponent } from './vsphere-wizard.component';

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

    it('getWizardValidity should return false', () => {
        expect(component['getWizardValidity']()).toBeFalsy();
    });

    it('should call create vsphere api when deploying', () => {
        const apiSpy = spyOn(component['apiClient'], 'createVSphereRegionalCluster').and.callThrough();
        component.providerType = 'vsphere';
        component.deploy();
        expect(apiSpy).toHaveBeenCalled();
    });

    it('should register services', () => {
        const apiSpy = spyOn(AppServices.dataServiceRegistrar, 'register').and.callThrough();
        component.ngOnInit();
        expect(apiSpy).toHaveBeenCalledWith(TanzuEventType.VSPHERE_GET_COMPUTE_RESOURCE, jasmine.anything(), jasmine.anything());
        expect(apiSpy).toHaveBeenCalledWith(TanzuEventType.VSPHERE_GET_DATA_STORES, jasmine.anything(), jasmine.anything());
        expect(apiSpy).toHaveBeenCalledWith(TanzuEventType.VSPHERE_GET_OS_IMAGES, jasmine.anything(), jasmine.anything());
        expect(apiSpy).toHaveBeenCalledWith(TanzuEventType.VSPHERE_GET_RESOURCE_POOLS, jasmine.anything(), jasmine.anything());
        expect(apiSpy).toHaveBeenCalledWith(TanzuEventType.VSPHERE_GET_VM_FOLDERS, jasmine.anything(), jasmine.anything());
        expect(apiSpy).toHaveBeenCalledWith(TanzuEventType.VSPHERE_GET_VM_NETWORKS, jasmine.anything(), jasmine.anything());
    });
});
