export interface Providers {
    [key: string]: string;
}

export const PROVIDERS: Providers = {
    VSPHERE: 'vsphere',
    AWS: 'aws',
    AZURE: 'azure',
    DOCKER: 'docker'
};

export const IAAS_DEFAULT_CIDRS = {
    CLUSTER_SVC_CIDR: '100.64.0.0/13',
    CLUSTER_POD_CIDR: '100.96.0.0/11',
    CLUSTER_SVC_IPV6_CIDR: 'fd00:100:96::/108',
    CLUSTER_POD_IPV6_CIDR: 'fd00:100:64::/48'
};

export enum IpFamilyEnum {
    IPv4 = 'ipv4',
    IPv6 = 'ipv6'
};
