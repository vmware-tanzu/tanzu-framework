import { OSInfo } from '.';
export interface VSphereVirtualMachine {
    isTemplate: boolean;
    k8sVersion?: string;
    moid?: string;
    name?: string;
    osInfo?: OSInfo;
}
