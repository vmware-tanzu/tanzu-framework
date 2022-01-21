import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DefaultHttpOptions, HttpOptions } from './';
import { APIClient } from './api-client.service';
import * as models from './models';
import * as i0 from "@angular/core";
/**
 * Created with https://github.com/flowup/api-client-generator
 */
export declare class GuardedAPIClient extends APIClient {
    readonly httpClient: HttpClient;
    constructor(httpClient: HttpClient, domain?: string, options?: DefaultHttpOptions);
    getUI(requestHttpOptions?: HttpOptions): Observable<File>;
    getUIFile(args: {
        filename: string;
    }, requestHttpOptions?: HttpOptions): Observable<File>;
    getFeatureFlags(requestHttpOptions?: HttpOptions): Observable<models.Features>;
    getTanzuEdition(requestHttpOptions?: HttpOptions): Observable<string>;
    verifyLdapConnect(args: {
        credentials?: models.LdapParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    verifyLdapBind(requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    verifyLdapUserSearch(requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    verifyLdapGroupSearch(requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    getAviClouds(requestHttpOptions?: HttpOptions): Observable<models.AviCloud[]>;
    getAviServiceEngineGroups(requestHttpOptions?: HttpOptions): Observable<models.AviServiceEngineGroup[]>;
    getAviVipNetworks(requestHttpOptions?: HttpOptions): Observable<models.AviVipNetwork[]>;
    getProvider(requestHttpOptions?: HttpOptions): Observable<models.ProviderInfo>;
    getVsphereThumbprint(args: {
        host: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereThumbprint>;
    setVSphereEndpoint(args: {
        credentials?: models.VSphereCredentials;
    }, requestHttpOptions?: HttpOptions): Observable<models.VsphereInfo>;
    getVSphereDatacenters(requestHttpOptions?: HttpOptions): Observable<models.VSphereDatacenter[]>;
    getVSphereDatastores(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereDatastore[]>;
    getVSphereFolders(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereFolder[]>;
    getVSphereComputeResources(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereManagementObject[]>;
    getVSphereResourcePools(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereResourcePool[]>;
    getVSphereNetworks(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereNetwork[]>;
    getVSphereNodeTypes(requestHttpOptions?: HttpOptions): Observable<models.NodeType[]>;
    getVSphereOSImages(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereVirtualMachine[]>;
    exportTKGConfigForVsphere(args: {
        params: models.VsphereRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    applyTKGConfigForVsphere(args: {
        params: models.VsphereRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    importTKGConfigForVsphere(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.VsphereRegionalClusterParams>;
    createVSphereRegionalCluster(args: {
        params: models.VsphereRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    getVPCs(requestHttpOptions?: HttpOptions): Observable<models.Vpc[]>;
    getAWSNodeTypes(args: {
        az?: string;
    }, requestHttpOptions?: HttpOptions): Observable<string[]>;
    getAWSRegions(requestHttpOptions?: HttpOptions): Observable<string[]>;
    getAWSOSImages(args: {
        region: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AWSVirtualMachine[]>;
    getAWSCredentialProfiles(requestHttpOptions?: HttpOptions): Observable<string[]>;
    getAWSAvailabilityZones(requestHttpOptions?: HttpOptions): Observable<models.AWSAvailabilityZone[]>;
    getAWSSubnets(args: {
        vpcId: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AWSSubnet[]>;
    exportTKGConfigForAWS(args: {
        params: models.AWSRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    applyTKGConfigForAWS(args: {
        params: models.AWSRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    createAWSRegionalCluster(args: {
        params: models.AWSRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    importTKGConfigForAWS(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.AWSRegionalClusterParams>;
    getAzureEndpoint(requestHttpOptions?: HttpOptions): Observable<models.AzureAccountParams>;
    getAzureResourceGroups(args: {
        location: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureResourceGroup[]>;
    createAzureResourceGroup(args: {
        params: models.AzureResourceGroup;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    getAzureVnets(args: {
        resourceGroupName: string;
        location: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureVirtualNetwork[]>;
    createAzureVirtualNetwork(args: {
        resourceGroupName: string;
        params: models.AzureVirtualNetwork;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    getAzureOSImages(requestHttpOptions?: HttpOptions): Observable<models.AzureVirtualMachine[]>;
    getAzureRegions(requestHttpOptions?: HttpOptions): Observable<models.AzureLocation[]>;
    getAzureInstanceTypes(args: {
        location: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureInstanceType[]>;
    exportTKGConfigForAzure(args: {
        params: models.AzureRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    applyTKGConfigForAzure(args: {
        params: models.AzureRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    createAzureRegionalCluster(args: {
        params: models.AzureRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    importTKGConfigForAzure(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureRegionalClusterParams>;
    checkIfDockerDaemonAvailable(requestHttpOptions?: HttpOptions): Observable<models.DockerDaemonStatus>;
    exportTKGConfigForDocker(args: {
        params: models.DockerRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    applyTKGConfigForDocker(args: {
        params: models.DockerRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    createDockerRegionalCluster(args: {
        params: models.DockerRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    importTKGConfigForDocker(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.DockerRegionalClusterParams>;
    static ɵfac: i0.ɵɵFactoryDeclaration<GuardedAPIClient, [null, { optional: true; }, { optional: true; }]>;
    static ɵprov: i0.ɵɵInjectableDeclaration<GuardedAPIClient>;
}
