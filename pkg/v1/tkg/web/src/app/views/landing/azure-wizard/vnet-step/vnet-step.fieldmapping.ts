import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureField } from '../azure-wizard.constants';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const AzureVnetStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, required: true },
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME, required: true },
        { name: AzureField.VNET_CONTROLPLANE_SUBNET_NAME, required: true },
        { name: AzureField.VNET_CUSTOM_NAME, required: true },
        { name: AzureField.VNET_EXISTING_NAME, required: true },
        { name: AzureField.VNET_EXISTING_OR_CUSTOM, required: true },
        { name: AzureField.VNET_PRIVATE_CLUSTER },
        { name: AzureField.VNET_PRIVATE_IP },
        { name: AzureField.VNET_RESOURCE_GROUP, required: true },
        // special hidden field used to capture existing subnet cidr when user selects existing subnet
        { name:  AzureField.VNET_CONTROLPLANE_SUBNET_CIDR },
        // defaultCidrFields
        { name: AzureField.VNET_CUSTOM_CIDR, required: true,
            validators: [SimpleValidator.NO_WHITE_SPACE, SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, required: true,
            validators: [SimpleValidator.NO_WHITE_SPACE, SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
    ]
};
export const AzureVnetStepMapping: StepMapping = {
    fieldMappings: [
        ...AzureVnetStandaloneStepMapping.fieldMappings,
        { name: AzureField.VNET_WORKER_NEWSUBNET_NAME, required: true },
        { name: AzureField.VNET_WORKER_NEWSUBNET_CIDR, required: true },
        { name: AzureField.VNET_WORKER_SUBNET_NAME, required: true },
        // defaultCidrFields
        { name: AzureField.VNET_WORKER_NEWSUBNET_CIDR, required: true,
            validators: [SimpleValidator.NO_WHITE_SPACE, SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
    ]
}
