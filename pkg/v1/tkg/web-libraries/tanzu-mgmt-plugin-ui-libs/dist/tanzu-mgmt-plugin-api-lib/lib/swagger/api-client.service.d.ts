import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { InjectionToken } from '@angular/core';
import { Observable } from 'rxjs';
import { DefaultHttpOptions, HttpOptions, APIClientInterface } from './';
import * as models from './models';
import * as i0 from "@angular/core";
export declare const USE_DOMAIN: InjectionToken<string>;
export declare const USE_HTTP_OPTIONS: InjectionToken<HttpOptions>;
declare type APIHttpOptions = HttpOptions & {
    headers: HttpHeaders;
    params: HttpParams;
    responseType?: 'arraybuffer' | 'blob' | 'text' | 'json';
};
/**
 * Created with https://github.com/flowup/api-client-generator
 */
export declare class APIClient implements APIClientInterface {
    private readonly http;
    readonly options: APIHttpOptions;
    readonly domain: string;
    constructor(http: HttpClient, domain?: string, options?: DefaultHttpOptions);
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getUI(requestHttpOptions?: HttpOptions): Observable<File>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getUIFile(args: {
        filename: string;
    }, requestHttpOptions?: HttpOptions): Observable<File>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getFeatureFlags(requestHttpOptions?: HttpOptions): Observable<models.Features>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getTanzuEdition(requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    verifyAccount(args: {
        credentials?: models.AviControllerParams;
    }, requestHttpOptions?: HttpOptions): Observable<void>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapConnect(args: {
        credentials?: models.LdapParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapBind(requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapUserSearch(requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapGroupSearch(requestHttpOptions?: HttpOptions): Observable<models.LdapTestResult>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    verifyLdapCloseConnection(requestHttpOptions?: HttpOptions): Observable<void>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAviClouds(requestHttpOptions?: HttpOptions): Observable<models.AviCloud[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAviServiceEngineGroups(requestHttpOptions?: HttpOptions): Observable<models.AviServiceEngineGroup[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAviVipNetworks(requestHttpOptions?: HttpOptions): Observable<models.AviVipNetwork[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getProvider(requestHttpOptions?: HttpOptions): Observable<models.ProviderInfo>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVsphereThumbprint(args: {
        host: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereThumbprint>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    setVSphereEndpoint(args: {
        credentials?: models.VSphereCredentials;
    }, requestHttpOptions?: HttpOptions): Observable<models.VsphereInfo>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereDatacenters(requestHttpOptions?: HttpOptions): Observable<models.VSphereDatacenter[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereDatastores(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereDatastore[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereFolders(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereFolder[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereComputeResources(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereManagementObject[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereResourcePools(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereResourcePool[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereNetworks(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereNetwork[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereNodeTypes(requestHttpOptions?: HttpOptions): Observable<models.NodeType[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereOSImages(args: {
        dc: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.VSphereVirtualMachine[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForVsphere(args: {
        params: models.VsphereRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForVsphere(args: {
        params: models.VsphereRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForVsphere(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.VsphereRegionalClusterParams>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createVSphereRegionalCluster(args: {
        params: models.VsphereRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    setAWSEndpoint(args: {
        accountParams?: models.AWSAccountParams;
    }, requestHttpOptions?: HttpOptions): Observable<void>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVPCs(requestHttpOptions?: HttpOptions): Observable<models.Vpc[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSNodeTypes(args: {
        az?: string;
    }, requestHttpOptions?: HttpOptions): Observable<string[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSRegions(requestHttpOptions?: HttpOptions): Observable<string[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSOSImages(args: {
        region: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AWSVirtualMachine[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSCredentialProfiles(requestHttpOptions?: HttpOptions): Observable<string[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSAvailabilityZones(requestHttpOptions?: HttpOptions): Observable<models.AWSAvailabilityZone[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSSubnets(args: {
        vpcId: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AWSSubnet[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForAWS(args: {
        params: models.AWSRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForAWS(args: {
        params: models.AWSRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createAWSRegionalCluster(args: {
        params: models.AWSRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForAWS(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.AWSRegionalClusterParams>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureEndpoint(requestHttpOptions?: HttpOptions): Observable<models.AzureAccountParams>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    setAzureEndpoint(args: {
        accountParams?: models.AzureAccountParams;
    }, requestHttpOptions?: HttpOptions): Observable<void>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureResourceGroups(args: {
        location: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureResourceGroup[]>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    createAzureResourceGroup(args: {
        params: models.AzureResourceGroup;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureVnets(args: {
        resourceGroupName: string;
        location: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureVirtualNetwork[]>;
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    createAzureVirtualNetwork(args: {
        resourceGroupName: string;
        params: models.AzureVirtualNetwork;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureOSImages(requestHttpOptions?: HttpOptions): Observable<models.AzureVirtualMachine[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureRegions(requestHttpOptions?: HttpOptions): Observable<models.AzureLocation[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureInstanceTypes(args: {
        location: string;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureInstanceType[]>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForAzure(args: {
        params: models.AzureRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForAzure(args: {
        params: models.AzureRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createAzureRegionalCluster(args: {
        params: models.AzureRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForAzure(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.AzureRegionalClusterParams>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    checkIfDockerDaemonAvailable(requestHttpOptions?: HttpOptions): Observable<models.DockerDaemonStatus>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForDocker(args: {
        params: models.DockerRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForDocker(args: {
        params: models.DockerRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<models.ConfigFileInfo>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createDockerRegionalCluster(args: {
        params: models.DockerRegionalClusterParams;
    }, requestHttpOptions?: HttpOptions): Observable<string>;
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForDocker(args: {
        params: models.ConfigFile;
    }, requestHttpOptions?: HttpOptions): Observable<models.DockerRegionalClusterParams>;
    private sendRequest;
    static ɵfac: i0.ɵɵFactoryDeclaration<APIClient, [null, { optional: true; }, { optional: true; }]>;
    static ɵprov: i0.ɵɵInjectableDeclaration<APIClient>;
}
export {};
