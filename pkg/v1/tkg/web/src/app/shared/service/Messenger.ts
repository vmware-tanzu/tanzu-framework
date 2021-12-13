import { ReplaySubject } from 'rxjs';
import { Injectable } from '@angular/core';

/**
 * Types of Event being supported by this broker
 */

export enum TkgEventType {
    // vSphere events
    VC_AUTHENTICATED,
    DATACENTER_RESET,
    DATACENTER_CHANGED,
    GET_RESOURCE_POOLS,
    GET_COMPUTE_RESOURCE,
    GET_VM_NETWORKS,
    GET_DATA_STORES,
    GET_VM_FOLDERS,
    VSPHERE_GET_OS_IMAGES,
    CONTROL_PLANE_ENDPOINT_PROVIDER_CHANGED,
    IP_FAMILY_CHANGE,

    // AWS events
    AWS_REGION_CHANGED,
    AWS_VPC_TYPE_CHANGED,
    AWS_VPC_CHANGED,
    AWS_GET_EXISTING_VPCS,
    AWS_GET_AVAILABILITY_ZONES,
    AWS_GET_SUBNETS,
    AWS_GET_NODE_TYPES,
    AWS_GET_OS_IMAGES,
    AWS_AIRGAPPED_VPC_CHANGE,

    // AZURE events
    AZURE_REGION_CHANGED,
    AZURE_RESOURCEGROUP_CHANGED,
    AZURE_GET_RESOURCE_GROUPS,
    AZURE_GET_VNETS,
    AZURE_GET_INSTANCE_TYPES,
    AZURE_GET_OS_IMAGES,

    // Common provider events
    NETWORK_STEP_GET_NO_PROXY_INFO,

    // CLI
    CLI_CHANGED,

    // APP
    BRANDING_CHANGED,
    CONFIG_FILE_IMPORTED,
    CONFIG_FILE_IMPORT_ERROR
}

/**
 * Event type definition
 */
export interface TkgEvent {
    type: TkgEventType,
    payload?: any;
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
    subjects = new Map<TkgEventType, ReplaySubject<TkgEvent>>();

    /**
     * Return the subject based on event type
     * @param eventType event type to get the subject for
     */
    getSubject(eventType: TkgEventType) {
        let subject = this.subjects.get(eventType);
        if (!subject) {
            subject = new ReplaySubject<TkgEvent>(1);
            this.subjects.set(eventType, subject);
        }
        return subject;
    }

    /**
     * Publish an event to all its subscribers
     * @param event the Event to be published
     */
    publish(event: TkgEvent) {
        const subject = this.getSubject(event.type);
        subject.next(event);
    }

    /**
     * Clears specified event from the Messenger event map. Once this is done
     * subscribers will no longer receive this event until the event is re-dispatched.
     * @param eventType the event to delete from the Messenger buffer
     */
    clearEvent(eventType: TkgEventType) {
        this.subjects.delete(eventType);
    }

    /**
     * Reset/Clear the Messenger ReplaySubject event map in the rare use cases where
     * we want to force ALL events to be purged from the buffer.
     * This should be used only when necessary.
     */
    reset() {
        this.subjects = new Map<TkgEventType, ReplaySubject<TkgEvent>>();
    }
}
