import { Injectable } from '@angular/core';

import { BehaviorSubject, Observable, of, throwError } from 'rxjs';

import { APIClient } from '../../swagger/api-client.service';
import Broker from './broker';
import { TkgEventType } from './Messenger';
import { WizardFormBase } from './wizard-form-base';

const DataSources = [
    TkgEventType.GET_RESOURCE_POOLS,
    TkgEventType.GET_COMPUTE_RESOURCE,
    TkgEventType.VSPHERE_GET_VM_NETWORKS,
    TkgEventType.GET_DATA_STORES,
    TkgEventType.GET_VM_FOLDERS,
    TkgEventType.VSPHERE_GET_OS_IMAGES
];

// DataSpec refers to APIClient method names for HTTP GET operations
const DataSpec = {
    // vSphere events
    [TkgEventType.GET_RESOURCE_POOLS]: "getVSphereResourcePools",
    [TkgEventType.GET_COMPUTE_RESOURCE]: "getVSphereComputeResources",
    [TkgEventType.VSPHERE_GET_VM_NETWORKS]: "getVSphereNetworks",
    [TkgEventType.GET_DATA_STORES]: "getVSphereDatastores",
    [TkgEventType.GET_VM_FOLDERS]: "getVSphereFolders",
    [TkgEventType.VSPHERE_GET_OS_IMAGES]: "getVSphereOSImages"
};

const ErrorSpec = {
    // vSphere events
    [TkgEventType.GET_RESOURCE_POOLS]: "Failed to retrieve list of resource pools from the specified vCenter Server.",
    [TkgEventType.GET_COMPUTE_RESOURCE]: "Failed to retrieve list of compute resources from the specified datacenter.",
    [TkgEventType.VSPHERE_GET_VM_NETWORKS]: "Failed to retrieve list of VM networks from the specified vCenter Server.",
    [TkgEventType.GET_DATA_STORES]: "Failed to retrieve list of datastores from the specified vCenter Server.",
    [TkgEventType.GET_VM_FOLDERS]: "Failed to retrieve list of vm folders from the specified vCenter Server.",
    [TkgEventType.VSPHERE_GET_OS_IMAGES]: "Failed to retrieve list of OS images from the specified vCenter Server."
};

@Injectable({
    providedIn: 'root'
})
export class VSphereWizardFormService extends WizardFormBase {
    formData;

    // vsphere globals
    datacenterMoid: string;

    private vSphereDatacenterMoid = new BehaviorSubject<string | null>(null);

    constructor(private apiClient: APIClient) {
        super(DataSources, ErrorSpec);
        this.vSphereDatacenterMoid.subscribe((moid) => {
            this.datacenterMoid = moid;
        });

        // Messenger handlers
        Broker.messenger.getSubject(TkgEventType.VSPHERE_DATACENTER_CHANGED)
            .subscribe(event => {
                this.datacenterMoid = event.payload;
                DataSources.forEach(source => {
                    Broker.messenger.publish({
                        type: source
                    });
                });
            });
    }

    retrieveDataForSource(source: TkgEventType): Observable<any> {
        const method = DataSpec[source];
        if (!method) {
            return throwError({ message: `Unknown data source ${source}` });
        }
        if (this.datacenterMoid) {
            return this.apiClient[method]({ dc: this.datacenterMoid });
        }

        return of([]);
    }
}
