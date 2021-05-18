// Angular imports
import { Injectable } from '@angular/core';
import { finalize } from 'rxjs/operators'

// Application imports
import { Messenger, TkgEventType } from 'src/app/shared/service/Messenger';
import { APIClient } from 'src/app/swagger';
import { brandingDefault, brandingTce } from '../constants/branding.constants';

export interface BrandingObj {
    logoClass: string;
    title: string;
    intro: string;
}

export interface BrandingData {
    landingPage: BrandingObj;
}

export interface EditionData {
    branding: BrandingData;
    edition: string;
}

@Injectable({
    providedIn: 'root'
})
export class BrandingService {

    constructor(private apiClient: APIClient,
                private messenger: Messenger) {
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
     * @param edition - Optional parameter. 'tce' to retrieve tce branding; otherwise retrieves default branding.
     */
    private setBrandingByEdition(edition?: string): void {
        let brandingPayload: EditionData = brandingDefault;

        if (edition && edition !== 'tkg') {
            brandingPayload = brandingTce;
        }

        this.messenger.publish({
            type: TkgEventType.BRANDING_CHANGED,
            payload: brandingPayload
        });
    }
}
