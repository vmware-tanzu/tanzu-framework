import { IdentityManagementConfig, TKGNetwork } from '.';
export interface DockerRegionalClusterParams {
    annotations?: {
        [key: string]: string;
    };
    ceipOptIn?: boolean;
    clusterName?: string;
    controlPlaneFlavor?: string;
    identityManagement?: IdentityManagementConfig;
    kubernetesVersion?: string;
    labels?: {
        [key: string]: string;
    };
    machineHealthCheckEnabled?: boolean;
    networking?: TKGNetwork;
    numOfWorkerNodes?: string;
}
