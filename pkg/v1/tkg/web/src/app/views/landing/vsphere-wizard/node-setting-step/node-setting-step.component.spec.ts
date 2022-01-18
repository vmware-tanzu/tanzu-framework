// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-ui-api-lib';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { SharedModule } from '../../../../shared/shared.module';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
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
                FieldMapUtilities,
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
        component.setInputs('BozoWizard', 'vsphereNodeSettingForm', new FormBuilder().group({}));

        fixture.detectChanges();
    });

    it('should set correct value for card clicking', () => {
        component.cardClick('prod');
        expect(component.formGroup.controls['controlPlaneSetting'].value).toBe('prod')
    });

    it('should get correct env value', () => {
        component.cardClick('prod');
        expect(component.getEnvType()).toEqual('prod');
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        const controlPlaneSettingControl = component.formGroup.get('controlPlaneSetting');

        expect(component.dynamicDescription()).toEqual('Specify the resources backing the  cluster');

        component.setClusterTypeDescriptor('VANILLA');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereNodeSettingForm',
                description: 'Specify the resources backing the VANILLA cluster',
            }
        });

        controlPlaneSettingControl.setValue('dev');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereNodeSettingForm',
                description: 'Development cluster selected: 1 node control plane'
            }
        });

        controlPlaneSettingControl.setValue('prod');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: 'vsphereNodeSettingForm',
                description: 'Production cluster selected: 3 node control plane'
            }
        });
    });
});
