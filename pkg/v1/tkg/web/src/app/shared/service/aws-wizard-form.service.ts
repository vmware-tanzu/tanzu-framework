import { Injectable } from '@angular/core';

import { Observable, throwError } from 'rxjs';

import { APIClient } from '../../swagger/api-client.service';
import { Messenger, TkgEventType } from './Messenger';
import { WizardFormBase } from './wizard-form-base';

const DataSources = [
    TkgEventType.AWS_GET_EXISTING_VPCS,
    TkgEventType.AWS_GET_AVAILABILITY_ZONES,
    TkgEventType.AWS_GET_SUBNETS,
    TkgEventType.AWS_GET_NODE_TYPES,
    TkgEventType.AWS_GET_OS_IMAGES
];
const DataSpec = {
    [TkgEventType.AWS_GET_EXISTING_VPCS]: "getVPCs",
    [TkgEventType.AWS_GET_AVAILABILITY_ZONES]: "getAWSAvailabilityZones",
    [TkgEventType.AWS_GET_SUBNETS]: "getAWSSubnets",
    [TkgEventType.AWS_GET_NODE_TYPES]: "getAWSNodeTypes",
    [TkgEventType.AWS_GET_OS_IMAGES]: "getAWSOSImages"
};

const ErrorSpec = {
    [TkgEventType.AWS_GET_EXISTING_VPCS]: "Failed to retrieve list of existing VPCs from the specified AWS Account.",
    [TkgEventType.AWS_GET_AVAILABILITY_ZONES]: "Failed to retrieve list of availability zones from the specified AWS Account.",
    [TkgEventType.AWS_GET_SUBNETS]: "Failed to retrieve list of VPC subnets from the specified AWS Account.",
    [TkgEventType.AWS_GET_NODE_TYPES]: "Failed to retrieve list of node types from the specified AWS Account.",
    [TkgEventType.AWS_GET_OS_IMAGES]: "Failed to retrieve list of OS images from the specified AWS Server."
};

@Injectable({
    providedIn: 'root'
})
export class AwsWizardFormService extends WizardFormBase {
    // aws globals
    region: string;

    constructor(private apiClient: APIClient, messenger: Messenger) {
        super(DataSources, ErrorSpec, messenger);

        // Messenger handler for AWS region change
        this.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .subscribe(event => {
                this.region = event.payload;
            });
    }

    retrieveDataForSource(source: TkgEventType, payload?: any): Observable<any> {
        const method = DataSpec[source];
        if (!method) {
            return throwError({ message: `Unknown data source ${source}` });
        }

        return this.apiClient[method]({region: this.region, ...payload});
    }
}
