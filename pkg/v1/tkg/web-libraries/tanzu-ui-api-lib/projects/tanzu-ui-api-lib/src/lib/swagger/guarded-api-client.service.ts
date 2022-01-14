/* tslint:disable */

import { HttpClient } from '@angular/common/http';
import { Inject, Injectable, Optional } from '@angular/core';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { DefaultHttpOptions, HttpOptions } from './';
import { USE_DOMAIN, USE_HTTP_OPTIONS, APIClient } from './api-client.service';

import * as models from './models';
import * as guards from './guards';

/**
 * Created with https://github.com/flowup/api-client-generator
 */
@Injectable()
export class GuardedAPIClient extends APIClient {

  constructor(readonly httpClient: HttpClient,
              @Optional() @Inject(USE_DOMAIN) domain?: string,
              @Optional() @Inject(USE_HTTP_OPTIONS) options?: DefaultHttpOptions) {
    super(httpClient, domain, options);
  }

  getUI(
    requestHttpOptions?: HttpOptions
  ): Observable<File> {
    return super.getUI(requestHttpOptions)
      .pipe(tap((res: any) => guards.isFile(res) || console.error(`TypeGuard for response 'File' caught inconsistency.`, res)));
  }

  getUIFile(
    args: {
      filename: string,  // UI file name
    },
    requestHttpOptions?: HttpOptions
  ): Observable<File> {
    return super.getUIFile(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isFile(res) || console.error(`TypeGuard for response 'File' caught inconsistency.`, res)));
  }

  getFeatureFlags(
    requestHttpOptions?: HttpOptions
  ): Observable<models.Features> {
    return super.getFeatureFlags(requestHttpOptions)
      .pipe(tap((res: any) => guards.isFeatures(res) || console.error(`TypeGuard for response 'Features' caught inconsistency.`, res)));
  }

  getTanzuEdition(
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.getTanzuEdition(requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  verifyLdapConnect(
    args: {
      credentials?: models.LdapParams,  // (optional) LDAP configuration
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    return super.verifyLdapConnect(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
  }

  verifyLdapBind(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    return super.verifyLdapBind(requestHttpOptions)
      .pipe(tap((res: any) => guards.isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
  }

  verifyLdapUserSearch(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    return super.verifyLdapUserSearch(requestHttpOptions)
      .pipe(tap((res: any) => guards.isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
  }

  verifyLdapGroupSearch(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult> {
    return super.verifyLdapGroupSearch(requestHttpOptions)
      .pipe(tap((res: any) => guards.isLdapTestResult(res) || console.error(`TypeGuard for response 'LdapTestResult' caught inconsistency.`, res)));
  }

  getAviClouds(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviCloud[]> {
    return super.getAviClouds(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAviCloud(res) || console.error(`TypeGuard for response 'AviCloud' caught inconsistency.`, res)));
  }

  getAviServiceEngineGroups(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviServiceEngineGroup[]> {
    return super.getAviServiceEngineGroups(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAviServiceEngineGroup(res) || console.error(`TypeGuard for response 'AviServiceEngineGroup' caught inconsistency.`, res)));
  }

  getAviVipNetworks(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviVipNetwork[]> {
    return super.getAviVipNetworks(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAviVipNetwork(res) || console.error(`TypeGuard for response 'AviVipNetwork' caught inconsistency.`, res)));
  }

  getProvider(
    requestHttpOptions?: HttpOptions
  ): Observable<models.ProviderInfo> {
    return super.getProvider(requestHttpOptions)
      .pipe(tap((res: any) => guards.isProviderInfo(res) || console.error(`TypeGuard for response 'ProviderInfo' caught inconsistency.`, res)));
  }

  getVsphereThumbprint(
    args: {
      host: string,  // vSphere host
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereThumbprint> {
    return super.getVsphereThumbprint(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereThumbprint(res) || console.error(`TypeGuard for response 'VSphereThumbprint' caught inconsistency.`, res)));
  }

  setVSphereEndpoint(
    args: {
      credentials?: models.VSphereCredentials,  // (optional) vSphere credentials
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VsphereInfo> {
    return super.setVSphereEndpoint(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVsphereInfo(res) || console.error(`TypeGuard for response 'VsphereInfo' caught inconsistency.`, res)));
  }

  getVSphereDatacenters(
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereDatacenter[]> {
    return super.getVSphereDatacenters(requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereDatacenter(res) || console.error(`TypeGuard for response 'VSphereDatacenter' caught inconsistency.`, res)));
  }

  getVSphereDatastores(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereDatastore[]> {
    return super.getVSphereDatastores(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereDatastore(res) || console.error(`TypeGuard for response 'VSphereDatastore' caught inconsistency.`, res)));
  }

  getVSphereFolders(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereFolder[]> {
    return super.getVSphereFolders(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereFolder(res) || console.error(`TypeGuard for response 'VSphereFolder' caught inconsistency.`, res)));
  }

  getVSphereComputeResources(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereManagementObject[]> {
    return super.getVSphereComputeResources(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereManagementObject(res) || console.error(`TypeGuard for response 'VSphereManagementObject' caught inconsistency.`, res)));
  }

  getVSphereResourcePools(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereResourcePool[]> {
    return super.getVSphereResourcePools(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereResourcePool(res) || console.error(`TypeGuard for response 'VSphereResourcePool' caught inconsistency.`, res)));
  }

  getVSphereNetworks(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereNetwork[]> {
    return super.getVSphereNetworks(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereNetwork(res) || console.error(`TypeGuard for response 'VSphereNetwork' caught inconsistency.`, res)));
  }

  getVSphereNodeTypes(
    requestHttpOptions?: HttpOptions
  ): Observable<models.NodeType[]> {
    return super.getVSphereNodeTypes(requestHttpOptions)
      .pipe(tap((res: any) => guards.isNodeType(res) || console.error(`TypeGuard for response 'NodeType' caught inconsistency.`, res)));
  }

  getVSphereOSImages(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereVirtualMachine[]> {
    return super.getVSphereOSImages(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVSphereVirtualMachine(res) || console.error(`TypeGuard for response 'VSphereVirtualMachine' caught inconsistency.`, res)));
  }

  exportTKGConfigForVsphere(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to generate tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.exportTKGConfigForVsphere(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  applyTKGConfigForVsphere(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to apply changes to tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    return super.applyTKGConfigForVsphere(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
  }

  importTKGConfigForVsphere(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VsphereRegionalClusterParams> {
    return super.importTKGConfigForVsphere(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isVsphereRegionalClusterParams(res) || console.error(`TypeGuard for response 'VsphereRegionalClusterParams' caught inconsistency.`, res)));
  }

  createVSphereRegionalCluster(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.createVSphereRegionalCluster(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  getVPCs(
    requestHttpOptions?: HttpOptions
  ): Observable<models.Vpc[]> {
    return super.getVPCs(requestHttpOptions)
      .pipe(tap((res: any) => guards.isVpc(res) || console.error(`TypeGuard for response 'Vpc' caught inconsistency.`, res)));
  }

  getAWSNodeTypes(
    args: {
      az?: string,  // (optional) AWS availability zone, e.g. us-west-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string[]> {
    return super.getAWSNodeTypes(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  getAWSRegions(
    requestHttpOptions?: HttpOptions
  ): Observable<string[]> {
    return super.getAWSRegions(requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  getAWSOSImages(
    args: {
      region: string,  // AWS region, e.g. us-west-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSVirtualMachine[]> {
    return super.getAWSOSImages(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAWSVirtualMachine(res) || console.error(`TypeGuard for response 'AWSVirtualMachine' caught inconsistency.`, res)));
  }

  getAWSCredentialProfiles(
    requestHttpOptions?: HttpOptions
  ): Observable<string[]> {
    return super.getAWSCredentialProfiles(requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  getAWSAvailabilityZones(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSAvailabilityZone[]> {
    return super.getAWSAvailabilityZones(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAWSAvailabilityZone(res) || console.error(`TypeGuard for response 'AWSAvailabilityZone' caught inconsistency.`, res)));
  }

  getAWSSubnets(
    args: {
      vpcId: string,  // VPC Id
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSSubnet[]> {
    return super.getAWSSubnets(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAWSSubnet(res) || console.error(`TypeGuard for response 'AWSSubnet' caught inconsistency.`, res)));
  }

  exportTKGConfigForAWS(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to generate TKG configuration file for AWS
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.exportTKGConfigForAWS(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  applyTKGConfigForAWS(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to apply changes to TKG configuration file for AWS
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    return super.applyTKGConfigForAWS(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
  }

  createAWSRegionalCluster(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.createAWSRegionalCluster(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  importTKGConfigForAWS(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for aws
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSRegionalClusterParams> {
    return super.importTKGConfigForAWS(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAWSRegionalClusterParams(res) || console.error(`TypeGuard for response 'AWSRegionalClusterParams' caught inconsistency.`, res)));
  }

  getAzureEndpoint(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureAccountParams> {
    return super.getAzureEndpoint(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureAccountParams(res) || console.error(`TypeGuard for response 'AzureAccountParams' caught inconsistency.`, res)));
  }

  getAzureResourceGroups(
    args: {
      location: string,  // Azure region
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureResourceGroup[]> {
    return super.getAzureResourceGroups(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureResourceGroup(res) || console.error(`TypeGuard for response 'AzureResourceGroup' caught inconsistency.`, res)));
  }

  createAzureResourceGroup(
    args: {
      params: models.AzureResourceGroup,  // parameters to create a new Azure resource group
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.createAzureResourceGroup(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  getAzureVnets(
    args: {
      resourceGroupName: string,  // Name of the Azure resource group
      location: string,  // Azure region
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureVirtualNetwork[]> {
    return super.getAzureVnets(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureVirtualNetwork(res) || console.error(`TypeGuard for response 'AzureVirtualNetwork' caught inconsistency.`, res)));
  }

  createAzureVirtualNetwork(
    args: {
      resourceGroupName: string,  // Name of the Azure resource group
      params: models.AzureVirtualNetwork,  // parameters to create a new Azure Virtual network
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.createAzureVirtualNetwork(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  getAzureOSImages(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureVirtualMachine[]> {
    return super.getAzureOSImages(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureVirtualMachine(res) || console.error(`TypeGuard for response 'AzureVirtualMachine' caught inconsistency.`, res)));
  }

  getAzureRegions(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureLocation[]> {
    return super.getAzureRegions(requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureLocation(res) || console.error(`TypeGuard for response 'AzureLocation' caught inconsistency.`, res)));
  }

  getAzureInstanceTypes(
    args: {
      location: string,  // Azure region name
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureInstanceType[]> {
    return super.getAzureInstanceTypes(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureInstanceType(res) || console.error(`TypeGuard for response 'AzureInstanceType' caught inconsistency.`, res)));
  }

  exportTKGConfigForAzure(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to generate TKG configuration file for Azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.exportTKGConfigForAzure(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  applyTKGConfigForAzure(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to apply changes to TKG configuration file for Azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    return super.applyTKGConfigForAzure(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
  }

  createAzureRegionalCluster(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.createAzureRegionalCluster(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  importTKGConfigForAzure(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureRegionalClusterParams> {
    return super.importTKGConfigForAzure(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isAzureRegionalClusterParams(res) || console.error(`TypeGuard for response 'AzureRegionalClusterParams' caught inconsistency.`, res)));
  }

  checkIfDockerDaemonAvailable(
    requestHttpOptions?: HttpOptions
  ): Observable<models.DockerDaemonStatus> {
    return super.checkIfDockerDaemonAvailable(requestHttpOptions)
      .pipe(tap((res: any) => guards.isDockerDaemonStatus(res) || console.error(`TypeGuard for response 'DockerDaemonStatus' caught inconsistency.`, res)));
  }

  exportTKGConfigForDocker(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to generate TKG configuration file for Docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.exportTKGConfigForDocker(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  applyTKGConfigForDocker(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to apply changes to TKG configuration file for Docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo> {
    return super.applyTKGConfigForDocker(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isConfigFileInfo(res) || console.error(`TypeGuard for response 'ConfigFileInfo' caught inconsistency.`, res)));
  }

  createDockerRegionalCluster(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string> {
    return super.createDockerRegionalCluster(args, requestHttpOptions)
      .pipe(tap((res: any) => typeof res === 'string' || console.error(`TypeGuard for response 'string' caught inconsistency.`, res)));
  }

  importTKGConfigForDocker(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.DockerRegionalClusterParams> {
    return super.importTKGConfigForDocker(args, requestHttpOptions)
      .pipe(tap((res: any) => guards.isDockerRegionalClusterParams(res) || console.error(`TypeGuard for response 'DockerRegionalClusterParams' caught inconsistency.`, res)));
  }

}
