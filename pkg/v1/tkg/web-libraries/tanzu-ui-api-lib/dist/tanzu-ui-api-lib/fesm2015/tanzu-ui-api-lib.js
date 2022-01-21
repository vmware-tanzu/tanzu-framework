import * as i0 from '@angular/core';
import { InjectionToken, Injectable, Optional, Inject, NgModule } from '@angular/core';
import * as i1 from '@angular/common/http';
import { HttpHeaders, HttpParams } from '@angular/common/http';
import { throwError } from 'rxjs';
import { tap } from 'rxjs/operators';

/* tslint:disable */
const USE_DOMAIN = new InjectionToken('APIClient_USE_DOMAIN');
const USE_HTTP_OPTIONS = new InjectionToken('APIClient_USE_HTTP_OPTIONS');
/**
 * Created with https://github.com/flowup/api-client-generator
 */
class APIClient {
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

/* tslint:disable */
/* pre-prepared guards for build in complex types */
function _isBlob(arg) {
    return arg != null && typeof arg.size === 'number' && typeof arg.type === 'string' && typeof arg.slice === 'function';
}
function isFile(arg) {
    return arg != null && typeof arg.lastModified === 'number' && typeof arg.name === 'string' && _isBlob(arg);
}
/* generated type guards */
function isAviCloud(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // location?: string
        (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // uuid?: string
        (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
        true);
}
function isAviConfig(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // ca_cert?: string
        (typeof arg.ca_cert === 'undefined' || typeof arg.ca_cert === 'string') &&
        // cloud?: string
        (typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string') &&
        // controller?: string
        (typeof arg.controller === 'undefined' || typeof arg.controller === 'string') &&
        // controlPlaneHaProvider?: boolean
        (typeof arg.controlPlaneHaProvider === 'undefined' || typeof arg.controlPlaneHaProvider === 'boolean') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // managementClusterVipNetworkCidr?: string
        (typeof arg.managementClusterVipNetworkCidr === 'undefined' || typeof arg.managementClusterVipNetworkCidr === 'string') &&
        // managementClusterVipNetworkName?: string
        (typeof arg.managementClusterVipNetworkName === 'undefined' || typeof arg.managementClusterVipNetworkName === 'string') &&
        // network?: AviNetworkParams
        (typeof arg.network === 'undefined' || isAviNetworkParams(arg.network)) &&
        // password?: string
        (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
        // service_engine?: string
        (typeof arg.service_engine === 'undefined' || typeof arg.service_engine === 'string') &&
        // username?: string
        (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
        true);
}
function isAviControllerParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // CAData?: string
        (typeof arg.CAData === 'undefined' || typeof arg.CAData === 'string') &&
        // host?: string
        (typeof arg.host === 'undefined' || typeof arg.host === 'string') &&
        // password?: string
        (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
        // tenant?: string
        (typeof arg.tenant === 'undefined' || typeof arg.tenant === 'string') &&
        // username?: string
        (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
        true);
}
function isAviNetworkParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isAviServiceEngineGroup(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // location?: string
        (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // uuid?: string
        (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
        true);
}
function isAviSubnet(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // family?: string
        (typeof arg.family === 'undefined' || typeof arg.family === 'string') &&
        // subnet?: string
        (typeof arg.subnet === 'undefined' || typeof arg.subnet === 'string') &&
        true);
}
function isAviVipNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cloud?: string
        (typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string') &&
        // configedSubnets?: AviSubnet[]
        (typeof arg.configedSubnets === 'undefined' || (Array.isArray(arg.configedSubnets) && arg.configedSubnets.every((item) => isAviSubnet(item)))) &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // uuid?: string
        (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
        true);
}
function isAWSAccountParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // accessKeyID?: string
        (typeof arg.accessKeyID === 'undefined' || typeof arg.accessKeyID === 'string') &&
        // profileName?: string
        (typeof arg.profileName === 'undefined' || typeof arg.profileName === 'string') &&
        // region?: string
        (typeof arg.region === 'undefined' || typeof arg.region === 'string') &&
        // secretAccessKey?: string
        (typeof arg.secretAccessKey === 'undefined' || typeof arg.secretAccessKey === 'string') &&
        // sessionToken?: string
        (typeof arg.sessionToken === 'undefined' || typeof arg.sessionToken === 'string') &&
        true);
}
function isAWSAvailabilityZone(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isAWSNodeAz(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // privateSubnetID?: string
        (typeof arg.privateSubnetID === 'undefined' || typeof arg.privateSubnetID === 'string') &&
        // publicSubnetID?: string
        (typeof arg.publicSubnetID === 'undefined' || typeof arg.publicSubnetID === 'string') &&
        // workerNodeType?: string
        (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
        true);
}
function isAWSRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // awsAccountParams?: AWSAccountParams
        (typeof arg.awsAccountParams === 'undefined' || isAWSAccountParams(arg.awsAccountParams)) &&
        // bastionHostEnabled?: boolean
        (typeof arg.bastionHostEnabled === 'undefined' || typeof arg.bastionHostEnabled === 'boolean') &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // controlPlaneNodeType?: string
        (typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string') &&
        // createCloudFormationStack?: boolean
        (typeof arg.createCloudFormationStack === 'undefined' || typeof arg.createCloudFormationStack === 'boolean') &&
        // enableAuditLogging?: boolean
        (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // loadbalancerSchemeInternal?: boolean
        (typeof arg.loadbalancerSchemeInternal === 'undefined' || typeof arg.loadbalancerSchemeInternal === 'boolean') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNode?: number
        (typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number') &&
        // os?: AWSVirtualMachine
        (typeof arg.os === 'undefined' || isAWSVirtualMachine(arg.os)) &&
        // sshKeyName?: string
        (typeof arg.sshKeyName === 'undefined' || typeof arg.sshKeyName === 'string') &&
        // vpc?: AWSVpc
        (typeof arg.vpc === 'undefined' || isAWSVpc(arg.vpc)) &&
        // workerNodeType?: string
        (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
        true);
}
function isAWSRoute(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // DestinationCidrBlock?: string
        (typeof arg.DestinationCidrBlock === 'undefined' || typeof arg.DestinationCidrBlock === 'string') &&
        // GatewayId?: string
        (typeof arg.GatewayId === 'undefined' || typeof arg.GatewayId === 'string') &&
        // State?: string
        (typeof arg.State === 'undefined' || typeof arg.State === 'string') &&
        true);
}
function isAWSRouteTable(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // routes?: AWSRoute[]
        (typeof arg.routes === 'undefined' || (Array.isArray(arg.routes) && arg.routes.every((item) => isAWSRoute(item)))) &&
        // vpcId?: string
        (typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string') &&
        true);
}
function isAWSSubnet(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // availabilityZoneId?: string
        (typeof arg.availabilityZoneId === 'undefined' || typeof arg.availabilityZoneId === 'string') &&
        // availabilityZoneName?: string
        (typeof arg.availabilityZoneName === 'undefined' || typeof arg.availabilityZoneName === 'string') &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // isPublic: boolean
        (typeof arg.isPublic === 'boolean') &&
        // state?: string
        (typeof arg.state === 'undefined' || typeof arg.state === 'string') &&
        // vpcId?: string
        (typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string') &&
        true);
}
function isAWSVirtualMachine(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // osInfo?: OSInfo
        (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
        true);
}
function isAWSVpc(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // azs?: AWSNodeAz[]
        (typeof arg.azs === 'undefined' || (Array.isArray(arg.azs) && arg.azs.every((item) => isAWSNodeAz(item)))) &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // vpcID?: string
        (typeof arg.vpcID === 'undefined' || typeof arg.vpcID === 'string') &&
        true);
}
function isAzureAccountParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // azureCloud?: string
        (typeof arg.azureCloud === 'undefined' || typeof arg.azureCloud === 'string') &&
        // clientId?: string
        (typeof arg.clientId === 'undefined' || typeof arg.clientId === 'string') &&
        // clientSecret?: string
        (typeof arg.clientSecret === 'undefined' || typeof arg.clientSecret === 'string') &&
        // subscriptionId?: string
        (typeof arg.subscriptionId === 'undefined' || typeof arg.subscriptionId === 'string') &&
        // tenantId?: string
        (typeof arg.tenantId === 'undefined' || typeof arg.tenantId === 'string') &&
        true);
}
function isAzureInstanceType(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // family?: string
        (typeof arg.family === 'undefined' || typeof arg.family === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // size?: string
        (typeof arg.size === 'undefined' || typeof arg.size === 'string') &&
        // tier?: string
        (typeof arg.tier === 'undefined' || typeof arg.tier === 'string') &&
        // zones?: string[]
        (typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every((item) => typeof item === 'string'))) &&
        true);
}
function isAzureLocation(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // displayName?: string
        (typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isAzureRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // azureAccountParams?: AzureAccountParams
        (typeof arg.azureAccountParams === 'undefined' || isAzureAccountParams(arg.azureAccountParams)) &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // controlPlaneMachineType?: string
        (typeof arg.controlPlaneMachineType === 'undefined' || typeof arg.controlPlaneMachineType === 'string') &&
        // controlPlaneSubnet?: string
        (typeof arg.controlPlaneSubnet === 'undefined' || typeof arg.controlPlaneSubnet === 'string') &&
        // controlPlaneSubnetCidr?: string
        (typeof arg.controlPlaneSubnetCidr === 'undefined' || typeof arg.controlPlaneSubnetCidr === 'string') &&
        // enableAuditLogging?: boolean
        (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
        // frontendPrivateIp?: string
        (typeof arg.frontendPrivateIp === 'undefined' || typeof arg.frontendPrivateIp === 'string') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // isPrivateCluster?: boolean
        (typeof arg.isPrivateCluster === 'undefined' || typeof arg.isPrivateCluster === 'boolean') &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // location?: string
        (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNodes?: string
        (typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string') &&
        // os?: AzureVirtualMachine
        (typeof arg.os === 'undefined' || isAzureVirtualMachine(arg.os)) &&
        // resourceGroup?: string
        (typeof arg.resourceGroup === 'undefined' || typeof arg.resourceGroup === 'string') &&
        // sshPublicKey?: string
        (typeof arg.sshPublicKey === 'undefined' || typeof arg.sshPublicKey === 'string') &&
        // vnetCidr?: string
        (typeof arg.vnetCidr === 'undefined' || typeof arg.vnetCidr === 'string') &&
        // vnetName?: string
        (typeof arg.vnetName === 'undefined' || typeof arg.vnetName === 'string') &&
        // vnetResourceGroup?: string
        (typeof arg.vnetResourceGroup === 'undefined' || typeof arg.vnetResourceGroup === 'string') &&
        // workerMachineType?: string
        (typeof arg.workerMachineType === 'undefined' || typeof arg.workerMachineType === 'string') &&
        // workerNodeSubnet?: string
        (typeof arg.workerNodeSubnet === 'undefined' || typeof arg.workerNodeSubnet === 'string') &&
        // workerNodeSubnetCidr?: string
        (typeof arg.workerNodeSubnetCidr === 'undefined' || typeof arg.workerNodeSubnetCidr === 'string') &&
        true);
}
function isAzureResourceGroup(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // location: string
        (typeof arg.location === 'string') &&
        // name: string
        (typeof arg.name === 'string') &&
        true);
}
function isAzureSubnet(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isAzureVirtualMachine(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // osInfo?: OSInfo
        (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
        true);
}
function isAzureVirtualNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidrBlock: string
        (typeof arg.cidrBlock === 'string') &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        // location: string
        (typeof arg.location === 'string') &&
        // name: string
        (typeof arg.name === 'string') &&
        // subnets?: AzureSubnet[]
        (typeof arg.subnets === 'undefined' || (Array.isArray(arg.subnets) && arg.subnets.every((item) => isAzureSubnet(item)))) &&
        true);
}
function isConfigFile(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // filecontents?: string
        (typeof arg.filecontents === 'undefined' || typeof arg.filecontents === 'string') &&
        true);
}
function isConfigFileInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // path?: string
        (typeof arg.path === 'undefined' || typeof arg.path === 'string') &&
        true);
}
function isDockerDaemonStatus(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // status?: boolean
        (typeof arg.status === 'undefined' || typeof arg.status === 'boolean') &&
        true);
}
function isDockerRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNodes?: string
        (typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string') &&
        true);
}
function isError(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // message?: string
        (typeof arg.message === 'undefined' || typeof arg.message === 'string') &&
        true);
}
function isFeatureMap(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // [key: string]: string
        (Object.values(arg).every((value) => typeof value === 'string')) &&
        true);
}
function isFeatures(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // [key: string]: FeatureMap
        (Object.values(arg).every((value) => isFeatureMap(value))) &&
        true);
}
function isHTTPProxyConfiguration(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // enabled?: boolean
        (typeof arg.enabled === 'undefined' || typeof arg.enabled === 'boolean') &&
        // HTTPProxyPassword?: string
        (typeof arg.HTTPProxyPassword === 'undefined' || typeof arg.HTTPProxyPassword === 'string') &&
        // HTTPProxyURL?: string
        (typeof arg.HTTPProxyURL === 'undefined' || typeof arg.HTTPProxyURL === 'string') &&
        // HTTPProxyUsername?: string
        (typeof arg.HTTPProxyUsername === 'undefined' || typeof arg.HTTPProxyUsername === 'string') &&
        // HTTPSProxyPassword?: string
        (typeof arg.HTTPSProxyPassword === 'undefined' || typeof arg.HTTPSProxyPassword === 'string') &&
        // HTTPSProxyURL?: string
        (typeof arg.HTTPSProxyURL === 'undefined' || typeof arg.HTTPSProxyURL === 'string') &&
        // HTTPSProxyUsername?: string
        (typeof arg.HTTPSProxyUsername === 'undefined' || typeof arg.HTTPSProxyUsername === 'string') &&
        // noProxy?: string
        (typeof arg.noProxy === 'undefined' || typeof arg.noProxy === 'string') &&
        true);
}
function isIdentityManagementConfig(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // idm_type: 'oidc' | 'ldap' | 'none'
        (['oidc', 'ldap', 'none'].includes(arg.idm_type)) &&
        // ldap_bind_dn?: string
        (typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string') &&
        // ldap_bind_password?: string
        (typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string') &&
        // ldap_group_search_base_dn?: string
        (typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string') &&
        // ldap_group_search_filter?: string
        (typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string') &&
        // ldap_group_search_group_attr?: string
        (typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string') &&
        // ldap_group_search_name_attr?: string
        (typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string') &&
        // ldap_group_search_user_attr?: string
        (typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string') &&
        // ldap_root_ca?: string
        (typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string') &&
        // ldap_url?: string
        (typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string') &&
        // ldap_user_search_base_dn?: string
        (typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string') &&
        // ldap_user_search_email_attr?: string
        (typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string') &&
        // ldap_user_search_filter?: string
        (typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string') &&
        // ldap_user_search_id_attr?: string
        (typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string') &&
        // ldap_user_search_name_attr?: string
        (typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string') &&
        // ldap_user_search_username?: string
        (typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string') &&
        // oidc_claim_mappings?: { [key: string]: string }
        (typeof arg.oidc_claim_mappings === 'undefined' || typeof arg.oidc_claim_mappings === 'string') &&
        // oidc_client_id?: string
        (typeof arg.oidc_client_id === 'undefined' || typeof arg.oidc_client_id === 'string') &&
        // oidc_client_secret?: string
        (typeof arg.oidc_client_secret === 'undefined' || typeof arg.oidc_client_secret === 'string') &&
        // oidc_provider_name?: string
        (typeof arg.oidc_provider_name === 'undefined' || typeof arg.oidc_provider_name === 'string') &&
        // oidc_provider_url?: string
        (typeof arg.oidc_provider_url === 'undefined' || typeof arg.oidc_provider_url === 'string') &&
        // oidc_scope?: string
        (typeof arg.oidc_scope === 'undefined' || typeof arg.oidc_scope === 'string') &&
        // oidc_skip_verify_cert?: boolean
        (typeof arg.oidc_skip_verify_cert === 'undefined' || typeof arg.oidc_skip_verify_cert === 'boolean') &&
        true);
}
function isLdapParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // ldap_bind_dn?: string
        (typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string') &&
        // ldap_bind_password?: string
        (typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string') &&
        // ldap_group_search_base_dn?: string
        (typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string') &&
        // ldap_group_search_filter?: string
        (typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string') &&
        // ldap_group_search_group_attr?: string
        (typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string') &&
        // ldap_group_search_name_attr?: string
        (typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string') &&
        // ldap_group_search_user_attr?: string
        (typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string') &&
        // ldap_root_ca?: string
        (typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string') &&
        // ldap_test_group?: string
        (typeof arg.ldap_test_group === 'undefined' || typeof arg.ldap_test_group === 'string') &&
        // ldap_test_user?: string
        (typeof arg.ldap_test_user === 'undefined' || typeof arg.ldap_test_user === 'string') &&
        // ldap_url?: string
        (typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string') &&
        // ldap_user_search_base_dn?: string
        (typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string') &&
        // ldap_user_search_email_attr?: string
        (typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string') &&
        // ldap_user_search_filter?: string
        (typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string') &&
        // ldap_user_search_id_attr?: string
        (typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string') &&
        // ldap_user_search_name_attr?: string
        (typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string') &&
        // ldap_user_search_username?: string
        (typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string') &&
        true);
}
function isLdapTestResult(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // code?: number
        (typeof arg.code === 'undefined' || typeof arg.code === 'number') &&
        // desc?: string
        (typeof arg.desc === 'undefined' || typeof arg.desc === 'string') &&
        true);
}
function isNodeType(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cpu?: number
        (typeof arg.cpu === 'undefined' || typeof arg.cpu === 'number') &&
        // disk?: number
        (typeof arg.disk === 'undefined' || typeof arg.disk === 'number') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // ram?: number
        (typeof arg.ram === 'undefined' || typeof arg.ram === 'number') &&
        true);
}
function isOSInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // arch?: string
        (typeof arg.arch === 'undefined' || typeof arg.arch === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // version?: string
        (typeof arg.version === 'undefined' || typeof arg.version === 'string') &&
        true);
}
function isProviderInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // provider?: string
        (typeof arg.provider === 'undefined' || typeof arg.provider === 'string') &&
        // tkrVersion?: string
        (typeof arg.tkrVersion === 'undefined' || typeof arg.tkrVersion === 'string') &&
        true);
}
function isTKGNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // clusterDNSName?: string
        (typeof arg.clusterDNSName === 'undefined' || typeof arg.clusterDNSName === 'string') &&
        // clusterNodeCIDR?: string
        (typeof arg.clusterNodeCIDR === 'undefined' || typeof arg.clusterNodeCIDR === 'string') &&
        // clusterPodCIDR?: string
        (typeof arg.clusterPodCIDR === 'undefined' || typeof arg.clusterPodCIDR === 'string') &&
        // clusterServiceCIDR?: string
        (typeof arg.clusterServiceCIDR === 'undefined' || typeof arg.clusterServiceCIDR === 'string') &&
        // cniType?: string
        (typeof arg.cniType === 'undefined' || typeof arg.cniType === 'string') &&
        // httpProxyConfiguration?: HTTPProxyConfiguration
        (typeof arg.httpProxyConfiguration === 'undefined' || isHTTPProxyConfiguration(arg.httpProxyConfiguration)) &&
        // networkName?: string
        (typeof arg.networkName === 'undefined' || typeof arg.networkName === 'string') &&
        true);
}
function isVpc(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // cidr?: string
        (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
        // id?: string
        (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
        true);
}
function isVSphereAvailabilityZone(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isVSphereCredentials(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // host?: string
        (typeof arg.host === 'undefined' || typeof arg.host === 'string') &&
        // insecure?: boolean
        (typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean') &&
        // password?: string
        (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
        // thumbprint?: string
        (typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string') &&
        // username?: string
        (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
        true);
}
function isVSphereDatacenter(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isVSphereDatastore(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isVSphereFolder(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isVsphereInfo(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // hasPacific?: string
        (typeof arg.hasPacific === 'undefined' || typeof arg.hasPacific === 'string') &&
        // version?: string
        (typeof arg.version === 'undefined' || typeof arg.version === 'string') &&
        true);
}
function isVSphereManagementObject(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // parentMoid?: string
        (typeof arg.parentMoid === 'undefined' || typeof arg.parentMoid === 'string') &&
        // path?: string
        (typeof arg.path === 'undefined' || typeof arg.path === 'string') &&
        // resourceType?: 'datacenter' | 'cluster' | 'hostgroup' | 'folder' | 'respool' | 'vm' | 'datastore' | 'host' | 'network'
        (typeof arg.resourceType === 'undefined' || ['datacenter', 'cluster', 'hostgroup', 'folder', 'respool', 'vm', 'datastore', 'host', 'network'].includes(arg.resourceType)) &&
        true);
}
function isVSphereNetwork(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // displayName?: string
        (typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string') &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isVSphereRegion(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // zones?: VSphereAvailabilityZone[]
        (typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every((item) => isVSphereAvailabilityZone(item)))) &&
        true);
}
function isVsphereRegionalClusterParams(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // annotations?: { [key: string]: string }
        (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
        // aviConfig?: AviConfig
        (typeof arg.aviConfig === 'undefined' || isAviConfig(arg.aviConfig)) &&
        // ceipOptIn?: boolean
        (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
        // clusterName?: string
        (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
        // controlPlaneEndpoint?: string
        (typeof arg.controlPlaneEndpoint === 'undefined' || typeof arg.controlPlaneEndpoint === 'string') &&
        // controlPlaneFlavor?: string
        (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
        // controlPlaneNodeType?: string
        (typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string') &&
        // datacenter?: string
        (typeof arg.datacenter === 'undefined' || typeof arg.datacenter === 'string') &&
        // datastore?: string
        (typeof arg.datastore === 'undefined' || typeof arg.datastore === 'string') &&
        // enableAuditLogging?: boolean
        (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
        // folder?: string
        (typeof arg.folder === 'undefined' || typeof arg.folder === 'string') &&
        // identityManagement?: IdentityManagementConfig
        (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
        // ipFamily?: string
        (typeof arg.ipFamily === 'undefined' || typeof arg.ipFamily === 'string') &&
        // kubernetesVersion?: string
        (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
        // labels?: { [key: string]: string }
        (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
        // machineHealthCheckEnabled?: boolean
        (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
        // networking?: TKGNetwork
        (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
        // numOfWorkerNode?: number
        (typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number') &&
        // os?: VSphereVirtualMachine
        (typeof arg.os === 'undefined' || isVSphereVirtualMachine(arg.os)) &&
        // resourcePool?: string
        (typeof arg.resourcePool === 'undefined' || typeof arg.resourcePool === 'string') &&
        // ssh_key?: string
        (typeof arg.ssh_key === 'undefined' || typeof arg.ssh_key === 'string') &&
        // vsphereCredentials?: VSphereCredentials
        (typeof arg.vsphereCredentials === 'undefined' || isVSphereCredentials(arg.vsphereCredentials)) &&
        // workerNodeType?: string
        (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
        true);
}
function isVSphereResourcePool(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        true);
}
function isVSphereThumbprint(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // insecure?: boolean
        (typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean') &&
        // thumbprint?: string
        (typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string') &&
        true);
}
function isVSphereVirtualMachine(arg) {
    return (arg != null &&
        typeof arg === 'object' &&
        // isTemplate: boolean
        (typeof arg.isTemplate === 'boolean') &&
        // k8sVersion?: string
        (typeof arg.k8sVersion === 'undefined' || typeof arg.k8sVersion === 'string') &&
        // moid?: string
        (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
        // name?: string
        (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
        // osInfo?: OSInfo
        (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
        true);
}

/**
 * Created with https://github.com/flowup/api-client-generator
 */
class GuardedAPIClient extends APIClient {
    constructor(httpClient, domain, options) {
        super(httpClient, domain, options);
        this.httpClient = httpClient;
    }
    getUI(requestHttpOptions) {
        return super.getUI(requestHttpOptions)
            .pipe(tap((res) => isFile(res) || console.error(`TypeGuard for response 'File' caught inconsistency.`, res)));
    }
    getUIFile(args, requestHttpOptions) {
        return super.getUIFile(args, requestHttpOptions)
            .pipe(tap((res) => isFile(res) || console.error(`TypeGuard for response 'File' caught inconsistency.`, res)));
    }
    getFeatureFlags(requestHttpOptions) {
        return super.getFeatureFlags(requestHttpOptions)
            .pipe(tap((res) => isFeatures(res) || console.error(`TypeGuard for response 'Features' caught inconsistency.`, res)));
    }
    getTanzuEdition(requestHttpOptions) {
        return super.getTanzuEdition(requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    verifyLdapConnect(args, requestHttpOptions) {
        return super.verifyLdapConnect(args, requestHttpOptions)
            .pipe(tap((res) => isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
    }
    verifyLdapBind(requestHttpOptions) {
        return super.verifyLdapBind(requestHttpOptions)
            .pipe(tap((res) => isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
    }
    verifyLdapUserSearch(requestHttpOptions) {
        return super.verifyLdapUserSearch(requestHttpOptions)
            .pipe(tap((res) => isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
    }
    verifyLdapGroupSearch(requestHttpOptions) {
        return super.verifyLdapGroupSearch(requestHttpOptions)
            .pipe(tap((res) => isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
    }
    getAviClouds(requestHttpOptions) {
        return super.getAviClouds(requestHttpOptions)
            .pipe(tap((res) => isAviCloud(res) || console.error(`TypeGuard for response 'AviCloud' caught inconsistency.`, res)));
    }
    getAviServiceEngineGroups(requestHttpOptions) {
        return super.getAviServiceEngineGroups(requestHttpOptions)
            .pipe(tap((res) => isAviServiceEngineGroup(res) || console.error(`TypeGuard for response 'AviServiceEngineGroup' caught inconsistency.`, res)));
    }
    getAviVipNetworks(requestHttpOptions) {
        return super.getAviVipNetworks(requestHttpOptions)
            .pipe(tap((res) => isAviVipNetwork(res) || console.error(`TypeGuard for response 'AviVipNetwork' caught inconsistency.`, res)));
    }
    getProvider(requestHttpOptions) {
        return super.getProvider(requestHttpOptions)
            .pipe(tap((res) => isProviderInfo(res) || console.error(`TypeGuard for response 'ProviderInfo' caught inconsistency.`, res)));
    }
    getVsphereThumbprint(args, requestHttpOptions) {
        return super.getVsphereThumbprint(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereThumbprint(res) || console.error(`TypeGuard for response 'VSphereThumbprint' caught inconsistency.`, res)));
    }
    setVSphereEndpoint(args, requestHttpOptions) {
        return super.setVSphereEndpoint(args, requestHttpOptions)
            .pipe(tap((res) => isVsphereInfo(res) || console.error(`TypeGuard for response 'VsphereInfo' caught inconsistency.`, res)));
    }
    getVSphereDatacenters(requestHttpOptions) {
        return super.getVSphereDatacenters(requestHttpOptions)
            .pipe(tap((res) => isVSphereDatacenter(res) || console.error(`TypeGuard for response 'VSphereDatacenter' caught inconsistency.`, res)));
    }
    getVSphereDatastores(args, requestHttpOptions) {
        return super.getVSphereDatastores(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereDatastore(res) || console.error(`TypeGuard for response 'VSphereDatastore' caught inconsistency.`, res)));
    }
    getVSphereFolders(args, requestHttpOptions) {
        return super.getVSphereFolders(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereFolder(res) || console.error(`TypeGuard for response 'VSphereFolder' caught inconsistency.`, res)));
    }
    getVSphereComputeResources(args, requestHttpOptions) {
        return super.getVSphereComputeResources(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereManagementObject(res) || console.error(`TypeGuard for response 'VSphereManagementObject' caught inconsistency.`, res)));
    }
    getVSphereResourcePools(args, requestHttpOptions) {
        return super.getVSphereResourcePools(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereResourcePool(res) || console.error(`TypeGuard for response 'VSphereResourcePool' caught inconsistency.`, res)));
    }
    getVSphereNetworks(args, requestHttpOptions) {
        return super.getVSphereNetworks(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereNetwork(res) || console.error(`TypeGuard for response 'VSphereNetwork' caught inconsistency.`, res)));
    }
    getVSphereNodeTypes(requestHttpOptions) {
        return super.getVSphereNodeTypes(requestHttpOptions)
            .pipe(tap((res) => isNodeType(res) || console.error(`TypeGuard for response 'NodeType' caught inconsistency.`, res)));
    }
    getVSphereOSImages(args, requestHttpOptions) {
        return super.getVSphereOSImages(args, requestHttpOptions)
            .pipe(tap((res) => isVSphereVirtualMachine(res) || console.error(`TypeGuard for response 'VSphereVirtualMachine' caught inconsistency.`, res)));
    }
    exportTKGConfigForVsphere(args, requestHttpOptions) {
        return super.exportTKGConfigForVsphere(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    applyTKGConfigForVsphere(args, requestHttpOptions) {
        return super.applyTKGConfigForVsphere(args, requestHttpOptions)
            .pipe(tap((res) => isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
    }
    importTKGConfigForVsphere(args, requestHttpOptions) {
        return super.importTKGConfigForVsphere(args, requestHttpOptions)
            .pipe(tap((res) => isVsphereRegionalClusterParams(res) || console.error(`TypeGuard for response 'VsphereRegionalClusterParams' caught inconsistency.`, res)));
    }
    createVSphereRegionalCluster(args, requestHttpOptions) {
        return super.createVSphereRegionalCluster(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    getVPCs(requestHttpOptions) {
        return super.getVPCs(requestHttpOptions)
            .pipe(tap((res) => isVpc(res) || console.error(`TypeGuard for response 'Vpc' caught inconsistency.`, res)));
    }
    getAWSNodeTypes(args, requestHttpOptions) {
        return super.getAWSNodeTypes(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    getAWSRegions(requestHttpOptions) {
        return super.getAWSRegions(requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    getAWSOSImages(args, requestHttpOptions) {
        return super.getAWSOSImages(args, requestHttpOptions)
            .pipe(tap((res) => isAWSVirtualMachine(res) || console.error(`TypeGuard for response 'AWSVirtualMachine' caught inconsistency.`, res)));
    }
    getAWSCredentialProfiles(requestHttpOptions) {
        return super.getAWSCredentialProfiles(requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    getAWSAvailabilityZones(requestHttpOptions) {
        return super.getAWSAvailabilityZones(requestHttpOptions)
            .pipe(tap((res) => isAWSAvailabilityZone(res) || console.error(`TypeGuard for response 'AWSAvailabilityZone' caught inconsistency.`, res)));
    }
    getAWSSubnets(args, requestHttpOptions) {
        return super.getAWSSubnets(args, requestHttpOptions)
            .pipe(tap((res) => isAWSSubnet(res) || console.error(`TypeGuard for response 'AWSSubnet' caught inconsistency.`, res)));
    }
    exportTKGConfigForAWS(args, requestHttpOptions) {
        return super.exportTKGConfigForAWS(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    applyTKGConfigForAWS(args, requestHttpOptions) {
        return super.applyTKGConfigForAWS(args, requestHttpOptions)
            .pipe(tap((res) => isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
    }
    createAWSRegionalCluster(args, requestHttpOptions) {
        return super.createAWSRegionalCluster(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    importTKGConfigForAWS(args, requestHttpOptions) {
        return super.importTKGConfigForAWS(args, requestHttpOptions)
            .pipe(tap((res) => isAWSRegionalClusterParams(res) || console.error(`TypeGuard for response 'AWSRegionalClusterParams' caught inconsistency.`, res)));
    }
    getAzureEndpoint(requestHttpOptions) {
        return super.getAzureEndpoint(requestHttpOptions)
            .pipe(tap((res) => isAzureAccountParams(res) || console.error(`TypeGuard for response 'AzureAccountParams' caught inconsistency.`, res)));
    }
    getAzureResourceGroups(args, requestHttpOptions) {
        return super.getAzureResourceGroups(args, requestHttpOptions)
            .pipe(tap((res) => isAzureResourceGroup(res) || console.error(`TypeGuard for response 'AzureResourceGroup' caught inconsistency.`, res)));
    }
    createAzureResourceGroup(args, requestHttpOptions) {
        return super.createAzureResourceGroup(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    getAzureVnets(args, requestHttpOptions) {
        return super.getAzureVnets(args, requestHttpOptions)
            .pipe(tap((res) => isAzureVirtualNetwork(res) || console.error(`TypeGuard for response 'AzureVirtualNetwork' caught inconsistency.`, res)));
    }
    createAzureVirtualNetwork(args, requestHttpOptions) {
        return super.createAzureVirtualNetwork(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    getAzureOSImages(requestHttpOptions) {
        return super.getAzureOSImages(requestHttpOptions)
            .pipe(tap((res) => isAzureVirtualMachine(res) || console.error(`TypeGuard for response 'AzureVirtualMachine' caught inconsistency.`, res)));
    }
    getAzureRegions(requestHttpOptions) {
        return super.getAzureRegions(requestHttpOptions)
            .pipe(tap((res) => isAzureLocation(res) || console.error(`TypeGuard for response 'AzureLocation' caught inconsistency.`, res)));
    }
    getAzureInstanceTypes(args, requestHttpOptions) {
        return super.getAzureInstanceTypes(args, requestHttpOptions)
            .pipe(tap((res) => isAzureInstanceType(res) || console.error(`TypeGuard for response 'AzureInstanceType' caught inconsistency.`, res)));
    }
    exportTKGConfigForAzure(args, requestHttpOptions) {
        return super.exportTKGConfigForAzure(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    applyTKGConfigForAzure(args, requestHttpOptions) {
        return super.applyTKGConfigForAzure(args, requestHttpOptions)
            .pipe(tap((res) => isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
    }
    createAzureRegionalCluster(args, requestHttpOptions) {
        return super.createAzureRegionalCluster(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    importTKGConfigForAzure(args, requestHttpOptions) {
        return super.importTKGConfigForAzure(args, requestHttpOptions)
            .pipe(tap((res) => isAzureRegionalClusterParams(res) || console.error(`TypeGuard for response 'AzureRegionalClusterParams' caught inconsistency.`, res)));
    }
    checkIfDockerDaemonAvailable(requestHttpOptions) {
        return super.checkIfDockerDaemonAvailable(requestHttpOptions)
            .pipe(tap((res) => isDockerDaemonStatus(res) || console.error(`TypeGuard for response 'DockerDaemonStatus' caught inconsistency.`, res)));
    }
    exportTKGConfigForDocker(args, requestHttpOptions) {
        return super.exportTKGConfigForDocker(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    applyTKGConfigForDocker(args, requestHttpOptions) {
        return super.applyTKGConfigForDocker(args, requestHttpOptions)
            .pipe(tap((res) => isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
    }
    createDockerRegionalCluster(args, requestHttpOptions) {
        return super.createDockerRegionalCluster(args, requestHttpOptions)
            .pipe(tap((res) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
    }
    importTKGConfigForDocker(args, requestHttpOptions) {
        return super.importTKGConfigForDocker(args, requestHttpOptions)
            .pipe(tap((res) => isDockerRegionalClusterParams(res) || console.error(`TypeGuard for response 'DockerRegionalClusterParams' caught inconsistency.`, res)));
    }
}
GuardedAPIClient.ɵfac = i0.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: GuardedAPIClient, deps: [{ token: i1.HttpClient }, { token: USE_DOMAIN, optional: true }, { token: USE_HTTP_OPTIONS, optional: true }], target: i0.ɵɵFactoryTarget.Injectable });
GuardedAPIClient.ɵprov = i0.ɵɵngDeclareInjectable({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: GuardedAPIClient });
i0.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: GuardedAPIClient, decorators: [{
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

/* tslint:disable */
class APIClientModule {
    /**
     * Use this method in your root module to provide the APIClientModule
     *
     * If you are not providing
     * @param { APIClientModuleConfig } config
     * @returns { ModuleWithProviders }
     */
    static forRoot(config = {}) {
        return {
            ngModule: APIClientModule,
            providers: [
                ...(config.domain != null ? [{ provide: USE_DOMAIN, useValue: config.domain }] : []),
                ...(config.httpOptions ? [{ provide: USE_HTTP_OPTIONS, useValue: config.httpOptions }] : []),
                ...(config.guardResponses ? [{ provide: APIClient, useClass: GuardedAPIClient }] : [APIClient]),
            ]
        };
    }
}
APIClientModule.ɵfac = i0.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule, deps: [], target: i0.ɵɵFactoryTarget.NgModule });
APIClientModule.ɵmod = i0.ɵɵngDeclareNgModule({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule });
APIClientModule.ɵinj = i0.ɵɵngDeclareInjector({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule });
i0.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0, type: APIClientModule, decorators: [{
            type: NgModule,
            args: [{}]
        }] });

/* tslint:disable */

/*
 * Public API Surface of tanzu-ui-api-lib
 * Exports swagger generated APIClient and modules, and swagger generated models
 */

/**
 * Generated bundle index. Do not edit.
 */

export { APIClient, APIClientModule, GuardedAPIClient };
//# sourceMappingURL=tanzu-ui-api-lib.js.map
