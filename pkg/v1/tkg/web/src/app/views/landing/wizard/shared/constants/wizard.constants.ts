
export interface NodeType {
    id: string;
    name: string;
}

export const managementClusterPlugin = 'management-cluster';

// ClusterType enum are data values sent to the backend to specify the cluster type
// To reference the string on the right, use '' + ClusterType.Management
export enum ClusterType {
    Management = 'management',
    Standalone = 'standalone',
}

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
