// Angular imports
import { Component } from '@angular/core';
import { takeUntil } from 'rxjs/operators';

// App imports
import { BasicSubscriber } from './shared/abstracts/basic-subscriber';
import { APIClient } from './swagger/api-client.service';
import { ProviderInfo } from './swagger/models/provider-info.model';
import { BrandingService } from './shared/service/branding.service';
import { Features } from "./swagger/models";
import Broker from "./shared/service/broker";

@Component({
    selector: 'tkg-kickstart-ui-app',
    templateUrl: './app.component.html',
    styleUrls: ['./app.component.scss']
})
export class AppComponent extends BasicSubscriber {
    providerType: string = null;

    constructor(private apiClient: APIClient,
                private editionService: BrandingService) {
        super();

        Broker.appDataService.setProviderType(null);

        this.apiClient.getProvider()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(((res: ProviderInfo) => {
                this.providerType = res.provider;
                Broker.appDataService.setProviderType(res.provider);
                Broker.appDataService.setTkrVersion(res.tkrVersion);
            }),
            ((err) => {
                console.log('Failed to retrieve provider type and Kubernetes version.');
            })
        );

        this.apiClient.getFeatureFlags()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(((features: Features) => {
                Broker.appDataService.setFeatureFlags(features);
                this.editionService.initBranding(); // NOTE: the branding may depend on feature flags
            }),
            ((err) => {
                console.log('Failed to retrieve feature flags.');
            })
        );
    }
}
