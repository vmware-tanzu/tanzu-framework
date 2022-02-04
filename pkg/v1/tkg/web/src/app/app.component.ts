// Angular imports
import { Component } from '@angular/core';
import { takeUntil } from 'rxjs/operators';

// App imports
import { APIClient } from 'tanzu-management-cluster-ng-api';
import { Features, ProviderInfo } from 'tanzu-management-cluster-ng-api';
import { BasicSubscriber } from './shared/abstracts/basic-subscriber';
import { BrandingService } from './shared/service/branding.service';
import AppServices from "./shared/service/appServices";

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

        AppServices.appDataService.setProviderType(null);

        this.apiClient.getProvider()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(((res: ProviderInfo) => {
                this.providerType = res.provider;
                AppServices.appDataService.setProviderType(res.provider);
                AppServices.appDataService.setTkrVersion(res.tkrVersion);
            }),
            ((err) => {
                console.log('Failed to retrieve provider type and Kubernetes version.');
            })
        );

        this.apiClient.getFeatureFlags()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe(((features: Features) => {
                AppServices.appDataService.setFeatureFlags(features);
                this.editionService.initBranding(); // NOTE: the branding may depend on feature flags
            }),
            ((err) => {
                console.log('Failed to retrieve feature flags.');
            })
        );
    }
}
