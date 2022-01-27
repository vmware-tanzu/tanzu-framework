// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-management-cluster-api';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { NodeSettingStepComponent } from './node-setting-step.component';
import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { AzureForm } from '../azure-wizard.constants';

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
        component.setInputs('EggplantWizard', AzureForm.NODESETTING,  new FormBuilder().group({}));

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should be invalid when cluster name has leading/trailing spaces', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get('managementClusterName').setValue(" test");
            expect(component.formGroup.valid).toBeFalsy();
            component.formGroup.get('managementClusterName').setValue("test   ");
            expect(component.formGroup.valid).toBeFalsy();
        });
    });

    xit('Dev instance type should be reset if control plane is prod', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get("devInstanceType").setValue("t3.small");
            const cards = fixture.debugElement.queryAll(By.css("a.card"));
            cards[1].triggerEventHandler('click', {});
            expect(component.formGroup.get("devInstanceType").value).toBeFalsy();
        });
    });

    xit('Prod instance type should be reset if control plane is dev', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get("prodInstanceType").setValue("t3.small");
            const cards = fixture.debugElement.queryAll(By.css("a.card"));
            cards[0].triggerEventHandler('click', {});
            expect(component.formGroup.get("prodInstanceType").value).toBeFalsy();
        });
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        component.ngOnInit();
        component.nodeType = '';

        const staticDescription = component.dynamicDescription();
        expect(staticDescription).toEqual('Specify the resources backing the  cluster');

        component.setClusterTypeDescriptor('FUDGE');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'EggplantWizard',
                step: AzureForm.NODESETTING,
                description: 'Specify the resources backing the FUDGE cluster',
            }
        });

        const planeSettingControl = component.formGroup.get('controlPlaneSetting');
        planeSettingControl.setValue('prod');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'EggplantWizard',
                step: AzureForm.NODESETTING,
                description: 'Control plane type: prod',
            }
        });
    });
});
