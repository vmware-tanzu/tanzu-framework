// Angular imports
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';
// App imports
import { APIClient } from 'src/app/swagger/api-client.service';
import { DaemonValidationStepComponent } from './daemon-validation-step.component';
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../../wizard/shared/validation/validation.service';

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
