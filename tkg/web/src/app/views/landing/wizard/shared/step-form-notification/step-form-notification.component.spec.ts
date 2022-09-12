import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { StepFormNotificationComponent } from './step-form-notification.component';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { SharedModule } from 'src/app/shared/shared.module';
import { ValidationService } from '../validation/validation.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';

describe('StepFormNotificationComponent', () => {
    let component: StepFormNotificationComponent;
    let fixture: ComponentFixture<StepFormNotificationComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [StepFormNotificationComponent],
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                ValidationService,
                FormBuilder
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(StepFormNotificationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
