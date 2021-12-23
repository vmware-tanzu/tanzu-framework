import { StepMapping } from '../../../field-mapping/FieldMapping';
import { SimpleValidator } from '../../../constants/validation.constants';

export const LoadBalancerStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'controllerHost', validators: [SimpleValidator.IS_VALID_FQDN_OR_IP] },
        { name: 'username' },
        { name: 'password' },
        { name: 'cloudName' },
        { name: 'serviceEngineGroupName' },
        { name: 'networkName' },
        { name: 'networkCIDR' },
        { name: 'managementClusterNetworkName' },
        { name: 'managementClusterNetworkCIDR', validators: [SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
        { name: 'controllerCert' },
        { name: 'clusterLabels' },
        { name: 'newLabelKey', validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
        { name: 'newLabelValue', validators: [SimpleValidator.IS_VALID_LABEL_OR_ANNOTATION] },
    ]
}
