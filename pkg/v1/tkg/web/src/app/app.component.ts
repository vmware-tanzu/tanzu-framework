// Angular imports
import { Component } from '@angular/core';
import { takeUntil } from 'rxjs/operators';

// App imports
import { BasicSubscriber } from './shared/abstracts/basic-subscriber';
import { APIClient } from './swagger/api-client.service';
import { ProviderInfo } from './swagger/models/provider-info.model';
import { AppDataService } from './shared/service/app-data.service';
import { BrandingService } from './shared/service/branding.service';
import { Features } from "./swagger/models";

@Component({
    selector: 'tkg-kickstart-ui-app',
    templateUrl: './app.component.html',
    styleUrls: ['./app.component.scss']
})
export class AppComponent extends BasicSubscriber {
    providerType: string = null;

    constructor(private apiClient: APIClient,
                private appDataService: AppDataService,
                private editionService: BrandingService) {
        super();

        this.appDataService.setProviderType(null);

        this.editionService.initBranding();

        this.apiClient.getProvider()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(((res: ProviderInfo) => {
                this.providerType = res.provider;
                this.appDataService.setProviderType(res.provider);
                this.appDataService.setTkrVersion(res.tkrVersion);
            }),
            ((err) => {
                console.log('Failed to retrieve provider type and Kubernetes version.');
            })
        );
        this.apiClient.getFeatureFlags()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(((features:Features) => {
                this.appDataService.setFeatureFlags(features);
            }),
            ((err) => {
                console.log('Failed to retrieve feature flags.');
            })
        );
    }
}
