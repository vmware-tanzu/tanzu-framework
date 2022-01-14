import { TestBed, async } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { HttpClientModule } from '@angular/common/http';

import { APIClient } from 'tanzu-ui-api-lib';
import { AppDataService } from 'src/app/shared/service/app-data.service';

describe('AppDataService', () => {
    let service: AppDataService;

    beforeEach(() => TestBed.configureTestingModule({
        imports: [
            HttpClientTestingModule
        ],
        providers: [
            APIClient
        ]
    }));

    beforeEach(() => {
        service = TestBed.get(AppDataService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should provide getter/setter for providerType', async(() => {
        const provider = service.getProviderType();

        service.setProviderType('aws');

        provider.subscribe(prov => {
            expect(prov).toBe('aws');
        })
    }));

    it('should provide getter/setter for hasPacificCluster', async(() => {
        const isPacific = service.getIsProjPacific();

        service.setIsProjPacific(true);

        isPacific.subscribe(pacific => {
            expect(pacific).toBe(true);
        })
    }));

    it('should provide getter/setter for tkrVersion', async(() => {
        const tkrVersion = service.getTkrVersion();

        service.setTkrVersion('1.17.3');

        tkrVersion.subscribe(ver => {
            expect(ver).toBe('1.17.3');
        })
    }));
});
