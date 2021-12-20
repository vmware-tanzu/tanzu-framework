import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureField, ResourceGroupOption } from '../azure-wizard.constants';

export const AzureProviderFieldMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, required: true },
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME, required: true },
        { name: AzureField.VNET_CONTROLPLANE_SUBNET_NAME, required: true },
        { name: AzureField.VNET_CUSTOM_NAME, required: true },
        { name: AzureField.VNET_EXISTING_NAME, required: true },
        { name: AzureField.VNET_EXISTING_OR_CUSTOM, required: true },
        { name: AzureField.VNET_RESOURCE_GROUP, required: true },
    ]
}

