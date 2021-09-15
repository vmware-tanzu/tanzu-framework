// Angular imports
import { Injectable } from '@angular/core';
import { finalize } from 'rxjs/operators'

// Application imports
import { TkgEventType } from 'src/app/shared/service/Messenger';
import { APIClient } from 'src/app/swagger';
import { AppEdition, brandingDefault, brandingTce, brandingTceStandalone } from '../constants/branding.constants';
import Broker from './broker';

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
    clusterType: string;
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
     * @param edition - Optional parameter. 'tce' or 'tce-standalone' to retrieve tce branding; otherwise retrieves
     * default branding.
     */
    private setBrandingByEdition(edition?: string): void {
        let brandingPayload: EditionData = brandingDefault;

        if (edition && edition === AppEdition.TCE) {
            brandingPayload = brandingTce;
        } else if (edition && edition === AppEdition.TCE_STANDALONE) {
            brandingPayload = brandingTceStandalone;
        }

        Broker.messenger.publish({
            type: TkgEventType.BRANDING_CHANGED,
            payload: brandingPayload
        });
    }
}
