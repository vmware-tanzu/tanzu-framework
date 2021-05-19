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

    // AWS events
    AWS_REGION_CHANGED,
    AWS_VPC_TYPE_CHANGED,
    AWS_VPC_CHANGED,
    AWS_GET_EXISTING_VPCS,
    AWS_GET_AVAILABILITY_ZONES,
    AWS_GET_SUBNETS,
    AWS_GET_NODE_TYPES,
    AWS_GET_NO_PROXY_INFO,
    AWS_GET_OS_IMAGES,
    AWS_AIRGAPPED_VPC_CHANGE,

    // AZURE events
    AZURE_REGION_CHANGED,
    AZURE_RESOURCEGROUP_CHANGED,
    AZURE_GET_RESOURCE_GROUPS,
    AZURE_GET_VNETS,
    AZURE_GET_INSTANCE_TYPES,
    AZURE_GET_OS_IMAGES,

    // CLI
    CLI_CHANGED,

    // APP
    BRANDING_CHANGED
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
        const subject = this.subjects.get(eventType) || new ReplaySubject<TkgEvent>(1);
        this.subjects.set(eventType, subject);
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
}
