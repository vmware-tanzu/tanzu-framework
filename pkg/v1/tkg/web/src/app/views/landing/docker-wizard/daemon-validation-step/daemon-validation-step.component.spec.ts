// Angular modules
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';

// Library imports
import { APIClient } from 'tanzu-management-cluster-ng-api';

// App imports
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';
import { DaemonValidationStepComponent } from './daemon-validation-step.component';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';

describe('DaemonValidationStepComponent', () => {
    let component: DaemonValidationStepComponent;
    let fixture: ComponentFixture<DaemonValidationStepComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes([
                    { path: 'ui', component: DaemonValidationStepComponent }
                ]),
                ReactiveFormsModule,
                SharedModule,
                BrowserAnimationsModule
            ],
            declarations: [ DaemonValidationStepComponent ],
            providers: [
                APIClient,
                FieldMapUtilities,
                ValidationService
            ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(DaemonValidationStepComponent);
        component = fixture.componentInstance;
        component.formGroup = fb.group({
        });
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
