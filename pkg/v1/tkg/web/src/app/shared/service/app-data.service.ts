// Angular imports
import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

@Injectable({
    providedIn: 'root'
})
export class AppDataService {

    private providerType = new BehaviorSubject<string|null>(null);
    private hasPacificCluster = new BehaviorSubject<boolean>(false);
    private tkrVersion = new BehaviorSubject<string|null>(null);
    private featureFlags = new BehaviorSubject<Map<String, String>|null>(null);
    private vsphereVersion = new BehaviorSubject<string|null>(null);

    constructor() {
        this.providerType.asObservable().subscribe((data) => {
            if (data) {
                console.log("TKG Kickstart UI launched with provider type >>>>> " + data);
            }
        });
    }

    setProviderType(provider: string) {
        this.providerType.next(provider);
    }

    getProviderType() {
        return this.providerType;
    }

    setIsProjPacific(flag: boolean) {
        this.hasPacificCluster.next(flag);
    }

    getIsProjPacific() {
        return this.hasPacificCluster;
    }

    setTkrVersion(version: string) {
        this.tkrVersion.next(version);
    }

    getTkrVersion() {
        return this.tkrVersion;
    }

    setFeatureFlags(flags: Map<String, String>) {
        this.featureFlags.next(flags);
    }

    getFeatureFlags() {
        return this.featureFlags;
    }

    setVsphereVersion(version: string) {
        this.vsphereVersion.next(version);
    }

    getVsphereVersion() {
        return this.vsphereVersion;
    }
}
