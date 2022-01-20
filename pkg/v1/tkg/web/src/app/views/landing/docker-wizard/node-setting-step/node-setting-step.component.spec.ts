import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

import { NodeSettingStepComponent } from './node-setting-step.component';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';

describe('NodeSettingStepComponent', () => {
    let component: NodeSettingStepComponent;
    let fixture: ComponentFixture<NodeSettingStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ NodeSettingStepComponent ],
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                ValidationService,
                FormBuilder,
                FieldMapUtilities,
            ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        const fb = new FormBuilder();

        fixture = TestBed.createComponent(NodeSettingStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
            clusterName: 'testClusterName'
        });
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should have a cluster name', () => {
        expect(component.formGroup.get('clusterName').value).toBe('testClusterName');
    });
});
