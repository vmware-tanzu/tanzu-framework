/* tslint:disable */
import {
  AWSAccountParams,
  AWSVirtualMachine,
  AWSVpc,
  IdentityManagementConfig,
  TKGNetwork,
} from '.';

export interface AWSRegionalClusterParams {
  annotations?: { [key: string]: string };
  awsAccountParams?: AWSAccountParams;
  bastionHostEnabled?: boolean;
  ceipOptIn?: boolean;
  clusterName?: string;
  controlPlaneFlavor?: string;
  controlPlaneNodeType?: string;
  createCloudFormationStack?: boolean;
  enableAuditLogging?: boolean;
  identityManagement?: IdentityManagementConfig;
  kubernetesVersion?: string;
  labels?: { [key: string]: string };
  loadbalancerSchemeInternal?: boolean;
  machineHealthCheckEnabled?: boolean;
  networking?: TKGNetwork;
  numOfWorkerNode?: number;
  os?: AWSVirtualMachine;
  sshKeyName?: string;
  vpc?: AWSVpc;
  workerNodeType?: string;
}
