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
        { name: NetworkField.CNI_TYPE, defaultValue: 'antrea' },
        { name: NetworkField.HTTP_PROXY_URL, initWithSavedValue: true },
        { name: NetworkField.HTTP_PROXY_USERNAME, initWithSavedValue: true },
        { name: NetworkField.HTTP_PROXY_PASSWORD, initWithSavedValue: true },
        { name: NetworkField.HTTPS_PROXY_URL, initWithSavedValue: true },
        { name: NetworkField.HTTPS_PROXY_USERNAME, initWithSavedValue: true },
        { name: NetworkField.HTTPS_PROXY_PASSWORD, initWithSavedValue: true },
        { name: NetworkField.NO_PROXY, initWithSavedValue: true },
        { name: NetworkField.NETWORK_NAME, initWithSavedValue: true },
        { name: NetworkField.PROXY_SETTINGS, isBoolean: true },
        { name: NetworkField.HTTPS_IS_SAME_AS_HTTP, isBoolean: true, defaultValue: true },
    ];

export const NetworkIpv4StepMapping: StepMapping = {
    fieldMappings: [
        ...BasicNetworkFieldMappings,
        { name: NetworkField.CLUSTER_SERVICE_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR },
        { name: NetworkField.CLUSTER_POD_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR }
    ]
}

export const NetworkIpv6StepMapping: StepMapping = {
    fieldMappings: [
        ...BasicNetworkFieldMappings,
        { name: NetworkField.CLUSTER_SERVICE_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR },
        { name: NetworkField.CLUSTER_POD_CIDR, defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR }
    ]
}
