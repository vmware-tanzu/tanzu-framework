// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SharedModule } from '../../../../shared/shared.module';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { NodeSettingStepComponent } from './node-setting-step.component';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

describe('NodeSettingStepComponent', () => {
    let component: NodeSettingStepComponent;
    let fixture: ComponentFixture<NodeSettingStepComponent>;

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
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [NodeSettingStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();

        fixture = TestBed.createComponent(NodeSettingStepComponent);
        component = fixture.componentInstance;
        component.setStepRegistrantData({ wizard: 'BozoWizard', step: 'vsphereNodeSettingForm', formGroup: new FormBuilder().group({}),
            eventFileImported: TanzuEventType.VSPHERE_CONFIG_FILE_IMPORTED,
            eventFileImportError: TanzuEventType.VSPHERE_CONFIG_FILE_IMPORT_ERROR});

        fixture.detectChanges();
    });

    it('should get correct env value', () => {
        component.cardClickProd();
        expect(component.isClusterPlanProd).toBeTrue();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();

        expect(component.dynamicDescription()).toEqual('Specify the resources backing the  cluster');

        component.setClusterTypeDescriptor('VANILLA');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereNodeSettingForm',
                description: 'Specify the resources backing the VANILLA cluster',
            }
        });

        component.cardClickDev();
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereNodeSettingForm',
                description: 'Development cluster selected: 1 node control plane'
            }
        });

        component.cardClickProd();
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereNodeSettingForm',
                description: 'Production cluster selected: 3 node control plane'
            }
        });
    });
});
