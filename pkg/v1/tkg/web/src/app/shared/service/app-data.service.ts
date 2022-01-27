// Angular imports
import { Injectable } from '@angular/core';

// Library imports
import { BehaviorSubject } from 'rxjs';
import { Features } from 'tanzu-management-cluster-api';

// App imports
import { FeatureFlags, managementClusterPlugin } from '../../views/landing/wizard/shared/constants/wizard.constants';

@Injectable({
    providedIn: 'root'
})
export class AppDataService {

    private providerType = new BehaviorSubject<string|null>(null);
    private hasPacificCluster = new BehaviorSubject<boolean>(false);
    private tkrVersion = new BehaviorSubject<string|null>(null);
    private featureFlags = new BehaviorSubject<Features|null>(null);
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

    setFeatureFlags(features: Features) {
        this.featureFlags.next(features);
    }

    setVsphereVersion(version: string) {
        this.vsphereVersion.next(version);
    }

    getVsphereVersion() {
        return this.vsphereVersion;
    }

    isPluginFeatureActivated(plugin: string, feature: string) {
        return this.isValueTrue(this.getPluginFeature(plugin, feature));
    }

    getPluginFeature(plugin: string, feature: string) {
        if (this.featureFlags == null || this.featureFlags.value == null) {
            return null;
        }
        if (this.featureFlags.value[plugin] == null || this.featureFlags.value[plugin][feature] == null) {
            return null;
        }
        return this.featureFlags.value[plugin][feature];
    }

    isValueTrue(value: string) {
        return value !== null && JSON.parse(value);
    }

    // returns true if the standalone-cluster-mode feature flag is activated. This is a convenience method
    // to avoid having the string parameters scattered throughout the code
    isModeClusterStandalone() {
        return this.isPluginFeatureActivated(managementClusterPlugin, FeatureFlags.STANDALONE_CLUSTER);
    }

    decodeBase64(source: string): string {
        let encodedString = source;
        const encodedPrefix = '<encoded:';
        const encodedSuffix = '>';
        if (encodedString.startsWith(encodedPrefix) && encodedString.endsWith(encodedSuffix)) {
            // remove beginning and ending strings by taking a substring
            encodedString = encodedString.substring(encodedPrefix.length, encodedString.length - encodedSuffix.length);
        }
        return atob(encodedString);
    }

    isClusterNameRequired(): boolean {
        return this.isPluginFeatureActivated(managementClusterPlugin, FeatureFlags.CLUSTER_NAME_REQUIRED);
    }
}
