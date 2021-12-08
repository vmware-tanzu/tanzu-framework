
export interface NodeType {
    id: string;
    name: string;
}

export const managementClusterPlugin = 'management-cluster';

// ClusterType enum are data values sent to the backend to specify the cluster type
export enum ClusterType {
    Management = 'management',
    Standalone = 'standalone',
}

export enum IdentityManagementType {
    LDAP = 'ldap',
    OIDC = 'oidc',
}

export enum WizardStep {
    IDENTITY = 'identity',
    METADATA= 'metadata',
    NETWORK = 'network',
    OSIMAGE = 'osimage',
}

export enum WizardForm {
    IDENTITY = 'identityForm',
    METADATA= 'metadataForm',
    NETWORK = 'networkForm',
    OSIMAGE = 'osImageForm',
}

export enum IdentityField {
    TYPE = 'identityType',
    ISSUER_URL = 'issuerURL',
    LDAP_ENDPOINT_IP = 'endpointIp',
    LDAP_ENDPOINT_PORT = 'endpointPort',
}
