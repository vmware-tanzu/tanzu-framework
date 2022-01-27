// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-management-cluster-api';

// App imports
import AppServices from 'src/app/shared/service/appServices';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { MetadataStepComponent } from './metadata-step.component';
import { SharedModule } from '../../../../../../../shared/shared.module';
import { ValidationService } from '../../../validation/validation.service';
import { WizardForm } from '../../../constants/wizard.constants';

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

    it('should add new label', () => {
        component.addLabel("somekey", "someval");
        component.addLabel("somekey2", "someval2");
        expect(component.clusterLabelsValue).toEqual("somekey:someval, somekey2:someval2");
    });

    it('should delete existing label', () => {
        component.addLabel("akey", "avalue");
        expect(component.clusterLabelsValue).toEqual('akey:avalue');
        component.deleteLabel("akey");
        expect(component.clusterLabelsValue).toEqual('');
    });

    it('should announce description change', () => {
        const msgSpy = spyOn(AppServices.messenger, 'publish').and.callThrough();
        const locationControl = component.formGroup.controls['clusterLocation'];

        component.setClusterTypeDescriptor('CLOWN');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.METADATA,
                description: 'Specify metadata for the CLOWN cluster'
            }
        });

        locationControl.setValue('UZBEKISTAN');
        expect(msgSpy).toHaveBeenCalledWith({
            type: TkgEventType.STEP_DESCRIPTION_CHANGE,
            payload: {
                wizard: 'BozoWizard',
                step: WizardForm.METADATA,
                description: 'Location: UZBEKISTAN'
            }
        });
    });
});
