// Angular imports
import {async, ComponentFixture, TestBed} from '@angular/core/testing';
import {CUSTOM_ELEMENTS_SCHEMA} from '@angular/core';
import {FormArray, FormBuilder, ReactiveFormsModule} from '@angular/forms';
// App imports
import {APIClient} from '../../../../../../../swagger/api-client.service';
import AppServices from 'src/app/shared/service/appServices';
import {Messenger, TanzuEventType} from 'src/app/shared/service/Messenger';
import {MetadataStepComponent} from './metadata-step.component';
import {SharedModule} from '../../../../../../../shared/shared.module';
import {ValidationService} from '../../../validation/validation.service';
import {WizardForm} from '../../../constants/wizard.constants';
import {MetadataField} from './metadata-step.fieldmapping';

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
        // NOTE: using Azure file import events just for testing
        component.setStepRegistrantData({
            wizard: 'BozoWizard', step: WizardForm.METADATA, formGroup: new FormBuilder().group({}),
            eventFileImported: TanzuEventType.AZURE_CONFIG_FILE_IMPORTED,
            eventFileImportError: TanzuEventType.AZURE_CONFIG_FILE_IMPORT_ERROR
        });

        component.ngOnInit();

        fixture.detectChanges();
    });

    it('should initialize tkgLabelsConfig', () => {
        component.setClusterTypeDescriptor('Management');
        component.ngOnInit();
        const config = component.tkgLabelsConfig;

        expect(config.label.title).toEqual('LABELS (OPTIONAL)');
        expect(config.forms.parent.get('clusterLabels')).toBeInstanceOf(FormArray);
        expect(config.fields.clusterTypeDescriptor).toEqual('Management');
    })

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
