import { Injectable } from '@angular/core';

import { Observable, throwError } from 'rxjs';

import { APIClient } from '../../swagger/api-client.service';
import Broker from './broker';
import { TkgEventType } from './Messenger';
import { WizardFormBase } from './wizard-form-base';

const DataSources = [
    TkgEventType.AZURE_GET_RESOURCE_GROUPS,
    TkgEventType.AZURE_GET_INSTANCE_TYPES,
    TkgEventType.AZURE_GET_OS_IMAGES
];
const DataSpec = {
    [TkgEventType.AZURE_GET_RESOURCE_GROUPS]: "getAzureResourceGroups",
    [TkgEventType.AZURE_GET_INSTANCE_TYPES]: "getAzureInstanceTypes",
    [TkgEventType.AZURE_GET_OS_IMAGES]: "getAzureOSImages"
};

const ErrorSpec = {
    [TkgEventType.AZURE_GET_RESOURCE_GROUPS]: "Failed to retrieve resource groups for the particular region.",
    [TkgEventType.AZURE_GET_INSTANCE_TYPES]: "Failed to retrieve Azure VM sizes",
    [TkgEventType.AZURE_GET_OS_IMAGES]: "Failed to retrieve list of OS images from the specified Azure Server."
};

@Injectable({
    providedIn: 'root'
})
export class AzureWizardFormService extends WizardFormBase {
    // Azure globals
    region: string;

    constructor(private apiClient: APIClient) {
        super(DataSources, ErrorSpec);

        // Messenger handler for Azure region change
        Broker.messenger.getSubject(TkgEventType.AZURE_REGION_CHANGED)
            .subscribe(event => {
                this.region = event.payload;
                DataSources.forEach(source => {
                    if (this.region) {
                        // publishing these events will cause a fetch of the event-related data
                        Broker.messenger.publish({
                            type: source
                        });
                    } else {
                        // publishing empty data will clear the event-related data
                        this.publishData(source, []);
                    }
                });
            });
    }

    retrieveDataForSource(source: TkgEventType): Observable<any> {
        const method = DataSpec[source];
        if (!method) {
            return throwError({ message: `Unknown data source ${source}` });
        }
        if (!this.region) {
            const msg = 'WARNING: azure-wizard-form.retrieveDataForSource() received event ' +
                source.toString() + ' but has no region set!';
            console.log(msg);
            return throwError({ message: msg });
        }

        return this.apiClient[method]({location: this.region});
    }
}
