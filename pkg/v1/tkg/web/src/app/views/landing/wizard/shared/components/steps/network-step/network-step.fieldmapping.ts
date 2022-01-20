import { FieldMapping, StepMapping } from '../../../field-mapping/FieldMapping';
import { IAAS_DEFAULT_CIDRS } from '../../../../../../../shared/constants/app.constants';

const BasicNetworkFieldMappings: FieldMapping[] =
    [
        { name: 'cniType', defaultValue: 'antrea' },
        { name: 'httpProxyUrl', initWithSavedValue: true },
        { name: 'httpProxyUsername', initWithSavedValue: true },
        { name: 'httpProxyPassword', initWithSavedValue: true },
        { name: 'httpsProxyUrl', initWithSavedValue: true },
        { name: 'httpsProxyUsername', initWithSavedValue: true },
        { name: 'httpsProxyPassword', initWithSavedValue: true },
        { name: 'noProxy', initWithSavedValue: true },
        { name: 'networkName', initWithSavedValue: true },
        { name: 'proxySettings', isBoolean: true },
        { name: 'isSameAsHttp', isBoolean: true, defaultValue: true },
    ];

export const NetworkIpv4StepMapping: StepMapping = {
    fieldMappings: [
        ...BasicNetworkFieldMappings,
        { name: 'clusterServiceCidr', defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_CIDR },
        { name: 'clusterPodCidr', defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_CIDR }
    ]
}

export const NetworkIpv6StepMapping: StepMapping = {
    fieldMappings: [
        ...BasicNetworkFieldMappings,
        { name: 'clusterServiceCidr', defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_SVC_IPV6_CIDR },
        { name: 'clusterPodCidr', defaultValue: IAAS_DEFAULT_CIDRS.CLUSTER_POD_IPV6_CIDR }
    ]
}
