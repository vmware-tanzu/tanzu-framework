
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
