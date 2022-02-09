// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../swagger';
import AppServices from '../../../../shared/service/appServices';
import { Messenger, TanzuEventType } from '../../../../shared/service/Messenger';
import { DataServiceRegistrarTestExtension } from '../../../../testing/data-service-registrar.testextension';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { NetworkField } from '../../wizard/shared/components/steps/network-step/network-step.fieldmapping';
import { ResourcePool } from '../resource-step/resource-step.component';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereDatastore, VSphereFolder, VSphereNetwork, VSphereResourcePool } from '../../../../swagger/models';
import { VsphereNetworkStepComponent } from './vsphere-network-step.component';
import { WizardForm } from '../../wizard/shared/constants/wizard.constants';

describe('NodeSettingStepComponent', () => {
    let component: VsphereNetworkStepComponent;
    let fixture: ComponentFixture<VsphereNetworkStepComponent>;

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
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [VsphereNetworkStepComponent]
        })
            .compileComponents();
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
        dataServiceRegistrar.simulateRegistration<VSphereFolder>(TanzuEventType.VSPHERE_GET_VM_NETWORKS);

        fixture = TestBed.createComponent(VsphereNetworkStepComponent);
        component = fixture.componentInstance;
        component.setStepRegistrantData({ wizard: 'BozoWizard', step: WizardForm.NETWORK, formGroup: new FormBuilder().group({}),
            eventFileImported: TanzuEventType.VSPHERE_CONFIG_FILE_IMPORTED,
            eventFileImportError: TanzuEventType.VSPHERE_CONFIG_FILE_IMPORT_ERROR});

        fixture.detectChanges();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const networkNameControl = component.formGroup.controls[NetworkField.NETWORK_NAME];
        const serviceCidrControl = component.formGroup.controls[NetworkField.CLUSTER_SERVICE_CIDR];
        const podCidrControl = component.formGroup.controls[NetworkField.CLUSTER_POD_CIDR];

        podCidrControl.setValue('');
        serviceCidrControl.setValue('');
        expect(component.dynamicDescription()).toEqual(VsphereNetworkStepComponent.description);

        networkNameControl.setValue({name: 'CHOCMINT', moid: 'CHIP'} as VSphereNetwork);
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.NETWORK,
                description: 'Network: CHOCMINT',
            }
        });

        podCidrControl.setValue('1.2.3.4/12');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.NETWORK,
                description: 'Network: CHOCMINT, Cluster Pod CIDR: 1.2.3.4/12'
            }
        });

        serviceCidrControl.setValue('5.6.7.8/16');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.NETWORK,
                description: 'Network: CHOCMINT, Cluster Service CIDR: 5.6.7.8/16, Cluster Pod CIDR: 1.2.3.4/12'
            }
        });
    });
});
