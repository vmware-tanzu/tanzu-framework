import { AviConfig, IdentityManagementConfig, TKGNetwork, VSphereCredentials, VSphereVirtualMachine } from '.';
export interface VsphereRegionalClusterParams {
    annotations?: {
        [key: string]: string;
    };
    aviConfig?: AviConfig;
    ceipOptIn?: boolean;
    clusterName?: string;
    controlPlaneEndpoint?: string;
    controlPlaneFlavor?: string;
    controlPlaneNodeType?: string;
    datacenter?: string;
    datastore?: string;
    enableAuditLogging?: boolean;
    folder?: string;
    identityManagement?: IdentityManagementConfig;
    ipFamily?: string;
    kubernetesVersion?: string;
    labels?: {
        [key: string]: string;
    };
    machineHealthCheckEnabled?: boolean;
    networking?: TKGNetwork;
    numOfWorkerNode?: number;
    os?: VSphereVirtualMachine;
    resourcePool?: string;
    ssh_key?: string;
    vsphereCredentials?: VSphereCredentials;
    workerNodeType?: string;
}
