// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { AzureForm } from '../azure-wizard.constants';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { NodeSettingField } from '../../wizard/shared/components/steps/node-setting-step/node-setting-step.fieldmapping';
import { NodeSettingStepComponent } from './node-setting-step.component';
import { SharedModule } from '../../../../shared/shared.module';
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
        component.setStepRegistrantData({ wizard: 'EggplantWizard', step: AzureForm.NODESETTING,
            formGroup: new FormBuilder().group({}),
            eventFileImported: TanzuEventType.AZURE_CONFIG_FILE_IMPORTED,
            eventFileImportError: TanzuEventType.AZURE_CONFIG_FILE_IMPORT_ERROR});

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should be invalid when cluster name has leading/trailing spaces', () => {
        fixture.whenStable().then(() => {
            component.formGroup.get(NodeSettingField.CLUSTER_NAME).setValue(" test");
            expect(component.formGroup.valid).toBeFalsy();
            component.formGroup.get(NodeSettingField.CLUSTER_NAME).setValue("test   ");
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
        component.clearClusterPlan();

        const staticDescription = component.dynamicDescription();
        expect(staticDescription).toEqual('Specify the resources backing the  cluster');

        component.setClusterTypeDescriptor('FUDGE');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'EggplantWizard',
                step: AzureForm.NODESETTING,
                description: 'Specify the resources backing the FUDGE cluster',
            }
        });

        component.cardClickProd();
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'EggplantWizard',
                step: AzureForm.NODESETTING,
                description: 'Production cluster selected: 3 node control plane',
            }
        });
    });
});
