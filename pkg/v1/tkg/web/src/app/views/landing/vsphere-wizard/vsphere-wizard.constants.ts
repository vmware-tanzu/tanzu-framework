import {NodeType} from "../wizard/shared/constants/wizard.constants";

export const VsphereNodeTypes: Array<NodeType> = [
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
    NETWORK_NAME = 'networkName',

    NODESETTING_CLUSTER_NAME = 'clusterName',
    NODESETTING_CONTROL_PLANE_ENDPOINT_IP = 'controlPlaneEndpointIP',
    NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER = 'controlPlaneEndpointProvider',
    NODESETTING_CONTROL_PLANE_SETTING = 'controlPlaneSetting',
    NODESETTING_ENABLE_AUDIT_LOGGING = 'enableAuditLogging',
    NODESETTING_INSTANCE_TYPE_DEV = 'devInstanceType',
    NODESETTING_INSTANCE_TYPE_PROD = 'prodInstanceType',
    NODESETTING_MACHINE_HEALTH_CHECKS_ENABLED = 'machineHealthChecksEnabled',
    NODESETTING_WORKER_NODE_INSTANCE_TYPE = 'workerNodeInstanceType',

    PROVIDER_CONNECTION_INSECURE = 'insecure',
    PROVIDER_DATA_CENTER = 'datacenter',
    PROVIDER_IP_FAMILY = 'ipFamily',
    PROVIDER_SSH_KEY = 'ssh_key',
    PROVIDER_SSH_KEY_FILE = 'ssh_key_file',
    PROVIDER_THUMBPRINT = 'thumbprint',
    PROVIDER_USER_NAME = 'username',
    PROVIDER_USER_PASSWORD = 'password',
    PROVIDER_VCENTER_ADDRESS = 'vcenterAddress',

    RESOURCE_DATASTORE = 'datastore',
    RESOURCE_POOL = 'resourcePool',
    RESOURCE_VMFOLDER = 'vmFolder',
}

export enum VsphereForm {
    NODESETTING = 'vsphereNodeSettingForm',
    PROVIDER = 'vsphereProviderForm',
    RESOURCE = 'resourceForm',
}
