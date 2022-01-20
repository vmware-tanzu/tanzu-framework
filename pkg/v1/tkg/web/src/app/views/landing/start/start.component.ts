// Angular imports
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Title } from '@angular/platform-browser';
// Third party imports
import { Observable } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
// App imports
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import AppServices from 'src/app/shared/service/appServices';
import { BasicSubscriber } from 'src/app/shared/abstracts/basic-subscriber';
import { BrandingObj, EditionData } from '../../../shared/service/branding.service';
import { PROVIDERS, Providers } from '../../../shared/constants/app.constants';
import { TkgEvent, TkgEventType } from 'src/app/shared/service/Messenger';

@Component({
    selector: 'tkg-kickstart-ui-start',
    templateUrl: './start.component.html',
    styleUrls: ['./start.component.scss']
})
export class StartComponent extends BasicSubscriber implements OnInit {
    APP_ROUTES: Routes = APP_ROUTES;
    PROVIDERS: Providers = PROVIDERS;

    edition: string;
    clusterTypeDescriptor: string;
    provider: Observable<string>;
    landingPageContent: BrandingObj;
    loading: boolean = false;

    constructor(private router: Router,
                private titleService: Title) {
        super();
        this.provider = AppServices.appDataService.getProviderType();
    }

    ngOnInit() {
        /**
         * Whenever branding data changes, load content in landing page
         */
        AppServices.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                const content: EditionData = data.payload;
                const title = content.branding.title;
                this.edition = content.edition;
                this.clusterTypeDescriptor = content.clusterTypeDescriptor;
                this.landingPageContent = content.branding.landingPage;
                this.titleService.setTitle(title);
            });
    }

    /**
     * @method navigateToWizard
     * @desc helper method to trigger router navigation to wizard
     * @param provider - the provider to load wizard for
     */
    navigateToWizard(provider: string): void {
        this.loading = true;
        AppServices.appDataService.setProviderType(provider);
        let wizard;
        switch (provider) {
            case PROVIDERS.VSPHERE: {
                wizard = APP_ROUTES.WIZARD_MGMT_CLUSTER;
                break;
            }
            case PROVIDERS.AWS: {
                wizard = APP_ROUTES.AWS_WIZARD;
                break;
            }
            case PROVIDERS.AZURE: {
                wizard = APP_ROUTES.AZURE_WIZARD;
                break;
            }
            case PROVIDERS.DOCKER: {
                wizard = APP_ROUTES.DOCKER_WIZARD;
                break;
            }
        }
        this.router.navigate([wizard]);
    }
}
