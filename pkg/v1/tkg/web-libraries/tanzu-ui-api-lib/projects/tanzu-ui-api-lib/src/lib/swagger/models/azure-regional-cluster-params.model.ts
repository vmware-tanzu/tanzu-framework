/* tslint:disable */
import {
  AzureAccountParams,
  AzureVirtualMachine,
  IdentityManagementConfig,
  TKGNetwork,
} from '.';

export interface AzureRegionalClusterParams {
  annotations?: { [key: string]: string };
  azureAccountParams?: AzureAccountParams;
  ceipOptIn?: boolean;
  clusterName?: string;
  controlPlaneFlavor?: string;
  controlPlaneMachineType?: string;
  controlPlaneSubnet?: string;
  controlPlaneSubnetCidr?: string;
  enableAuditLogging?: boolean;
  frontendPrivateIp?: string;
  identityManagement?: IdentityManagementConfig;
  isPrivateCluster?: boolean;
  kubernetesVersion?: string;
  labels?: { [key: string]: string };
  location?: string;
  machineHealthCheckEnabled?: boolean;
  networking?: TKGNetwork;
  numOfWorkerNodes?: string;
  os?: AzureVirtualMachine;
  resourceGroup?: string;
  sshPublicKey?: string;
  vnetCidr?: string;
  vnetName?: string;
  vnetResourceGroup?: string;
  workerMachineType?: string;
  workerNodeSubnet?: string;
  workerNodeSubnetCidr?: string;
}
