import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export enum LoadBalancerField {
    CLOUD_NAME = 'cloudName',
    CLUSTER_LABELS = 'clusterLabels',
    CONTROLLER_CERT = 'controllerCert',
    CONTROLLER_HOST = 'controllerHost',
    MANAGEMENT_CLUSTER_NETWORK_CIDR = 'managementClusterNetworkCIDR',
    MANAGEMENT_CLUSTER_NETWORK_NAME = 'managementClusterNetworkName',
    NETWORK_CIDR = 'networkCIDR',
    NETWORK_NAME = 'networkName',
    NEW_LABEL_KEY = 'newLabelKey',
    NEW_LABEL_VALUE = 'newLabelValue',
    PASSWORD = 'password',
    SERVICE_ENGINE_GROUP_NAME = 'serviceEngineGroupName',
    USERNAME = 'username',
}

export const LoadBalancerStepMapping: StepMapping = {
    fieldMappings: [
        { name: LoadBalancerField.CONTROLLER_HOST, validators: [SimpleValidator.IS_VALID_FQDN_OR_IP] },
        { name: LoadBalancerField.USERNAME },
        { name: LoadBalancerField.PASSWORD },
        { name: LoadBalancerField.CLOUD_NAME },
        { name: LoadBalancerField.SERVICE_ENGINE_GROUP_NAME },
        { name: LoadBalancerField.NETWORK_NAME },
        { name: LoadBalancerField.NETWORK_CIDR },
        { name: LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_NAME },
        { name: LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_CIDR, validators: [SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
        { name: LoadBalancerField.CONTROLLER_CERT },
        { name: LoadBalancerField.CLUSTER_LABELS },
        { name: LoadBalancerField.NETWORK_NAME, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
        { name: LoadBalancerField.NEW_LABEL_VALUE, validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
    ]
}
