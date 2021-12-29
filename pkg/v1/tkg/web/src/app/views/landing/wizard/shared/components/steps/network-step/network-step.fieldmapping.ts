import { FieldMapping, StepMapping } from '../../../field-mapping/FieldMapping';
import { IAAS_DEFAULT_CIDRS } from '../../../../../../../shared/constants/app.constants';

export enum NetworkField {
    CLUSTER_SERVICE_CIDR = 'clusterServiceCidr',
    CLUSTER_POD_CIDR = 'clusterPodCidr',
    CNI_TYPE = 'cniType',
    HTTP_PROXY_URL = 'httpProxyUrl',
    HTTP_PROXY_USERNAME = 'httpProxyUsername',
    HTTP_PROXY_PASSWORD = 'httpProxyPassword',
    HTTPS_IS_SAME_AS_HTTP = 'isSameAsHttp',
    HTTPS_PROXY_URL = 'httpsProxyUrl',
    HTTPS_PROXY_USERNAME = 'httpsProxyUsername',
    HTTPS_PROXY_PASSWORD = 'httpsProxyPassword',
    NETWORK_NAME = 'networkName',
    NO_PROXY = 'noProxy',
    PROXY_SETTINGS = 'proxySettings',
}

const BasicNetworkFieldMappings: FieldMapping[] =
    [
        { name: NetworkField.CNI_TYPE, defaultValue: 'antrea', label: 'CNI PROVIDER' },
        { name: NetworkField.HTTP_PROXY_URL, label: 'HTTP PROXY URL' },
        { name: NetworkField.HTTP_PROXY_USERNAME, label: 'HTTP PROXY USERNAME (OPTIONAL)' },
        { name: NetworkField.HTTP_PROXY_PASSWORD, mask: true, label: 'HTTP PROXY PASSWORD (OPTIONAL)' },
        { name: NetworkField.HTTPS_PROXY_URL, label: 'HTTPS PROXY URL' },
        { name: NetworkField.HTTPS_PROXY_USERNAME, label: 'HTTPS PROXY USERNAME (OPTIONAL)' },
        { name: NetworkField.HTTPS_PROXY_PASSWORD, mask: true, label: 'HTTPS PROXY PASSWORD (OPTIONAL)' },
        { name: NetworkField.NO_PROXY, label: 'NO PROXY (OPTIONAL)' },
        { name: NetworkField.NETWORK_NAME, label: 'NETWORK NAME' },
        { name: NetworkField.PROXY_SETTINGS, isBoolean: true, label: 'ENABLE PROXY SETTINGS' },
        { name: NetworkField.HTTPS_IS_SAME_AS_HTTP, isBoolean: true, defaultValue: true },
    ];

export const NetworkIpv4StepMapping: StepMapping = {
    fieldMappings: [
        ...BasicNetworkFieldMappings,
        { name: NetworkField.CLUSTER_SERVICE_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR, label: 'CLUSTER SERVICE CIDR' },
        { name: NetworkField.CLUSTER_POD_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR, label: 'CLUSTER POD CIDR' }
    ]
}

export const NetworkIpv6StepMapping: StepMapping = {
    fieldMappings: [
        ...BasicNetworkFieldMappings,
        { name: NetworkField.CLUSTER_SERVICE_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR, label: 'CLUSTER SERVICE CIDR' },
        { name: NetworkField.CLUSTER_POD_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR, label: 'CLUSTER POD CIDR' }
    ]
}
