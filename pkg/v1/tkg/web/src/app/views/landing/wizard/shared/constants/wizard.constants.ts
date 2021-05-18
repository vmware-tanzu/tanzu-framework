
export interface NodeType {
    id: string;
    name: string;
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

export const awsNodeTypes: Array<NodeType> = [
    {
        id: 't3.small',
        name: 't3.small'
    },
    {
        id: 't3.medium',
        name: 't3.medium'
    },
    {
        id: 't3.large',
        name: 't3.large'
    },
    {
        id: 't3.xlarge',
        name: 't3.xlarge'
    },
    {
        id: 'm5.large',
        name: 'm5.large'
    },
    {
        id: 'm5.xlarge',
        name: 'm5.xlarge'
    },
    {
        id: 'm5a.2xlarge',
        name: 'm5a.2xlarge'
    },
    {
        id: 'm5a.4xlarge',
        name: 'm5a.4xlarge'
    },
    {
        id: 'r4.8xlarge',
        name: 'r4.8xlarge'
    },
    {
        id: 'i3.xlarge',
        name: 'i3.xlarge'
    }
];
