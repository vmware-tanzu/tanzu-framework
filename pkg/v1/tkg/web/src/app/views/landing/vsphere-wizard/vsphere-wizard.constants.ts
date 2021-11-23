import {NodeType} from "../wizard/shared/constants/wizard.constants";

export const vSphereNodeTypes: Array<NodeType> = [
    {
        id: 'small',
        name: 'small (cpu: 2, ram: 4 GB, disk: 20 GB)'
    },
    {
        id: 'medium',
        name: 'medium (cpu: 2, ram: 8 GB, disk: 40 GB)'
    },
    {
        id: 'large',
        name: 'large (cpu: 4, ram: 16 GB, disk: 40 GB)'
    },
    {
        id: 'extra-large',
        name: 'extra-large (cpu: 8, ram: 32 GB, disk: 80 GB)'
    }
];

export enum VsphereField {
    NODESETTING_CONTROL_PLANE_SETTING = 'controlPlaneSetting',
    NODESETTING_INSTANCE_TYPE_DEV = 'devInstanceType',
    NODESETTING_INSTANCE_TYPE_PROD = 'prodInstanceType',
    NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED = 'machineHealthChecksEnabled',
    NODESETTING_WORKER_NODE_INSTANCE_TYPE = 'workerNodeInstanceType',
    NODESETTING_CLUSTER_NAME = 'clusterName',
    NODESETTING_CONTROL_PLANE_ENDPOINT_IP = 'controlPlaneEndpointIP',
    NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER = 'controlPlaneEndpointProvider',
}
