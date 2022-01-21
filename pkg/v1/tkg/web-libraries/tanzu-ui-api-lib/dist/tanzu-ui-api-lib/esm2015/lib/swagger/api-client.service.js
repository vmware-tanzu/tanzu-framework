/* tslint:disable */
import { HttpHeaders, HttpParams } from '@angular/common/http';
import { Inject, Injectable, InjectionToken, Optional } from '@angular/core';
import { throwError } from 'rxjs';
import * as i0 from "@angular/core";
import * as i1 from "@angular/common/http";
export const USE_DOMAIN = new InjectionToken('APIClient_USE_DOMAIN');
export const USE_HTTP_OPTIONS = new InjectionToken('APIClient_USE_HTTP_OPTIONS');
/**
 * Created with https://github.com/flowup/api-client-generator
 */
export class APIClient {
    constructor(http, domain, options) {
        this.http = http;
        this.domain = `//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}`;
        if (domain != null) {
            this.domain = domain;
        }
        this.options = Object.assign(Object.assign({ headers: new HttpHeaders(options && options.headers ? options.headers : {}), params: new HttpParams(options && options.params ? options.params : {}) }, (options && options.reportProgress ? { reportProgress: options.reportProgress } : {})), (options && options.withCredentials ? { withCredentials: options.withCredentials } : {}));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getUI(requestHttpOptions) {
        const path = `/`;
        const options = Object.assign(Object.assign(Object.assign({}, this.options), requestHttpOptions), { responseType: 'blob' });
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getUIFile(args, requestHttpOptions) {
        const path = `/${args.filename}`;
        const options = Object.assign(Object.assign(Object.assign({}, this.options), requestHttpOptions), { responseType: 'blob' });
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getFeatureFlags(requestHttpOptions) {
        const path = `/api/features`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getTanzuEdition(requestHttpOptions) {
        const path = `/api/edition`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    verifyAccount(args, requestHttpOptions) {
        const path = `/api/avi`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.credentials));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapConnect(args, requestHttpOptions) {
        const path = `/api/ldap/connect`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.credentials));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapBind(requestHttpOptions) {
        const path = `/api/ldap/bind`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapUserSearch(requestHttpOptions) {
        const path = `/api/ldap/users/search`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    verifyLdapGroupSearch(requestHttpOptions) {
        const path = `/api/ldap/groups/search`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options);
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    verifyLdapCloseConnection(requestHttpOptions) {
        const path = `/api/ldap/disconnect`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAviClouds(requestHttpOptions) {
        const path = `/api/avi/clouds`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAviServiceEngineGroups(requestHttpOptions) {
        const path = `/api/avi/serviceenginegroups`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAviVipNetworks(requestHttpOptions) {
        const path = `/api/avi/vipnetworks`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getProvider(requestHttpOptions) {
        const path = `/api/providers`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVsphereThumbprint(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/thumbprint`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('host' in args) {
            options.params = options.params.set('host', String(args.host));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    setVSphereEndpoint(args, requestHttpOptions) {
        const path = `/api/providers/vsphere`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.credentials));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereDatacenters(requestHttpOptions) {
        const path = `/api/providers/vsphere/datacenters`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereDatastores(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/datastores`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('dc' in args) {
            options.params = options.params.set('dc', String(args.dc));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereFolders(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/folders`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('dc' in args) {
            options.params = options.params.set('dc', String(args.dc));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereComputeResources(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/compute`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('dc' in args) {
            options.params = options.params.set('dc', String(args.dc));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereResourcePools(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/resourcepools`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('dc' in args) {
            options.params = options.params.set('dc', String(args.dc));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereNetworks(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/networks`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('dc' in args) {
            options.params = options.params.set('dc', String(args.dc));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereNodeTypes(requestHttpOptions) {
        const path = `/api/providers/vsphere/nodetypes`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVSphereOSImages(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/osimages`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('dc' in args) {
            options.params = options.params.set('dc', String(args.dc));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForVsphere(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/config/export`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForVsphere(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/tkgconfig`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForVsphere(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/config/import`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createVSphereRegionalCluster(args, requestHttpOptions) {
        const path = `/api/providers/vsphere/create`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    setAWSEndpoint(args, requestHttpOptions) {
        const path = `/api/providers/aws`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.accountParams));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getVPCs(requestHttpOptions) {
        const path = `/api/providers/aws/vpc`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSNodeTypes(args, requestHttpOptions) {
        const path = `/api/providers/aws/nodetypes`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('az' in args) {
            options.params = options.params.set('az', String(args.az));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSRegions(requestHttpOptions) {
        const path = `/api/providers/aws/regions`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSOSImages(args, requestHttpOptions) {
        const path = `/api/providers/aws/osimages`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('region' in args) {
            options.params = options.params.set('region', String(args.region));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSCredentialProfiles(requestHttpOptions) {
        const path = `/api/providers/aws/profiles`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSAvailabilityZones(requestHttpOptions) {
        const path = `/api/providers/aws/AvailabilityZones`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAWSSubnets(args, requestHttpOptions) {
        const path = `/api/providers/aws/subnets`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('vpcId' in args) {
            options.params = options.params.set('vpcId', String(args.vpcId));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForAWS(args, requestHttpOptions) {
        const path = `/api/providers/aws/config/export`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForAWS(args, requestHttpOptions) {
        const path = `/api/providers/aws/tkgconfig`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createAWSRegionalCluster(args, requestHttpOptions) {
        const path = `/api/providers/aws/create`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForAWS(args, requestHttpOptions) {
        const path = `/api/providers/aws/config/import`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureEndpoint(requestHttpOptions) {
        const path = `/api/providers/azure`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    setAzureEndpoint(args, requestHttpOptions) {
        const path = `/api/providers/azure`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.accountParams));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureResourceGroups(args, requestHttpOptions) {
        const path = `/api/providers/azure/resourcegroups`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('location' in args) {
            options.params = options.params.set('location', String(args.location));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    createAzureResourceGroup(args, requestHttpOptions) {
        const path = `/api/providers/azure/resourcegroups`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureVnets(args, requestHttpOptions) {
        const path = `/api/providers/azure/resourcegroups/${args.resourceGroupName}/vnets`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        if ('location' in args) {
            options.params = options.params.set('location', String(args.location));
        }
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 201 ] HTTP response code.
     */
    createAzureVirtualNetwork(args, requestHttpOptions) {
        const path = `/api/providers/azure/resourcegroups/${args.resourceGroupName}/vnets`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureOSImages(requestHttpOptions) {
        const path = `/api/providers/azure/osimages`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureRegions(requestHttpOptions) {
        const path = `/api/providers/azure/regions`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    getAzureInstanceTypes(args, requestHttpOptions) {
        const path = `/api/providers/azure/regions/${args.location}/instanceTypes`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForAzure(args, requestHttpOptions) {
        const path = `/api/providers/azure/config/export`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForAzure(args, requestHttpOptions) {
        const path = `/api/providers/azure/tkgconfig`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createAzureRegionalCluster(args, requestHttpOptions) {
        const path = `/api/providers/azure/create`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForAzure(args, requestHttpOptions) {
        const path = `/api/providers/azure/config/import`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    checkIfDockerDaemonAvailable(requestHttpOptions) {
        const path = `/api/providers/docker/daemon`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('GET', path, options);
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    exportTKGConfigForDocker(args, requestHttpOptions) {
        const path = `/api/providers/docker/config/export`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    applyTKGConfigForDocker(args, requestHttpOptions) {
        const path = `/api/providers/docker/tkgconfig`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    createDockerRegionalCluster(args, requestHttpOptions) {
        const path = `/api/providers/docker/create`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    /**
     * Response generated for [ 200 ] HTTP response code.
     */
    importTKGConfigForDocker(args, requestHttpOptions) {
        const path = `/api/providers/docker/config/import`;
        const options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
        return this.sendRequest('POST', path, options, JSON.stringify(args.params));
    }
    sendRequest(method, path, options, body) {
        switch (method) {
            case 'DELETE':
                return this.http.delete(`${this.domain}${path}`, options);
            case 'GET':
                return this.http.get(`${this.domain}${path}`, options);
            case 'HEAD':
                return this.http.head(`${this.domain}${path}`, options);
            case 'OPTIONS':
                return this.http.options(`${this.domain}${path}`, options);
            case 'PATCH':
                return this.http.patch(`${this.domain}${path}`, body, options);
            case 'POST':
                return this.http.post(`${this.domain}${path}`, body, options);
            case 'PUT':
                return this.http.put(`${this.domain}${path}`, body, options);
            default:
                console.error(`Unsupported request: ${method}`);
                return throwError(`Unsupported request: ${method}`);
        }
    }
}
APIClient.ɵfac = i0.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClient, deps: [{ token: i1.HttpClient }, { token: USE_DOMAIN, optional: true }, { token: USE_HTTP_OPTIONS, optional: true }], target: i0.ɵɵFactoryTarget.Injectable });
APIClient.ɵprov = i0.ɵɵngDeclareInjectable({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClient });
i0.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClient, decorators: [{
            type: Injectable
        }], ctorParameters: function () { return [{ type: i1.HttpClient }, { type: undefined, decorators: [{
                    type: Optional
                }, {
                    type: Inject,
                    args: [USE_DOMAIN]
                }] }, { type: undefined, decorators: [{
                    type: Optional
                }, {
                    type: Inject,
                    args: [USE_HTTP_OPTIONS]
                }] }]; } });
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiYXBpLWNsaWVudC5zZXJ2aWNlLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiLi4vLi4vLi4vLi4vLi4vcHJvamVjdHMvdGFuenUtdWktYXBpLWxpYi9zcmMvbGliL3N3YWdnZXIvYXBpLWNsaWVudC5zZXJ2aWNlLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLG9CQUFvQjtBQUVwQixPQUFPLEVBQWMsV0FBVyxFQUFFLFVBQVUsRUFBRSxNQUFNLHNCQUFzQixDQUFDO0FBQzNFLE9BQU8sRUFBRSxNQUFNLEVBQUUsVUFBVSxFQUFFLGNBQWMsRUFBRSxRQUFRLEVBQUUsTUFBTSxlQUFlLENBQUM7QUFDN0UsT0FBTyxFQUFjLFVBQVUsRUFBRSxNQUFNLE1BQU0sQ0FBQzs7O0FBSzlDLE1BQU0sQ0FBQyxNQUFNLFVBQVUsR0FBRyxJQUFJLGNBQWMsQ0FBUyxzQkFBc0IsQ0FBQyxDQUFDO0FBQzdFLE1BQU0sQ0FBQyxNQUFNLGdCQUFnQixHQUFHLElBQUksY0FBYyxDQUFjLDRCQUE0QixDQUFDLENBQUM7QUFROUY7O0dBRUc7QUFFSCxNQUFNLE9BQU8sU0FBUztJQU1wQixZQUE2QixJQUFnQixFQUNELE1BQWUsRUFDVCxPQUE0QjtRQUZqRCxTQUFJLEdBQUosSUFBSSxDQUFZO1FBRnBDLFdBQU0sR0FBVyxLQUFLLE1BQU0sQ0FBQyxRQUFRLENBQUMsUUFBUSxHQUFHLE1BQU0sQ0FBQyxRQUFRLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxHQUFHLEdBQUMsTUFBTSxDQUFDLFFBQVEsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLEVBQUUsRUFBRSxDQUFDO1FBTS9HLElBQUksTUFBTSxJQUFJLElBQUksRUFBRTtZQUNsQixJQUFJLENBQUMsTUFBTSxHQUFHLE1BQU0sQ0FBQztTQUN0QjtRQUVELElBQUksQ0FBQyxPQUFPLGlDQUNWLE9BQU8sRUFBRSxJQUFJLFdBQVcsQ0FBQyxPQUFPLElBQUksT0FBTyxDQUFDLE9BQU8sQ0FBQyxDQUFDLENBQUMs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SSxDQUFDLE1BQU0sR0FBRyxJQUFJLEVBQUUsRUFBRSxJQUFJLEVBQUUsT0FBTyxDQUFDLENBQUM7WUFDbkUsS0FBSyxLQUFLO2dCQUNSLE9BQU8sSUFBSSxDQUFDLElBQUksQ0FBQyxHQUFHLENBQUksR0FBRyxJQUFJLENBQUMsTUFBTSxHQUFHLElBQUksRUFBRSxFQUFFLElBQUksRUFBRSxPQUFPLENBQUMsQ0FBQztZQUNsRTtnQkFDRSxPQUFPLENBQUMsS0FBSyxDQUFDLHdCQUF3QixNQUFNLEVBQUUsQ0FBQyxDQUFDO2dCQUNoRCxPQUFPLFVBQVUsQ0FBQyx3QkFBd0IsTUFBTSxFQUFFLENBQUMsQ0FBQztTQUN2RDtJQUNILENBQUM7O3VHQXZpQ1UsU0FBUyw0Q0FPWSxVQUFVLDZCQUNWLGdCQUFnQjsyR0FSckMsU0FBUzs0RkFBVCxTQUFTO2tCQURyQixVQUFVOzswQkFRSSxRQUFROzswQkFBSSxNQUFNOzJCQUFDLFVBQVU7OzBCQUM3QixRQUFROzswQkFBSSxNQUFNOzJCQUFDLGdCQUFnQiIsInNvdXJjZXNDb250ZW50IjpbIi8qIHRzbGludDpkaXNhYmxlICovXG5cbmltcG9ydCB7IEh0dHBDbGllbnQsIEh0dHBIZWFkZXJzLCBIdHRwUGFyYW1zIH0gZnJvbSAnQGFuZ3VsYXIvY29tbW9uL2h0dHAnO1xuaW1wb3J0IHsgSW5qZWN0LCBJbmplY3RhYmxlLCBJbmplY3Rpb25Ub2tlbiwgT3B0aW9uYWwgfSBmcm9tICdAYW5ndWxhci9jb3JlJztcbmltcG9ydCB7IE9ic2VydmFibGUsIHRocm93RXJyb3IgfSBmcm9tICdyeGpzJztcbmltcG9ydCB7IERlZmF1bHRIdHRwT3B0aW9ucywgSHR0cE9wdGlvbnMsIEFQSUNsaWVudEludGVyZmFjZSB9IGZyb20gJy4vJztcblxuaW1wb3J0ICogYXMgbW9kZWxzIGZyb20gJy4vbW9kZWxzJztcblxuZXhwb3J0IGNvbnN0IFVTRV9ET01BSU4gPSBuZXcgSW5qZWN0aW9uVG9rZW48c3RyaW5nPignQVBJQ2xpZW50X1VTRV9ET01BSU4nKTtcbmV4cG9ydCBjb25zdCBVU0VfSFRUUF9PUFRJT05TID0gbmV3IEluamVjdGlvblRva2VuPEh0dHBPcHRpb25zPignQVBJQ2xpZW50X1VTRV9IVFRQX09QVElPTlMnKTtcblxudHlwZSBBUElIdHRwT3B0aW9ucyA9IEh0dHBPcHRpb25zICYge1xuICBoZWFkZXJzOiBIdHRwSGVhZGVycztcbiAgcGFyYW1zOiBIdHRwUGFyYW1zO1xuICByZXNwb25zZVR5cGU/OiAnYXJyYXlidWZmZXInIHwgJ2Jsb2InIHwgJ3RleHQnIHwgJ2pzb24nO1xufTtcblxuLyoqXG4gKiBDcmVhdGVkIHdpdGggaHR0cHM6Ly9naXRodWIuY29tL2Zsb3d1cC9hcGktY2xpZW50LWdlbmVyYXRvclxuICovXG5ASW5qZWN0YWJsZSgpXG5leHBvcnQgY2xhc3MgQVBJQ2xpZW50IGltcGxlbWVudHMgQVBJQ2xpZW50SW50ZXJmYWNlIHtcblxuICByZWFkb25seSBvcHRpb25zOiBBUElIdHRwT3B0aW9ucztcblxuICByZWFkb25seSBkb21haW46IHN0cmluZyA9IGAvLyR7d2luZG93LmxvY2F0aW9uLmhvc3RuYW1lfSR7d2luZG93LmxvY2F0aW9uLnBvcnQgPyAnOicrd2luZG93LmxvY2F0aW9uLnBvcnQgOiAnJ31gO1xuXG4gIGNvbnN0cnVjdG9yKHByaXZhdGUgcmVhZG9ubHkgaHR0cDogSHR0cENsaWVudCxcbiAgICAgICAgICAgICAgQE9wdGlvbmFsKCkgQEluamVjdChVU0VfRE9NQUlOKSBkb21haW4/OiBzdHJpbmcsXG4gICAgICAgICAgICAgIEBPcHRpb25hbCgpIEBJbmplY3QoVVNFX0hUVFBfT1BUSU9OUykgb3B0aW9ucz86IERlZmF1bHRIdHRwT3B0aW9ucykge1xuXG4gICAgaWYgKGRvbWFpbiAhPSBudWxsKSB7XG4gICAgICB0aGlzLmRvbWFpbiA9IGRvbWFpbjtcbiAgICB9XG5cbiAgICB0aGlzLm9wdGlvbnMgPSB7XG4gICAgICBoZWFkZXJzOiBuZXcgSHR0cEhlYWRlcnMob3B0aW9ucyAmJiBvcHRpb25zLmhlYWRlcnMgPyBvcHRpb25zLmhlYWRlcnMgOiB7fSksXG4gICAgICBwYXJhbXM6IG5ldyBIdHRwUGFyYW1zKG9wdGlvbnMgJiYgb3B0aW9ucy5wYXJhbXMgPyBvcHRpb25zLnBhcmFtcyA6IHt9KSxcbiAgICAgIC4uLihvcHRpb25zICYmIG9wdGlvbnMucmVwb3J0UHJvZ3Jlc3MgPyB7IHJlcG9ydFByb2dyZXNzOiBvcHRpb25zLnJlcG9ydFByb2dyZXNzIH0gOiB7fSksXG4gICAgICAuLi4ob3B0aW9ucyAmJiBvcHRpb25zLndpdGhDcmVkZW50aWFscyA/IHsgd2l0aENyZWRlbnRpYWxzOiBvcHRpb25zLndpdGhDcmVkZW50aWFscyB9IDoge30pXG4gICAgfTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0VUkoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxGaWxlPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICAgIHJlc3BvbnNlVHlwZTogJ2Jsb2InLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxGaWxlPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldFVJRmlsZShcbiAgICBhcmdzOiB7XG4gICAgICBmaWxlbmFtZTogc3RyaW5nLCAgLy8gVUkgZmlsZSBuYW1lXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPEZpbGU+IHtcbiAgICBjb25zdCBwYXRoID0gYC8ke2FyZ3MuZmlsZW5hbWV9YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICAgIHJlc3BvbnNlVHlwZTogJ2Jsb2InLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxGaWxlPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEZlYXR1cmVGbGFncyhcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5GZWF0dXJlcz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9mZWF0dXJlc2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5GZWF0dXJlcz4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRUYW56dUVkaXRpb24oXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmc+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvZWRpdGlvbmA7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PHN0cmluZz4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDEgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICB2ZXJpZnlBY2NvdW50KFxuICAgIGFyZ3M6IHtcbiAgICAgIGNyZWRlbnRpYWxzPzogbW9kZWxzLkF2aUNvbnRyb2xsZXJQYXJhbXMsICAvLyAob3B0aW9uYWwpIEF2aSBjb250cm9sbGVyIGNyZWRlbnRpYWxzXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPHZvaWQ+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvYXZpYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8dm9pZD4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLmNyZWRlbnRpYWxzKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIHZlcmlmeUxkYXBDb25uZWN0KFxuICAgIGFyZ3M6IHtcbiAgICAgIGNyZWRlbnRpYWxzPzogbW9kZWxzLkxkYXBQYXJhbXMsICAvLyAob3B0aW9uYWwpIExEQVAgY29uZmlndXJhdGlvblxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuTGRhcFRlc3RSZXN1bHQ+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvbGRhcC9jb25uZWN0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkxkYXBUZXN0UmVzdWx0PignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MuY3JlZGVudGlhbHMpKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgdmVyaWZ5TGRhcEJpbmQoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuTGRhcFRlc3RSZXN1bHQ+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvbGRhcC9iaW5kYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkxkYXBUZXN0UmVzdWx0PignUE9TVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICB2ZXJpZnlMZGFwVXNlclNlYXJjaChcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5MZGFwVGVzdFJlc3VsdD4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9sZGFwL3VzZXJzL3NlYXJjaGA7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5MZGFwVGVzdFJlc3VsdD4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgdmVyaWZ5TGRhcEdyb3VwU2VhcmNoKFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkxkYXBUZXN0UmVzdWx0PiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL2xkYXAvZ3JvdXBzL3NlYXJjaGA7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5MZGFwVGVzdFJlc3VsdD4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAxIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgdmVyaWZ5TGRhcENsb3NlQ29ubmVjdGlvbihcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPHZvaWQ+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvbGRhcC9kaXNjb25uZWN0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8dm9pZD4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0QXZpQ2xvdWRzKFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkF2aUNsb3VkW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvYXZpL2Nsb3Vkc2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5BdmlDbG91ZFtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEF2aVNlcnZpY2VFbmdpbmVHcm91cHMoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuQXZpU2VydmljZUVuZ2luZUdyb3VwW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvYXZpL3NlcnZpY2VlbmdpbmVncm91cHNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQXZpU2VydmljZUVuZ2luZUdyb3VwW10+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0QXZpVmlwTmV0d29ya3MoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuQXZpVmlwTmV0d29ya1tdPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL2F2aS92aXBuZXR3b3Jrc2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5BdmlWaXBOZXR3b3JrW10+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0UHJvdmlkZXIoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuUHJvdmlkZXJJbmZvPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVyc2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5Qcm92aWRlckluZm8+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0VnNwaGVyZVRodW1icHJpbnQoXG4gICAgYXJnczoge1xuICAgICAgaG9zdDogc3RyaW5nLCAgLy8gdlNwaGVyZSBob3N0XG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5WU3BoZXJlVGh1bWJwcmludD4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvdnNwaGVyZS90aHVtYnByaW50YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgaWYgKCdob3N0JyBpbiBhcmdzKSB7XG4gICAgICBvcHRpb25zLnBhcmFtcyA9IG9wdGlvbnMucGFyYW1zLnNldCgnaG9zdCcsIFN0cmluZyhhcmdzLmhvc3QpKTtcbiAgICB9XG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLlZTcGhlcmVUaHVtYnByaW50PignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMSBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIHNldFZTcGhlcmVFbmRwb2ludChcbiAgICBhcmdzOiB7XG4gICAgICBjcmVkZW50aWFscz86IG1vZGVscy5WU3BoZXJlQ3JlZGVudGlhbHMsICAvLyAob3B0aW9uYWwpIHZTcGhlcmUgY3JlZGVudGlhbHNcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZzcGhlcmVJbmZvPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy92c3BoZXJlYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLlZzcGhlcmVJbmZvPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MuY3JlZGVudGlhbHMpKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0VlNwaGVyZURhdGFjZW50ZXJzKFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZTcGhlcmVEYXRhY2VudGVyW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL3ZzcGhlcmUvZGF0YWNlbnRlcnNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuVlNwaGVyZURhdGFjZW50ZXJbXT4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRWU3BoZXJlRGF0YXN0b3JlcyhcbiAgICBhcmdzOiB7XG4gICAgICBkYzogc3RyaW5nLCAgLy8gZGF0YWNlbnRlciBtYW5hZ2VkIG9iamVjdCBJZCwgZS5nLiBkYXRhY2VudGVyLTJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZTcGhlcmVEYXRhc3RvcmVbXT4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvdnNwaGVyZS9kYXRhc3RvcmVzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgaWYgKCdkYycgaW4gYXJncykge1xuICAgICAgb3B0aW9ucy5wYXJhbXMgPSBvcHRpb25zLnBhcmFtcy5zZXQoJ2RjJywgU3RyaW5nKGFyZ3MuZGMpKTtcbiAgICB9XG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLlZTcGhlcmVEYXRhc3RvcmVbXT4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRWU3BoZXJlRm9sZGVycyhcbiAgICBhcmdzOiB7XG4gICAgICBkYzogc3RyaW5nLCAgLy8gZGF0YWNlbnRlciBtYW5hZ2VkIG9iamVjdCBJZCwgZS5nLiBkYXRhY2VudGVyLTJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZTcGhlcmVGb2xkZXJbXT4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvdnNwaGVyZS9mb2xkZXJzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgaWYgKCdkYycgaW4gYXJncykge1xuICAgICAgb3B0aW9ucy5wYXJhbXMgPSBvcHRpb25zLnBhcmFtcy5zZXQoJ2RjJywgU3RyaW5nKGFyZ3MuZGMpKTtcbiAgICB9XG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLlZTcGhlcmVGb2xkZXJbXT4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRWU3BoZXJlQ29tcHV0ZVJlc291cmNlcyhcbiAgICBhcmdzOiB7XG4gICAgICBkYzogc3RyaW5nLCAgLy8gZGF0YWNlbnRlciBtYW5hZ2VkIG9iamVjdCBJZCwgZS5nLiBkYXRhY2VudGVyLTJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZTcGhlcmVNYW5hZ2VtZW50T2JqZWN0W10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL3ZzcGhlcmUvY29tcHV0ZWA7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIGlmICgnZGMnIGluIGFyZ3MpIHtcbiAgICAgIG9wdGlvbnMucGFyYW1zID0gb3B0aW9ucy5wYXJhbXMuc2V0KCdkYycsIFN0cmluZyhhcmdzLmRjKSk7XG4gICAgfVxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5WU3BoZXJlTWFuYWdlbWVudE9iamVjdFtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldFZTcGhlcmVSZXNvdXJjZVBvb2xzKFxuICAgIGFyZ3M6IHtcbiAgICAgIGRjOiBzdHJpbmcsICAvLyBkYXRhY2VudGVyIG1hbmFnZWQgb2JqZWN0IElkLCBlLmcuIGRhdGFjZW50ZXItMlxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuVlNwaGVyZVJlc291cmNlUG9vbFtdPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy92c3BoZXJlL3Jlc291cmNlcG9vbHNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICBpZiAoJ2RjJyBpbiBhcmdzKSB7XG4gICAgICBvcHRpb25zLnBhcmFtcyA9IG9wdGlvbnMucGFyYW1zLnNldCgnZGMnLCBTdHJpbmcoYXJncy5kYykpO1xuICAgIH1cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuVlNwaGVyZVJlc291cmNlUG9vbFtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldFZTcGhlcmVOZXR3b3JrcyhcbiAgICBhcmdzOiB7XG4gICAgICBkYzogc3RyaW5nLCAgLy8gZGF0YWNlbnRlciBtYW5hZ2VkIG9iamVjdCBJZCwgZS5nLiBkYXRhY2VudGVyLTJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZTcGhlcmVOZXR3b3JrW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL3ZzcGhlcmUvbmV0d29ya3NgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICBpZiAoJ2RjJyBpbiBhcmdzKSB7XG4gICAgICBvcHRpb25zLnBhcmFtcyA9IG9wdGlvbnMucGFyYW1zLnNldCgnZGMnLCBTdHJpbmcoYXJncy5kYykpO1xuICAgIH1cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuVlNwaGVyZU5ldHdvcmtbXT4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRWU3BoZXJlTm9kZVR5cGVzKFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLk5vZGVUeXBlW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL3ZzcGhlcmUvbm9kZXR5cGVzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLk5vZGVUeXBlW10+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0VlNwaGVyZU9TSW1hZ2VzKFxuICAgIGFyZ3M6IHtcbiAgICAgIGRjOiBzdHJpbmcsICAvLyBkYXRhY2VudGVyIG1hbmFnZWQgb2JqZWN0IElkLCBlLmcuIGRhdGFjZW50ZXItMlxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuVlNwaGVyZVZpcnR1YWxNYWNoaW5lW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL3ZzcGhlcmUvb3NpbWFnZXNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICBpZiAoJ2RjJyBpbiBhcmdzKSB7XG4gICAgICBvcHRpb25zLnBhcmFtcyA9IG9wdGlvbnMucGFyYW1zLnNldCgnZGMnLCBTdHJpbmcoYXJncy5kYykpO1xuICAgIH1cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuVlNwaGVyZVZpcnR1YWxNYWNoaW5lW10+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZXhwb3J0VEtHQ29uZmlnRm9yVnNwaGVyZShcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5Wc3BoZXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zLCAgLy8gcGFyYW1zIHRvIGdlbmVyYXRlIHRrZyBjb25maWd1cmF0aW9uIGZvciB2c3BoZXJlXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPHN0cmluZz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvdnNwaGVyZS9jb25maWcvZXhwb3J0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGFwcGx5VEtHQ29uZmlnRm9yVnNwaGVyZShcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5Wc3BoZXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zLCAgLy8gcGFyYW1zIHRvIGFwcGx5IGNoYW5nZXMgdG8gdGtnIGNvbmZpZ3VyYXRpb24gZm9yIHZzcGhlcmVcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkNvbmZpZ0ZpbGVJbmZvPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy92c3BoZXJlL3RrZ2NvbmZpZ2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PG1vZGVscy5Db25maWdGaWxlSW5mbz4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLnBhcmFtcykpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBpbXBvcnRUS0dDb25maWdGb3JWc3BoZXJlKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLkNvbmZpZ0ZpbGUsICAvLyBjb25maWcgZmlsZSBmcm9tIHdoaWNoIHRvIGdlbmVyYXRlIHRrZyBjb25maWd1cmF0aW9uIGZvciB2c3BoZXJlXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5Wc3BoZXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy92c3BoZXJlL2NvbmZpZy9pbXBvcnRgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuVnNwaGVyZVJlZ2lvbmFsQ2x1c3RlclBhcmFtcz4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLnBhcmFtcykpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBjcmVhdGVWU3BoZXJlUmVnaW9uYWxDbHVzdGVyKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLlZzcGhlcmVSZWdpb25hbENsdXN0ZXJQYXJhbXMsICAvLyBwYXJhbXMgdG8gY3JlYXRlIGEgcmVnaW9uYWwgY2x1c3RlclxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmc+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL3ZzcGhlcmUvY3JlYXRlYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMSBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIHNldEFXU0VuZHBvaW50KFxuICAgIGFyZ3M6IHtcbiAgICAgIGFjY291bnRQYXJhbXM/OiBtb2RlbHMuQVdTQWNjb3VudFBhcmFtcywgIC8vIChvcHRpb25hbCkgQVdTIGFjY291bnQgcGFyYW1ldGVyc1xuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTx2b2lkPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9hd3NgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDx2b2lkPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MuYWNjb3VudFBhcmFtcykpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRWUENzKFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLlZwY1tdPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9hd3MvdnBjYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLlZwY1tdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEFXU05vZGVUeXBlcyhcbiAgICBhcmdzOiB7XG4gICAgICBhej86IHN0cmluZywgIC8vIChvcHRpb25hbCkgQVdTIGF2YWlsYWJpbGl0eSB6b25lLCBlLmcuIHVzLXdlc3QtMlxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmdbXT4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXdzL25vZGV0eXBlc2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIGlmICgnYXonIGluIGFyZ3MpIHtcbiAgICAgIG9wdGlvbnMucGFyYW1zID0gb3B0aW9ucy5wYXJhbXMuc2V0KCdheicsIFN0cmluZyhhcmdzLmF6KSk7XG4gICAgfVxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PHN0cmluZ1tdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEFXU1JlZ2lvbnMoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmdbXT4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXdzL3JlZ2lvbnNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxzdHJpbmdbXT4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRBV1NPU0ltYWdlcyhcbiAgICBhcmdzOiB7XG4gICAgICByZWdpb246IHN0cmluZywgIC8vIEFXUyByZWdpb24sIGUuZy4gdXMtd2VzdC0yXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5BV1NWaXJ0dWFsTWFjaGluZVtdPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9hd3Mvb3NpbWFnZXNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICBpZiAoJ3JlZ2lvbicgaW4gYXJncykge1xuICAgICAgb3B0aW9ucy5wYXJhbXMgPSBvcHRpb25zLnBhcmFtcy5zZXQoJ3JlZ2lvbicsIFN0cmluZyhhcmdzLnJlZ2lvbikpO1xuICAgIH1cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQVdTVmlydHVhbE1hY2hpbmVbXT4oJ0dFVCcsIHBhdGgsIG9wdGlvbnMpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRBV1NDcmVkZW50aWFsUHJvZmlsZXMoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmdbXT4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXdzL3Byb2ZpbGVzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nW10+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZ2V0QVdTQXZhaWxhYmlsaXR5Wm9uZXMoXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuQVdTQXZhaWxhYmlsaXR5Wm9uZVtdPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9hd3MvQXZhaWxhYmlsaXR5Wm9uZXNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQVdTQXZhaWxhYmlsaXR5Wm9uZVtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEFXU1N1Ym5ldHMoXG4gICAgYXJnczoge1xuICAgICAgdnBjSWQ6IHN0cmluZywgIC8vIFZQQyBJZFxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuQVdTU3VibmV0W10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F3cy9zdWJuZXRzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgaWYgKCd2cGNJZCcgaW4gYXJncykge1xuICAgICAgb3B0aW9ucy5wYXJhbXMgPSBvcHRpb25zLnBhcmFtcy5zZXQoJ3ZwY0lkJywgU3RyaW5nKGFyZ3MudnBjSWQpKTtcbiAgICB9XG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkFXU1N1Ym5ldFtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGV4cG9ydFRLR0NvbmZpZ0ZvckFXUyhcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5BV1NSZWdpb25hbENsdXN0ZXJQYXJhbXMsICAvLyBwYXJhbWV0ZXJzIHRvIGdlbmVyYXRlIFRLRyBjb25maWd1cmF0aW9uIGZpbGUgZm9yIEFXU1xuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmc+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F3cy9jb25maWcvZXhwb3J0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGFwcGx5VEtHQ29uZmlnRm9yQVdTKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLkFXU1JlZ2lvbmFsQ2x1c3RlclBhcmFtcywgIC8vIHBhcmFtZXRlcnMgdG8gYXBwbHkgY2hhbmdlcyB0byBUS0cgY29uZmlndXJhdGlvbiBmaWxlIGZvciBBV1NcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkNvbmZpZ0ZpbGVJbmZvPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9hd3MvdGtnY29uZmlnYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkNvbmZpZ0ZpbGVJbmZvPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGNyZWF0ZUFXU1JlZ2lvbmFsQ2x1c3RlcihcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5BV1NSZWdpb25hbENsdXN0ZXJQYXJhbXMsICAvLyBwYXJhbWV0ZXJzIHRvIGNyZWF0ZSBhIHJlZ2lvbmFsIGNsdXN0ZXJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8c3RyaW5nPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9hd3MvY3JlYXRlYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGltcG9ydFRLR0NvbmZpZ0ZvckFXUyhcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5Db25maWdGaWxlLCAgLy8gY29uZmlnIGZpbGUgZnJvbSB3aGljaCB0byBnZW5lcmF0ZSB0a2cgY29uZmlndXJhdGlvbiBmb3IgYXdzXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5BV1NSZWdpb25hbENsdXN0ZXJQYXJhbXM+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F3cy9jb25maWcvaW1wb3J0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkFXU1JlZ2lvbmFsQ2x1c3RlclBhcmFtcz4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLnBhcmFtcykpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRBenVyZUVuZHBvaW50KFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkF6dXJlQWNjb3VudFBhcmFtcz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXp1cmVgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQXp1cmVBY2NvdW50UGFyYW1zPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMSBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIHNldEF6dXJlRW5kcG9pbnQoXG4gICAgYXJnczoge1xuICAgICAgYWNjb3VudFBhcmFtcz86IG1vZGVscy5BenVyZUFjY291bnRQYXJhbXMsICAvLyAob3B0aW9uYWwpIEF6dXJlIGFjY291bnQgcGFyYW1ldGVyc1xuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTx2b2lkPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9henVyZWA7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PHZvaWQ+KCdQT1NUJywgcGF0aCwgb3B0aW9ucywgSlNPTi5zdHJpbmdpZnkoYXJncy5hY2NvdW50UGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEF6dXJlUmVzb3VyY2VHcm91cHMoXG4gICAgYXJnczoge1xuICAgICAgbG9jYXRpb246IHN0cmluZywgIC8vIEF6dXJlIHJlZ2lvblxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuQXp1cmVSZXNvdXJjZUdyb3VwW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F6dXJlL3Jlc291cmNlZ3JvdXBzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgaWYgKCdsb2NhdGlvbicgaW4gYXJncykge1xuICAgICAgb3B0aW9ucy5wYXJhbXMgPSBvcHRpb25zLnBhcmFtcy5zZXQoJ2xvY2F0aW9uJywgU3RyaW5nKGFyZ3MubG9jYXRpb24pKTtcbiAgICB9XG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkF6dXJlUmVzb3VyY2VHcm91cFtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMSBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGNyZWF0ZUF6dXJlUmVzb3VyY2VHcm91cChcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5BenVyZVJlc291cmNlR3JvdXAsICAvLyBwYXJhbWV0ZXJzIHRvIGNyZWF0ZSBhIG5ldyBBenVyZSByZXNvdXJjZSBncm91cFxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmc+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F6dXJlL3Jlc291cmNlZ3JvdXBzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEF6dXJlVm5ldHMoXG4gICAgYXJnczoge1xuICAgICAgcmVzb3VyY2VHcm91cE5hbWU6IHN0cmluZywgIC8vIE5hbWUgb2YgdGhlIEF6dXJlIHJlc291cmNlIGdyb3VwXG4gICAgICBsb2NhdGlvbjogc3RyaW5nLCAgLy8gQXp1cmUgcmVnaW9uXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5BenVyZVZpcnR1YWxOZXR3b3JrW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F6dXJlL3Jlc291cmNlZ3JvdXBzLyR7YXJncy5yZXNvdXJjZUdyb3VwTmFtZX0vdm5ldHNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICBpZiAoJ2xvY2F0aW9uJyBpbiBhcmdzKSB7XG4gICAgICBvcHRpb25zLnBhcmFtcyA9IG9wdGlvbnMucGFyYW1zLnNldCgnbG9jYXRpb24nLCBTdHJpbmcoYXJncy5sb2NhdGlvbikpO1xuICAgIH1cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQXp1cmVWaXJ0dWFsTmV0d29ya1tdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMSBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGNyZWF0ZUF6dXJlVmlydHVhbE5ldHdvcmsoXG4gICAgYXJnczoge1xuICAgICAgcmVzb3VyY2VHcm91cE5hbWU6IHN0cmluZywgIC8vIE5hbWUgb2YgdGhlIEF6dXJlIHJlc291cmNlIGdyb3VwXG4gICAgICBwYXJhbXM6IG1vZGVscy5BenVyZVZpcnR1YWxOZXR3b3JrLCAgLy8gcGFyYW1ldGVycyB0byBjcmVhdGUgYSBuZXcgQXp1cmUgVmlydHVhbCBuZXR3b3JrXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPHN0cmluZz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXp1cmUvcmVzb3VyY2Vncm91cHMvJHthcmdzLnJlc291cmNlR3JvdXBOYW1lfS92bmV0c2A7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PHN0cmluZz4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLnBhcmFtcykpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBnZXRBenVyZU9TSW1hZ2VzKFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkF6dXJlVmlydHVhbE1hY2hpbmVbXT4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXp1cmUvb3NpbWFnZXNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQXp1cmVWaXJ0dWFsTWFjaGluZVtdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEF6dXJlUmVnaW9ucyhcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5BenVyZUxvY2F0aW9uW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F6dXJlL3JlZ2lvbnNgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQXp1cmVMb2NhdGlvbltdPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGdldEF6dXJlSW5zdGFuY2VUeXBlcyhcbiAgICBhcmdzOiB7XG4gICAgICBsb2NhdGlvbjogc3RyaW5nLCAgLy8gQXp1cmUgcmVnaW9uIG5hbWVcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkF6dXJlSW5zdGFuY2VUeXBlW10+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F6dXJlL3JlZ2lvbnMvJHthcmdzLmxvY2F0aW9ufS9pbnN0YW5jZVR5cGVzYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkF6dXJlSW5zdGFuY2VUeXBlW10+KCdHRVQnLCBwYXRoLCBvcHRpb25zKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgZXhwb3J0VEtHQ29uZmlnRm9yQXp1cmUoXG4gICAgYXJnczoge1xuICAgICAgcGFyYW1zOiBtb2RlbHMuQXp1cmVSZWdpb25hbENsdXN0ZXJQYXJhbXMsICAvLyBwYXJhbWV0ZXJzIHRvIGdlbmVyYXRlIFRLRyBjb25maWd1cmF0aW9uIGZpbGUgZm9yIEF6dXJlXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPHN0cmluZz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXp1cmUvY29uZmlnL2V4cG9ydGA7XG4gICAgY29uc3Qgb3B0aW9uczogQVBJSHR0cE9wdGlvbnMgPSB7XG4gICAgICAuLi50aGlzLm9wdGlvbnMsXG4gICAgICAuLi5yZXF1ZXN0SHR0cE9wdGlvbnMsXG4gICAgfTtcblxuICAgIHJldHVybiB0aGlzLnNlbmRSZXF1ZXN0PHN0cmluZz4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLnBhcmFtcykpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJlc3BvbnNlIGdlbmVyYXRlZCBmb3IgWyAyMDAgXSBIVFRQIHJlc3BvbnNlIGNvZGUuXG4gICAqL1xuICBhcHBseVRLR0NvbmZpZ0ZvckF6dXJlKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLkF6dXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zLCAgLy8gcGFyYW1ldGVycyB0byBhcHBseSBjaGFuZ2VzIHRvIFRLRyBjb25maWd1cmF0aW9uIGZpbGUgZm9yIEF6dXJlXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5Db25maWdGaWxlSW5mbz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXp1cmUvdGtnY29uZmlnYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkNvbmZpZ0ZpbGVJbmZvPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGNyZWF0ZUF6dXJlUmVnaW9uYWxDbHVzdGVyKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLkF6dXJlUmVnaW9uYWxDbHVzdGVyUGFyYW1zLCAgLy8gcGFyYW1ldGVycyB0byBjcmVhdGUgYSByZWdpb25hbCBjbHVzdGVyXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPHN0cmluZz4ge1xuICAgIGNvbnN0IHBhdGggPSBgL2FwaS9wcm92aWRlcnMvYXp1cmUvY3JlYXRlYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGltcG9ydFRLR0NvbmZpZ0ZvckF6dXJlKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLkNvbmZpZ0ZpbGUsICAvLyBjb25maWcgZmlsZSBmcm9tIHdoaWNoIHRvIGdlbmVyYXRlIHRrZyBjb25maWd1cmF0aW9uIGZvciBhenVyZVxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxtb2RlbHMuQXp1cmVSZWdpb25hbENsdXN0ZXJQYXJhbXM+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2F6dXJlL2NvbmZpZy9pbXBvcnRgO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuQXp1cmVSZWdpb25hbENsdXN0ZXJQYXJhbXM+KCdQT1NUJywgcGF0aCwgb3B0aW9ucywgSlNPTi5zdHJpbmdpZnkoYXJncy5wYXJhbXMpKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBSZXNwb25zZSBnZW5lcmF0ZWQgZm9yIFsgMjAwIF0gSFRUUCByZXNwb25zZSBjb2RlLlxuICAgKi9cbiAgY2hlY2tJZkRvY2tlckRhZW1vbkF2YWlsYWJsZShcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5Eb2NrZXJEYWVtb25TdGF0dXM+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2RvY2tlci9kYWVtb25gO1xuICAgIGNvbnN0IG9wdGlvbnM6IEFQSUh0dHBPcHRpb25zID0ge1xuICAgICAgLi4udGhpcy5vcHRpb25zLFxuICAgICAgLi4ucmVxdWVzdEh0dHBPcHRpb25zLFxuICAgIH07XG5cbiAgICByZXR1cm4gdGhpcy5zZW5kUmVxdWVzdDxtb2RlbHMuRG9ja2VyRGFlbW9uU3RhdHVzPignR0VUJywgcGF0aCwgb3B0aW9ucyk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGV4cG9ydFRLR0NvbmZpZ0ZvckRvY2tlcihcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5Eb2NrZXJSZWdpb25hbENsdXN0ZXJQYXJhbXMsICAvLyBwYXJhbWV0ZXJzIHRvIGdlbmVyYXRlIFRLRyBjb25maWd1cmF0aW9uIGZpbGUgZm9yIERvY2tlclxuICAgIH0sXG4gICAgcmVxdWVzdEh0dHBPcHRpb25zPzogSHR0cE9wdGlvbnNcbiAgKTogT2JzZXJ2YWJsZTxzdHJpbmc+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2RvY2tlci9jb25maWcvZXhwb3J0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGFwcGx5VEtHQ29uZmlnRm9yRG9ja2VyKFxuICAgIGFyZ3M6IHtcbiAgICAgIHBhcmFtczogbW9kZWxzLkRvY2tlclJlZ2lvbmFsQ2x1c3RlclBhcmFtcywgIC8vIHBhcmFtZXRlcnMgdG8gYXBwbHkgY2hhbmdlcyB0byBUS0cgY29uZmlndXJhdGlvbiBmaWxlIGZvciBEb2NrZXJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8bW9kZWxzLkNvbmZpZ0ZpbGVJbmZvPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9kb2NrZXIvdGtnY29uZmlnYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkNvbmZpZ0ZpbGVJbmZvPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGNyZWF0ZURvY2tlclJlZ2lvbmFsQ2x1c3RlcihcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5Eb2NrZXJSZWdpb25hbENsdXN0ZXJQYXJhbXMsICAvLyBwYXJhbWV0ZXJzIHRvIGNyZWF0ZSBhIHJlZ2lvbmFsIGNsdXN0ZXJcbiAgICB9LFxuICAgIHJlcXVlc3RIdHRwT3B0aW9ucz86IEh0dHBPcHRpb25zXG4gICk6IE9ic2VydmFibGU8c3RyaW5nPiB7XG4gICAgY29uc3QgcGF0aCA9IGAvYXBpL3Byb3ZpZGVycy9kb2NrZXIvY3JlYXRlYDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8c3RyaW5nPignUE9TVCcsIHBhdGgsIG9wdGlvbnMsIEpTT04uc3RyaW5naWZ5KGFyZ3MucGFyYW1zKSk7XG4gIH1cblxuICAvKipcbiAgICogUmVzcG9uc2UgZ2VuZXJhdGVkIGZvciBbIDIwMCBdIEhUVFAgcmVzcG9uc2UgY29kZS5cbiAgICovXG4gIGltcG9ydFRLR0NvbmZpZ0ZvckRvY2tlcihcbiAgICBhcmdzOiB7XG4gICAgICBwYXJhbXM6IG1vZGVscy5Db25maWdGaWxlLCAgLy8gY29uZmlnIGZpbGUgZnJvbSB3aGljaCB0byBnZW5lcmF0ZSB0a2cgY29uZmlndXJhdGlvbiBmb3IgZG9ja2VyXG4gICAgfSxcbiAgICByZXF1ZXN0SHR0cE9wdGlvbnM/OiBIdHRwT3B0aW9uc1xuICApOiBPYnNlcnZhYmxlPG1vZGVscy5Eb2NrZXJSZWdpb25hbENsdXN0ZXJQYXJhbXM+IHtcbiAgICBjb25zdCBwYXRoID0gYC9hcGkvcHJvdmlkZXJzL2RvY2tlci9jb25maWcvaW1wb3J0YDtcbiAgICBjb25zdCBvcHRpb25zOiBBUElIdHRwT3B0aW9ucyA9IHtcbiAgICAgIC4uLnRoaXMub3B0aW9ucyxcbiAgICAgIC4uLnJlcXVlc3RIdHRwT3B0aW9ucyxcbiAgICB9O1xuXG4gICAgcmV0dXJuIHRoaXMuc2VuZFJlcXVlc3Q8bW9kZWxzLkRvY2tlclJlZ2lvbmFsQ2x1c3RlclBhcmFtcz4oJ1BPU1QnLCBwYXRoLCBvcHRpb25zLCBKU09OLnN0cmluZ2lmeShhcmdzLnBhcmFtcykpO1xuICB9XG5cbiAgcHJpdmF0ZSBzZW5kUmVxdWVzdDxUPihtZXRob2Q6IHN0cmluZywgcGF0aDogc3RyaW5nLCBvcHRpb25zOiBIdHRwT3B0aW9ucywgYm9keT86IGFueSk6IE9ic2VydmFibGU8VD4ge1xuICAgIHN3aXRjaCAobWV0aG9kKSB7XG4gICAgICBjYXNlICdERUxFVEUnOlxuICAgICAgICByZXR1cm4gdGhpcy5odHRwLmRlbGV0ZTxUPihgJHt0aGlzLmRvbWFpbn0ke3BhdGh9YCwgb3B0aW9ucyk7XG4gICAgICBjYXNlICdHRVQnOlxuICAgICAgICByZXR1cm4gdGhpcy5odHRwLmdldDxUPihgJHt0aGlzLmRvbWFpbn0ke3BhdGh9YCwgb3B0aW9ucyk7XG4gICAgICBjYXNlICdIRUFEJzpcbiAgICAgICAgcmV0dXJuIHRoaXMuaHR0cC5oZWFkPFQ+KGAke3RoaXMuZG9tYWlufSR7cGF0aH1gLCBvcHRpb25zKTtcbiAgICAgIGNhc2UgJ09QVElPTlMnOlxuICAgICAgICByZXR1cm4gdGhpcy5odHRwLm9wdGlvbnM8VD4oYCR7dGhpcy5kb21haW59JHtwYXRofWAsIG9wdGlvbnMpO1xuICAgICAgY2FzZSAnUEFUQ0gnOlxuICAgICAgICByZXR1cm4gdGhpcy5odHRwLnBhdGNoPFQ+KGAke3RoaXMuZG9tYWlufSR7cGF0aH1gLCBib2R5LCBvcHRpb25zKTtcbiAgICAgIGNhc2UgJ1BPU1QnOlxuICAgICAgICByZXR1cm4gdGhpcy5odHRwLnBvc3Q8VD4oYCR7dGhpcy5kb21haW59JHtwYXRofWAsIGJvZHksIG9wdGlvbnMpO1xuICAgICAgY2FzZSAnUFVUJzpcbiAgICAgICAgcmV0dXJuIHRoaXMuaHR0cC5wdXQ8VD4oYCR7dGhpcy5kb21haW59JHtwYXRofWAsIGJvZHksIG9wdGlvbnMpO1xuICAgICAgZGVmYXVsdDpcbiAgICAgICAgY29uc29sZS5lcnJvcihgVW5zdXBwb3J0ZWQgcmVxdWVzdDogJHttZXRob2R9YCk7XG4gICAgICAgIHJldHVybiB0aHJvd0Vycm9yKGBVbnN1cHBvcnRlZCByZXF1ZXN0OiAke21ldGhvZH1gKTtcbiAgICB9XG4gIH1cbn1cbiJdfQ==