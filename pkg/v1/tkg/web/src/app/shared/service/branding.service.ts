// Angular imports
import { Injectable } from '@angular/core';
import { finalize } from 'rxjs/operators'

// Application imports
import { TanzuEventType } from 'src/app/shared/service/Messenger';
import { APIClient } from 'src/app/swagger';
import {AppEdition, brandingDefault, brandingStandalone, brandingTce} from '../constants/branding.constants';
import AppServices from './appServices';

export interface BrandingObj {
    logoClass: string;
    title: string;
    intro: string;
}

export interface BrandingData {
    title: string;
    landingPage: BrandingObj;
}

export interface EditionData {
    branding: BrandingData;
    clusterTypeDescriptor: string;
    edition: AppEdition;
}

@Injectable({
    providedIn: 'root'
})
export class BrandingService {

    constructor(private apiClient: APIClient) {
    }

    /**
     * @method initBranding
     * Initializes process of retrieving edition flag value and subsequently retrieving branding data using this flag.
     * Note that the caller should have retrieved feature flags before calling this method
     */
    initBranding(): void {
        let brandingEdition;
        this.apiClient.getTanzuEdition().pipe(
            finalize(() =>
                this.setBrandingByEdition(brandingEdition)
            ))
            .subscribe((data) => {
                brandingEdition = data;
            }, (err) => {
                console.log(`Unable to retrieve edition.`)
            });
    }

    /**
     * @method setBrandingByEdition
     * Helper method used to set branding content in Messenger payload depending on which edition is detected.
     * Dispatches 'BRANDING_CHANGED' message with branding data as payload.
     * @param edition - Optional parameter. 'tce' to retrieve tce branding; otherwise retrieves
     * default branding.
     */
    private setBrandingByEdition(edition?: string): void {
        let brandingPayload: EditionData = brandingDefault;

        if (edition && edition === AppEdition.TCE) {
            console.log('Setting branding based on edition: ' + AppEdition.TCE);
            brandingPayload = brandingTce;
        }
        if (AppServices.appDataService.isModeClusterStandalone()) {
            console.log('Due to standalone cluster mode, setting branding to edition: ' + AppEdition.TCE);
            brandingPayload = brandingTce;
            brandingPayload.clusterTypeDescriptor = brandingStandalone.clusterTypeDescriptor;
            brandingPayload.branding.landingPage.intro = brandingStandalone.branding.landingPage.intro;
        }

        AppServices.messenger.publish({
            type: TanzuEventType.BRANDING_CHANGED,
            payload: brandingPayload
        });
    }
}
