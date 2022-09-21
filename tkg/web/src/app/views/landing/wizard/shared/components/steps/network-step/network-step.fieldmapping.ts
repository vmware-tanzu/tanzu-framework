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
    FULL_PROXY_LIST = 'fullProxyList'
}

const ProviderNetworkFieldMapping: FieldMapping[] = [
    { name: NetworkField.CNI_TYPE, defaultValue: 'antrea', label: 'CNI PROVIDER' },
]
const BasicNetworkFieldMappings: FieldMapping[] = [
    { name: NetworkField.PROXY_SETTINGS, isBoolean: true, label: 'ACTIVATE PROXY SETTINGS' },
    { name: NetworkField.HTTP_PROXY_URL, label: 'HTTP PROXY URL' },
    { name: NetworkField.HTTP_PROXY_USERNAME, label: 'HTTP PROXY USERNAME (OPTIONAL)' },
    { name: NetworkField.HTTP_PROXY_PASSWORD, mask: true, label: 'HTTP PROXY PASSWORD (OPTIONAL)' },
    { name: NetworkField.HTTPS_PROXY_URL, label: 'HTTPS PROXY URL' },
    { name: NetworkField.HTTPS_IS_SAME_AS_HTTP, isBoolean: true, defaultValue: true, label: 'USE SAME CONFIGURATION FOR HTTPS PROXY' },
    { name: NetworkField.HTTPS_PROXY_USERNAME, label: 'HTTPS PROXY USERNAME (OPTIONAL)' },
    { name: NetworkField.HTTPS_PROXY_PASSWORD, mask: true, label: 'HTTPS PROXY PASSWORD (OPTIONAL)' },
    { name: NetworkField.NO_PROXY, label: 'NO PROXY (OPTIONAL)' },
    { name: NetworkField.FULL_PROXY_LIST, label: 'FULL NO PROXY LIST', displayOnly: true},
];

export const NetworkIpv4StepMapping: StepMapping = {
    // Because, by default, the ORDER of these fields is the order in which they are displayed,
    // we put the fields in the expected display order (and avoid having to order them elsewhere in the code)
    fieldMappings: [
        ...ProviderNetworkFieldMapping,
        { name: NetworkField.CLUSTER_SERVICE_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR, label: 'CLUSTER SERVICE CIDR' },
        { name: NetworkField.CLUSTER_POD_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR, label: 'CLUSTER POD CIDR' },
        ...BasicNetworkFieldMappings
    ]
}

export const NetworkIpv6StepMapping: StepMapping = {
    fieldMappings: [
        ...ProviderNetworkFieldMapping,
        { name: NetworkField.CLUSTER_SERVICE_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR, label: 'CLUSTER SERVICE CIDR' },
        { name: NetworkField.CLUSTER_POD_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR, label: 'CLUSTER POD CIDR' },
        ...BasicNetworkFieldMappings
    ]
}
