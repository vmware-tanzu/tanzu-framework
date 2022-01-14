/* tslint:disable */

import { Observable } from 'rxjs';
import { HttpOptions } from './';
import * as models from './models';

export interface APIClientInterface {

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getUI(
    requestHttpOptions?: HttpOptions
  ): Observable<File>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getUIFile(
    args: {
      filename: string,  // UI file name
    },
    requestHttpOptions?: HttpOptions
  ): Observable<File>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getFeatureFlags(
    requestHttpOptions?: HttpOptions
  ): Observable<models.Features>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getTanzuEdition(
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  verifyAccount(
    args: {
      credentials?: models.AviControllerParams,  // (optional) Avi controller credentials
    },
    requestHttpOptions?: HttpOptions
  ): Observable<void>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapConnect(
    args: {
      credentials?: models.LdapParams,  // (optional) LDAP configuration
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapBind(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapUserSearch(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  verifyLdapGroupSearch(
    requestHttpOptions?: HttpOptions
  ): Observable<models.LdapTestResult>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  verifyLdapCloseConnection(
    requestHttpOptions?: HttpOptions
  ): Observable<void>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAviClouds(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviCloud[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAviServiceEngineGroups(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviServiceEngineGroup[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAviVipNetworks(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AviVipNetwork[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getProvider(
    requestHttpOptions?: HttpOptions
  ): Observable<models.ProviderInfo>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVsphereThumbprint(
    args: {
      host: string,  // vSphere host
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereThumbprint>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  setVSphereEndpoint(
    args: {
      credentials?: models.VSphereCredentials,  // (optional) vSphere credentials
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VsphereInfo>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereDatacenters(
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereDatacenter[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereDatastores(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereDatastore[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereFolders(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereFolder[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereComputeResources(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereManagementObject[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereResourcePools(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereResourcePool[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereNetworks(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereNetwork[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereNodeTypes(
    requestHttpOptions?: HttpOptions
  ): Observable<models.NodeType[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVSphereOSImages(
    args: {
      dc: string,  // datacenter managed object Id, e.g. datacenter-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VSphereVirtualMachine[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForVsphere(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to generate tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForVsphere(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to apply changes to tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForVsphere(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for vsphere
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.VsphereRegionalClusterParams>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createVSphereRegionalCluster(
    args: {
      params: models.VsphereRegionalClusterParams,  // params to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  setAWSEndpoint(
    args: {
      accountParams?: models.AWSAccountParams,  // (optional) AWS account parameters
    },
    requestHttpOptions?: HttpOptions
  ): Observable<void>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getVPCs(
    requestHttpOptions?: HttpOptions
  ): Observable<models.Vpc[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSNodeTypes(
    args: {
      az?: string,  // (optional) AWS availability zone, e.g. us-west-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSRegions(
    requestHttpOptions?: HttpOptions
  ): Observable<string[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSOSImages(
    args: {
      region: string,  // AWS region, e.g. us-west-2
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSVirtualMachine[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSCredentialProfiles(
    requestHttpOptions?: HttpOptions
  ): Observable<string[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSAvailabilityZones(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSAvailabilityZone[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAWSSubnets(
    args: {
      vpcId: string,  // VPC Id
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSSubnet[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForAWS(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to generate TKG configuration file for AWS
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForAWS(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to apply changes to TKG configuration file for AWS
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createAWSRegionalCluster(
    args: {
      params: models.AWSRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForAWS(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for aws
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AWSRegionalClusterParams>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureEndpoint(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureAccountParams>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  setAzureEndpoint(
    args: {
      accountParams?: models.AzureAccountParams,  // (optional) Azure account parameters
    },
    requestHttpOptions?: HttpOptions
  ): Observable<void>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureResourceGroups(
    args: {
      location: string,  // Azure region
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureResourceGroup[]>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  createAzureResourceGroup(
    args: {
      params: models.AzureResourceGroup,  // parameters to create a new Azure resource group
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureVnets(
    args: {
      resourceGroupName: string,  // Name of the Azure resource group
      location: string,  // Azure region
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureVirtualNetwork[]>;

  /**
   * Response generated for [ 201 ] HTTP response code.
   */
  createAzureVirtualNetwork(
    args: {
      resourceGroupName: string,  // Name of the Azure resource group
      params: models.AzureVirtualNetwork,  // parameters to create a new Azure Virtual network
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureOSImages(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureVirtualMachine[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureRegions(
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureLocation[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  getAzureInstanceTypes(
    args: {
      location: string,  // Azure region name
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureInstanceType[]>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForAzure(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to generate TKG configuration file for Azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForAzure(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to apply changes to TKG configuration file for Azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createAzureRegionalCluster(
    args: {
      params: models.AzureRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForAzure(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for azure
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.AzureRegionalClusterParams>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  checkIfDockerDaemonAvailable(
    requestHttpOptions?: HttpOptions
  ): Observable<models.DockerDaemonStatus>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  exportTKGConfigForDocker(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to generate TKG configuration file for Docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  applyTKGConfigForDocker(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to apply changes to TKG configuration file for Docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.ConfigFileInfo>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  createDockerRegionalCluster(
    args: {
      params: models.DockerRegionalClusterParams,  // parameters to create a regional cluster
    },
    requestHttpOptions?: HttpOptions
  ): Observable<string>;

  /**
   * Response generated for [ 200 ] HTTP response code.
   */
  importTKGConfigForDocker(
    args: {
      params: models.ConfigFile,  // config file from which to generate tkg configuration for docker
    },
    requestHttpOptions?: HttpOptions
  ): Observable<models.DockerRegionalClusterParams>;

}
