import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { SharedModule } from 'src/app/shared/shared.module';
import { APIClient } from 'src/app/swagger';

import { AuditLoggingComponent } from './audit-logging.component';

describe('AuditLoggingComponent', () => {
    let component: AuditLoggingComponent;
    let fixture: ComponentFixture<AuditLoggingComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ReactiveFormsModule,
                SharedModule
            ],
            providers: [
                FormBuilder,
                APIClient
            ],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            declarations: [AuditLoggingComponent]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        const fb = new FormBuilder();
        fixture = TestBed.createComponent(AuditLoggingComponent);
        component = fixture.componentInstance;
        component.formName = "test";
        component.formGroup = fb.group({
        });

        fixture.detectChanges();
    });

});
