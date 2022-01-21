import { AviSubnet } from '.';
export interface AviVipNetwork {
    cloud?: string;
    configedSubnets?: AviSubnet[];
    name?: string;
    uuid?: string;
}
