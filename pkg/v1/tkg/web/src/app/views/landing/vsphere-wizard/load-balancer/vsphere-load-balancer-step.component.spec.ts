// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../swagger';
import AppServices from '../../../../shared/service/appServices';
import { DataServiceRegistrarTestExtension } from '../../../../testing/data-service-registrar.testextension';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { KUBE_VIP } from '../../wizard/shared/components/steps/load-balancer/load-balancer-step.component';
import { Messenger, TanzuEventType } from '../../../../shared/service/Messenger';
import { ResourcePool } from '../resource-step/resource-step.component';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { VSphereDatastore, VSphereFolder, VSphereResourcePool } from '../../../../swagger/models';
import { VsphereLoadBalancerStepComponent } from './vsphere-load-balancer-step.component';
import { WizardForm } from '../../wizard/shared/constants/wizard.constants';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

describe('VsphereLoadBalancerStepComponent', () => {
    let component: VsphereLoadBalancerStepComponent;
    let fixture: ComponentFixture<VsphereLoadBalancerStepComponent>;

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
                FieldMapUtilities,
                APIClient
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [VsphereLoadBalancerStepComponent]
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

        fixture = TestBed.createComponent(VsphereLoadBalancerStepComponent);
        component = fixture.componentInstance;
        component.setStepRegistrantData({ wizard: 'BozoWizard', step: WizardForm.LOADBALANCER, formGroup: new FormBuilder().group({}),
            eventFileImported: TkgEventType.VSPHERE_CONFIG_FILE_IMPORTED,
            eventFileImportError: TkgEventType.VSPHERE_CONFIG_FILE_IMPORT_ERROR});

        fixture.detectChanges();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const controllerHostControl = component.formGroup.get(LoadBalancerField.CONTROLLER_HOST);

        expect(component.dynamicDescription()).toEqual(VsphereLoadBalancerStepComponent.description);

        AppServices.messenger.publish({
            type: TanzuEventType.VSPHERE_CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
            payload: KUBE_VIP
        });
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.LOADBALANCER,
                description: VsphereLoadBalancerStepComponent.descriptionKubeVip,
            }
        });

        controllerHostControl.setValue('JAMOCA');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.LOADBALANCER,
                description: 'Controller: JAMOCA',
            }
        });
    });
});
