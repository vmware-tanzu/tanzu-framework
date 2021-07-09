import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { HttpClientModule } from '@angular/common/http';

import { APIClient } from '../../swagger/api-client.service';
import { VSphereWizardFormService } from './vsphere-wizard-form.service';

describe('VSphereWizardFormService', () => {
    beforeEach(() => TestBed.configureTestingModule({
        imports: [
            HttpClientTestingModule
        ],
        providers: [
            APIClient
        ]
    }));

    it('should be created', () => {
        const service: VSphereWizardFormService = TestBed.get(VSphereWizardFormService);
        expect(service).toBeTruthy();
    });
});
