import {ControlType, StepMapping} from '../../../field-mapping/FieldMapping';
import {SimpleValidator} from '../../../constants/validation.constants';

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
        {
            name: LoadBalancerField.CONTROLLER_HOST,
            validators: [SimpleValidator.IS_VALID_FQDN_OR_IP],
            label: 'CONTROLLER HOST',
        },
        {name: LoadBalancerField.USERNAME, label: 'USERNAME'},
        {name: LoadBalancerField.PASSWORD, mask: true, label: 'PASSWORD'},
        {name: LoadBalancerField.CLOUD_NAME, label: 'CLOUD NAME'},
        {name: LoadBalancerField.SERVICE_ENGINE_GROUP_NAME, label: 'SERVICE ENGINE GROUP NAME'},
        {name: LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_NAME, label: 'MANAGEMENT VIP NETWORK NAME'},
        {
            name: LoadBalancerField.MANAGEMENT_CLUSTER_NETWORK_CIDR,
            validators: [SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT],
            label: 'MANAGEMENT VIP NETWORK CIDR',
        },
        {name: LoadBalancerField.NETWORK_NAME, label: 'WORKLOAD VIP NETWORK NAME'},
        {name: LoadBalancerField.NETWORK_CIDR, label: 'WORKLOAD VIP NETWORK CIDR'},
        {name: LoadBalancerField.CONTROLLER_CERT, doNotAutoSave: true, label: 'CONTROLLER CERTIFICATE AUTHORITY'},
        {
            name: LoadBalancerField.CLUSTER_LABELS,
            label: 'CLUSTER LABELS (OPTIONAL)',
            controlType: ControlType.FormArray,
            displayFunction: labels => labels.map(label => `${label.key} : ${label.value}`).join(', '),
            children: [
                {
                    name: 'key',
                    defaultValue: '',
                    controlType: ControlType.FormControl,
                    validators: [
                        SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION,
                        SimpleValidator.RX_UNIQUE,
                        SimpleValidator.RX_REQUIRED_IF_VALUE
                    ]
                },
                {
                    name: 'value',
                    defaultValue: '',
                    controlType: ControlType.FormControl,
                    validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION, SimpleValidator.RX_REQUIRED_IF_KEY]
                }
            ],
        },
    ],
};
