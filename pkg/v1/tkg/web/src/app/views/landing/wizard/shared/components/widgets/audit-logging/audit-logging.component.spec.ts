// Angular modules
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

// Library imports
import { APIClient } from 'tanzu-ui-api-lib';

// App imports
import { SharedModule } from 'src/app/shared/shared.module';
import { AuditLoggingComponent } from './audit-logging.component';
import { FieldMapUtilities } from '../../../field-mapping/FieldMapUtilities';

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
                FieldMapUtilities,
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
