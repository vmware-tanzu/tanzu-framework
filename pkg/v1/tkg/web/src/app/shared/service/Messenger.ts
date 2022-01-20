// Angular imports
import { Injectable } from '@angular/core';
// Third party imports
import { ReplaySubject, Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

/**
 * Types of Event being supported by this broker
 */

export enum TanzuEventType {
    // vSphere events
    VSPHERE_CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
    VSPHERE_DATACENTER_CHANGED,
    DATACENTER_RESET,
    VSPHERE_CONFIG_FILE_IMPORTED,
    VSPHERE_CONFIG_FILE_IMPORT_ERROR,
    VSPHERE_GET_COMPUTE_RESOURCE,
    VSPHERE_GET_DATA_STORES,
    VSPHERE_GET_OS_IMAGES,
    VSPHERE_GET_RESOURCE_POOLS,
    VSPHERE_GET_VM_NETWORKS,
    VSPHERE_GET_VM_FOLDERS,
    VSPHERE_IP_FAMILY_CHANGE,
    VSPHERE_VC_AUTHENTICATED,

    // AWS events
    AWS_AIRGAPPED_VPC_CHANGE,
    AWS_CONFIG_FILE_IMPORTED,
    AWS_CONFIG_FILE_IMPORT_ERROR,
    AWS_GET_EXISTING_VPCS,
    AWS_GET_AVAILABILITY_ZONES,
    AWS_GET_SUBNETS,
    AWS_GET_NODE_TYPES,
    AWS_GET_OS_IMAGES,
    AWS_REGION_CHANGED,
    AWS_VPC_TYPE_CHANGED,
    AWS_VPC_CHANGED,

    // AZURE events
    AZURE_CONFIG_FILE_IMPORTED,
    AZURE_CONFIG_FILE_IMPORT_ERROR,
    AZURE_GET_RESOURCE_GROUPS,
    AZURE_GET_VNETS,
    AZURE_GET_INSTANCE_TYPES,
    AZURE_GET_OS_IMAGES,
    AZURE_REGION_CHANGED,
    AZURE_RESOURCEGROUP_CHANGED,

    // Docker events
    DOCKER_CONFIG_FILE_IMPORTED,
    DOCKER_CONFIG_FILE_IMPORT_ERROR,

    // Common provider events
    NETWORK_STEP_GET_NO_PROXY_INFO,

    // CLI
    CLI_CHANGED,

    // APP
    BRANDING_CHANGED,
    STEP_COMPLETED,
    STEP_DESCRIPTION_CHANGE,
}

// The payload structure expected on a STEP_DESCRIPTION_CHANGE event
export interface StepDescriptionChangePayload {
    wizard: string,
    step: string,
    description: string,
}
// The payload structure expected on a STEP_COMPLETED event
export interface StepCompletedPayload {
    wizard: string,
    step: string,
}

/**
 * Event type definition
 */
export interface TanzuEvent<PAYLOAD> {
    type: TanzuEventType,
    payload?: PAYLOAD;
}

/**
 * An extremely light-weight message broker implementation. The purpose of
 * this class is to replace Angular's @Input() and @Output() wiring
 * between components which can be cumbersome and verbose in cases
 * where a good amount of communication is needed.
 */
@Injectable({
    providedIn: 'root'
})
 export class Messenger {
    subjects = new Map<TanzuEventType, ReplaySubject<TanzuEvent<any>>>();

    subscribe<PAYLOAD>(eventType: TanzuEventType, onNext: (event: TanzuEvent<PAYLOAD>) => void, unsubscriber?: Subject<void>) {
        if (unsubscriber) {
            this.getSubject<PAYLOAD>(eventType)
                .pipe(takeUntil(unsubscriber))
                .subscribe(onNext);
        } else {
            this.getSubject<PAYLOAD>(eventType).subscribe(onNext);
        }
    }

    /**
     * Publish an event to all its subscribers
     * @param event the Event to be published
     */
    publish<PAYLOAD>(event: TanzuEvent<PAYLOAD>) {
        this.getSubject<PAYLOAD>(event.type).next(event);
    }

    /**
     * Clears specified event from the Messenger event map. Once this is done
     * subscribers will no longer receive this event until the event is re-dispatched.
     * @param eventType the event to delete from the Messenger buffer
     */
    clearEvent(eventType: TanzuEventType) {
        this.subjects.delete(eventType);
    }

    /**
     * Reset/Clear the Messenger ReplaySubject event map in the rare use cases where
     * we want to force ALL events to be purged from the buffer.
     * This should be used only when necessary.
     */
    reset<PAYLOAD>() {
        this.subjects = new Map<TanzuEventType, ReplaySubject<TanzuEvent<PAYLOAD>>>();
    }
    /**
     * Return the subject based on event type
     * @param eventType event type to get the subject for
     */
    private getSubject<PAYLOAD>(eventType: TanzuEventType) {
        let subject = this.subjects.get(eventType);
        if (!subject) {
            subject = new ReplaySubject<TanzuEvent<PAYLOAD>>(1);
            this.subjects.set(eventType, subject);
        }
        return subject;
    }
}
