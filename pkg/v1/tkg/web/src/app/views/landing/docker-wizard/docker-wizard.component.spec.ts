// Angular modules
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';

// Library imports
import { APIClient, DockerDaemonStatus } from 'tanzu-management-cluster-ng-api';

// App imports
import { SharedModule } from 'src/app/shared/shared.module';
import { DockerWizardComponent } from './docker-wizard.component';

describe('DockerWizardComponent', () => {
    let component: DockerWizardComponent;
    let fixture: ComponentFixture<DockerWizardComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                ReactiveFormsModule,
                BrowserAnimationsModule,
                SharedModule
            ],
            providers: [
                APIClient,
                FormBuilder
            ],
            declarations: [ DockerWizardComponent ]
        })
        .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(DockerWizardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
