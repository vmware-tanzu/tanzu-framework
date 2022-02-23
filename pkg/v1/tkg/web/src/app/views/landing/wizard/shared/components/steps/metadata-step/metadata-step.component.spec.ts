// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';
// App imports
import { APIClient } from '../../../../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger, TanzuEventType } from 'src/app/shared/service/Messenger';
import { MetadataStepComponent } from './metadata-step.component';
import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';
import { WizardForm } from '../../../constants/wizard.constants';
import { MetadataField } from './metadata-step.fieldmapping';

describe('MetadataStepComponent', () => {
    let component: MetadataStepComponent;
    let fixture: ComponentFixture<MetadataStepComponent>;

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
            declarations: [MetadataStepComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        AppServices.messenger = new Messenger();
        fixture = TestBed.createComponent(MetadataStepComponent);
        component = fixture.componentInstance;
        component.setInputs('BozoWizard', WizardForm.METADATA, new FormBuilder().group({}));
        component.ngOnInit();

        fixture.detectChanges();
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        const locationControl = component.formGroup.controls[MetadataField.CLUSTER_LOCATION];

        component.setClusterTypeDescriptor('CLOWN');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.METADATA,
                description: 'Specify metadata for the CLOWN cluster'
            }
        });

        locationControl.setValue('UZBEKISTAN');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TanzuEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.METADATA,
                description: 'Location: UZBEKISTAN'
            }
        });
    });
});
