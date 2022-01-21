import { AzureSubnet } from '.';
export interface AzureVirtualNetwork {
    cidrBlock: string;
    id?: string;
    location: string;
    name: string;
    subnets?: AzureSubnet[];
}
