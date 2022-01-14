/* tslint:disable */

import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Inject, Injectable, InjectionToken, Optional } from '@angular/core';
import { Observable, throwError } from 'rxjs';
import { DefaultHttpOptions, HttpOptions, APIClientInterface } from './';

import * as models from './models';

export const USE_DOMAIN = new InjectionToken<string>('APIClient_USE_DOMAIN');
export const USE_HTTP_OPTIONS = new InjectionToken<HttpOptions>('APIClient_USE_HTTP_OPTIONS');

type APIHttpOptions = HttpOptions & {
  headers: HttpHeaders;
  params: HttpParams;
  responseType?: 'arraybuffer' | 'blob' | 'text' | 'json';
};

/**
 * Created with https://github.com/flowup/api-client-generator
 */
@Injectable()
export class APIClient implements APIClientInterface {

  readonly options: APIHttpOptions;

  readonly domain: string = `//${window.location.hostname}${window.location.port ? ':'+window.location.port : ''}`;

  constructor(private readonly http: HttpClient,
              @Optional() @Inject(USE_DOMAIN) domain?: string,
              @Optional() @Inject(USE_HTTP_OPTIONS) options?: DefaultHttpOptions) {

    if (domain != null) {
      this.domain = domain;
    }

    this.options = {
      headers: new HttpHeaders(options && options.headers ? options.headers : {}),
      params: new HttpParams(options && options.params ? options.params : {}),
      ...(options && options.reportProgress ? { reportProgress: options.reportProgress } : {}),
      ...(options && options.withCredentials ? { withCredentials: options.withCredentials } : {})
    };
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getUI(
    requestHttpOptions?: HttpOptions
  ): Observable<File> {
    const path = `/`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
      responseType: 'blob',
    };

    return this.sendRequest<File>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getUIFile(
    args: {
      filename: string,  // UI file name
    },
    requestHttpOptions?: HttpOptions
  ): Observable<File> {
    const path = `/${args.filename}`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
      responseType: 'blob',
    };

    return this.sendRequest<File>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getFeatureFlags(
    requestHttpOptions?: HttpOptions
  ): Observable<models.Features> {
    const path = `/api/features`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.Features>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getTanzuEdition(
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/edition`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('GET', path, options);
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  verifyAccount(
    args: {
      credentials?: models.AviControllerParams,  // (optional) Avi controller credentials
    },
    requestHttpOptions?: HttpOptions
  ): Observable<void> {
    const path = `/api/avi`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<void>('POST', path, options, JSON.stringify(args.credentials));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapConnect(
    args: {
      credentials?: models.LdapParams,  // (optional) LDAP configuration
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    const path = `/api/ldap/connect`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.LdapTestResult>('POST', path, options, JSON.stringify(args.credentials));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapBind(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    const path = `/api/ldap/bind`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.LdapTestResult>('POST', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapUserSearch(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    const path = `/api/ldap/users/search`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.LdapTestResult>('POST', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapGroupSearch(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    const path = `/api/ldap/groups/search`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.LdapTestResult>('POST', path, options);
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  verifyLdapCloseConnection(
    requestHttpOptions?: HttpOptions
  ): Observable<void> {
    const path = `/api/ldap/disconnect`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<void>('POST', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAviClouds(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviCloud[]> {
    const path = `/api/avi/clouds`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AviCloud[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAviServiceEngineGroups(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviServiceEngineGroup[]> {
    const path = `/api/avi/serviceenginegroups`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AviServiceEngineGroup[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAviVipNetworks(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviVipNetwork[]> {
    const path = `/api/avi/vipnetworks`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AviVipNetwork[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getProvider(
    requestHttpOptions?: HttpOptions
  ): Observable<models.ProviderInfo> {
    const path = `/api/providers`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.ProviderInfo>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVsphereThumbprint(
    args: {
      host: string,  // vSphere host
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereThumbprint> {
    const path = `/api/providers/vsphere/thumbprint`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('host' in args) {
      options.params = options.params.set('host', String(args.host));
    }
    return this.sendRequest<models.VSphereThumbprint>('GET', path, options);
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  setVSphereEndpoint(
    args: {
      credentials?: models.VSphereCredentials,  // (optional) vSphere credentials
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VsphereInfo> {
    const path = `/api/providers/vsphere`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.VsphereInfo>('POST', path, options, JSON.stringify(args.credentials));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereDatacenters(
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereDatacenter[]> {
    const path = `/api/providers/vsphere/datacenters`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.VSphereDatacenter[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereDatastores(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereDatastore[]> {
    const path = `/api/providers/vsphere/datastores`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('dc' in args) {
      options.params = options.params.set('dc', String(args.dc));
    }
    return this.sendRequest<models.VSphereDatastore[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereFolders(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereFolder[]> {
    const path = `/api/providers/vsphere/folders`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('dc' in args) {
      options.params = options.params.set('dc', String(args.dc));
    }
    return this.sendRequest<models.VSphereFolder[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereComputeResources(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereManagementObject[]> {
    const path = `/api/providers/vsphere/compute`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('dc' in args) {
      options.params = options.params.set('dc', String(args.dc));
    }
    return this.sendRequest<models.VSphereManagementObject[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereResourcePools(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereResourcePool[]> {
    const path = `/api/providers/vsphere/resourcepools`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('dc' in args) {
      options.params = options.params.set('dc', String(args.dc));
    }
    return this.sendRequest<models.VSphereResourcePool[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereNetworks(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereNetwork[]> {
    const path = `/api/providers/vsphere/networks`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('dc' in args) {
      options.params = options.params.set('dc', String(args.dc));
    }
    return this.sendRequest<models.VSphereNetwork[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereNodeTypes(
    requestHttpOptions?: HttpOptions
  ): Observable<models.NodeType[]> {
    const path = `/api/providers/vsphere/nodetypes`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.NodeType[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereOSImages(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereVirtualMachine[]> {
    const path = `/api/providers/vsphere/osimages`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('dc' in args) {
      options.params = options.params.set('dc', String(args.dc));
    }
    return this.sendRequest<models.VSphereVirtualMachine[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForVsphere(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to generate tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/vsphere/config/export`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForVsphere(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to apply changes to tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    const path = `/api/providers/vsphere/tkgconfig`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.ConfigFileInfo>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForVsphere(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VsphereRegionalClusterParams> {
    const path = `/api/providers/vsphere/config/import`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.VsphereRegionalClusterParams>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createVSphereRegionalCluster(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/vsphere/create`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  setAWSEndpoint(
    args: {
      accountParams?: models.AWSAccountParams,  // (optional) AWS account parameters
    },
    requestHttpOptions?: HttpOptions
  ): Observable<void> {
    const path = `/api/providers/aws`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<void>('POST', path, options, JSON.stringify(args.accountParams));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVPCs(
    requestHttpOptions?: HttpOptions
  ): Observable<models.Vpc[]> {
    const path = `/api/providers/aws/vpc`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.Vpc[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSNodeTypes(
    args: {
      az?: string,  // (optional) AWS availability zone, e.g. us-west-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string[]> {
    const path = `/api/providers/aws/nodetypes`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('az' in args) {
      options.params = options.params.set('az', String(args.az));
    }
    return this.sendRequest<string[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSRegions(
    requestHttpOptions?: HttpOptions
  ): Observable<string[]> {
    const path = `/api/providers/aws/regions`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSOSImages(
    args: {
      region: string,  // AWS region, e.g. us-west-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSVirtualMachine[]> {
    const path = `/api/providers/aws/osimages`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('region' in args) {
      options.params = options.params.set('region', String(args.region));
    }
    return this.sendRequest<models.AWSVirtualMachine[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSCredentialProfiles(
    requestHttpOptions?: HttpOptions
  ): Observable<string[]> {
    const path = `/api/providers/aws/profiles`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSAvailabilityZones(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSAvailabilityZone[]> {
    const path = `/api/providers/aws/AvailabilityZones`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AWSAvailabilityZone[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSSubnets(
    args: {
      vpcId: string,  // VPC Id
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSSubnet[]> {
    const path = `/api/providers/aws/subnets`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('vpcId' in args) {
      options.params = options.params.set('vpcId', String(args.vpcId));
    }
    return this.sendRequest<models.AWSSubnet[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForAWS(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to generate TKG configuration file for AWS
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/aws/config/export`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForAWS(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to apply changes to TKG configuration file for AWS
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    const path = `/api/providers/aws/tkgconfig`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.ConfigFileInfo>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createAWSRegionalCluster(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/aws/create`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForAWS(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for aws
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSRegionalClusterParams> {
    const path = `/api/providers/aws/config/import`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AWSRegionalClusterParams>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureEndpoint(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureAccountParams> {
    const path = `/api/providers/azure`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AzureAccountParams>('GET', path, options);
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  setAzureEndpoint(
    args: {
      accountParams?: models.AzureAccountParams,  // (optional) Azure account parameters
    },
    requestHttpOptions?: HttpOptions
  ): Observable<void> {
    const path = `/api/providers/azure`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<void>('POST', path, options, JSON.stringify(args.accountParams));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureResourceGroups(
    args: {
      location: string,  // Azure region
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureResourceGroup[]> {
    const path = `/api/providers/azure/resourcegroups`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('location' in args) {
      options.params = options.params.set('location', String(args.location));
    }
    return this.sendRequest<models.AzureResourceGroup[]>('GET', path, options);
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  createAzureResourceGroup(
    args: {
      params: models.AzureResourceGroup,  // parameters to create a new Azure resource group
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/azure/resourcegroups`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureVnets(
    args: {
      resourceGroupName: string,  // Name of the Azure resource group
      location: string,  // Azure region
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureVirtualNetwork[]> {
    const path = `/api/providers/azure/resourcegroups/${args.resourceGroupName}/vnets`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    if ('location' in args) {
      options.params = options.params.set('location', String(args.location));
    }
    return this.sendRequest<models.AzureVirtualNetwork[]>('GET', path, options);
  }

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  createAzureVirtualNetwork(
    args: {
      resourceGroupName: string,  // Name of the Azure resource group
      params: models.AzureVirtualNetwork,  // parameters to create a new Azure Virtual network
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/azure/resourcegroups/${args.resourceGroupName}/vnets`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureOSImages(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureVirtualMachine[]> {
    const path = `/api/providers/azure/osimages`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AzureVirtualMachine[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureRegions(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureLocation[]> {
    const path = `/api/providers/azure/regions`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AzureLocation[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureInstanceTypes(
    args: {
      location: string,  // Azure region name
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureInstanceType[]> {
    const path = `/api/providers/azure/regions/${args.location}/instanceTypes`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AzureInstanceType[]>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForAzure(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to generate TKG configuration file for Azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/azure/config/export`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForAzure(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to apply changes to TKG configuration file for Azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    const path = `/api/providers/azure/tkgconfig`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.ConfigFileInfo>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createAzureRegionalCluster(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/azure/create`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForAzure(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureRegionalClusterParams> {
    const path = `/api/providers/azure/config/import`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.AzureRegionalClusterParams>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  checkIfDockerDaemonAvailable(
    requestHttpOptions?: HttpOptions
  ): Observable<models.DockerDaemonStatus> {
    const path = `/api/providers/docker/daemon`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.DockerDaemonStatus>('GET', path, options);
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForDocker(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to generate TKG configuration file for Docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/docker/config/export`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForDocker(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to apply changes to TKG configuration file for Docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    const path = `/api/providers/docker/tkgconfig`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.ConfigFileInfo>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createDockerRegionalCluster(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    const path = `/api/providers/docker/create`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<string>('POST', path, options, JSON.stringify(args.params));
  }

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForDocker(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.DockerRegionalClusterParams> {
    const path = `/api/providers/docker/config/import`;
    const options: APIHttpOptions = {
      ...this.options,
      ...requestHttpOptions,
    };

    return this.sendRequest<models.DockerRegionalClusterParams>('POST', path, options, JSON.stringify(args.params));
  }

  private sendRequest<T>(method: string, path: string, options: HttpOptions, body?: any): Observable<T> {
    switch (method) {
      case 'DELETE':
        return this.http.delete<T>(`${this.domain}${path}`, options);
      case 'GET':
        return this.http.get<T>(`${this.domain}${path}`, options);
      case 'HEAD':
        return this.http.head<T>(`${this.domain}${path}`, options);
      case 'OPTIONS':
        return this.http.options<T>(`${this.domain}${path}`, options);
      case 'PATCH':
        return this.http.patch<T>(`${this.domain}${path}`, body, options);
      case 'POST':
        return this.http.post<T>(`${this.domain}${path}`, body, options);
      case 'PUT':
        return this.http.put<T>(`${this.domain}${path}`, body, options);
      default:
        console.error(`Unsupported request: ${method}`);
        return throwError(`Unsupported request: ${method}`);
    }
  }
}
