import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { ResourcePool, ResourceStepComponent } from './resource-step.component';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { DataServiceRegistrarTestExtension } from '../../../../testing/data-service-registrar.testextension';
import { VSphereDatastore, VSphereFolder, VSphereResourcePool } from '../../../../swagger/models';
import { VsphereField } from '../vsphere-wizard.constants';

describe('ResourceStepComponent', () => {
    let component: ResourceStepComponent;
    let fixture: ComponentFixture<ResourceStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                APIClient,
                FormBuilder,
                FieldMapUtilities,
                ValidationService,
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [ResourceStepComponent]
        }).compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        const dataServiceRegistrar = new DataServiceRegistrarTestExtension();
        AppServices.dataServiceRegistrar = dataServiceRegistrar;
        // we expect the wizard to have registered for these events:
        dataServiceRegistrar.simulateRegistration<VSphereResourcePool>(TanzuEventType.VSPHERE_GET_RESOURCE_POOLS);
        dataServiceRegistrar.simulateRegistration<ResourcePool>(TanzuEventType.VSPHERE_GET_COMPUTE_RESOURCE);
        dataServiceRegistrar.simulateRegistration<VSphereDatastore>(TanzuEventType.VSPHERE_GET_DATA_STORES);
        dataServiceRegistrar.simulateRegistration<VSphereFolder>(TanzuEventType.VSPHERE_GET_VM_FOLDERS);

        TestBed.inject(ValidationService);
        fixture = TestBed.createComponent(ResourceStepComponent);
        component = fixture.componentInstance;
        component.setInputs('BozoWizard', 'resourceForm', new FormBuilder().group({}));
        component.setClusterTypeDescriptor('VANILLA');

        fixture.detectChanges();
    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should retrieve resources when load resources: case 1', () => {
        const retrieveRrcSpy = spyOn(component, 'retrieveResourcePools').and.callThrough();
        const retrieveDsSpy = spyOn(component, 'retrieveDatastores').and.callThrough();
        const retrieveVmSpy = spyOn(component, 'retrieveVMFolders').and.callThrough();
        component.loadResourceOptions();
        expect(retrieveRrcSpy).toHaveBeenCalled();
        expect(retrieveDsSpy).toHaveBeenCalled();
        expect(retrieveVmSpy).toHaveBeenCalled();
    });

    it('should retrieve resources when load resources: case 2', () => {
        component.resetFieldsUponDCChange();
        expect(component.formGroup.get('resourcePool').value).toBeFalsy();
        expect(component.formGroup.get('datastore').value).toBeFalsy();
        expect(component.formGroup.get('vmFolder').value).toBeFalsy();
    });

    it('should retrieve resources when load resources: case 3', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.retrieveResourcePools();
        expect(msgSpy).toHaveBeenCalled();
    });

    it('should retrieve ds when load resources', async () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.retrieveDatastores();
        expect(msgSpy).toHaveBeenCalled();
    });

    it('should retrieve vm folders when load resources', async () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.retrieveVMFolders();
        expect(msgSpy).toHaveBeenCalled();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const vmFolderControl = component.formGroup.get(VsphereField.RESOURCE_VMFOLDER);
        const datastoreControl = component.formGroup.get(VsphereField.RESOURCE_DATASTORE);
        const resourcePoolControl = component.formGroup.get(VsphereField.RESOURCE_POOL);

        expect(component.dynamicDescription()).toEqual('Specify the resources for this VANILLA cluster');

        vmFolderControl.setValue('VMFOLDER');
        datastoreControl.setValue('DATASTORE');
        resourcePoolControl.setValue('RESOURCE');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'resourceForm',
                description: 'Resource Pool: RESOURCE, VM Folder: VMFOLDER, Datastore: DATASTORE',
            }
        });
    });
});
