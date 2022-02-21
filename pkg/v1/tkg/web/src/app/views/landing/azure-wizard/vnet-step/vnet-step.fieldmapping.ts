import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureField, VnetOptionType } from '../azure-wizard.constants';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const AzureVnetStandaloneStepMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, required: true, doNotAutoRestore: true, label: 'CONTROL PLANE SUBNET CIDR' },
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_NAME, required: true, doNotAutoRestore: true, label: 'CONTROL PLANE SUBNET NAME' },
        { name: AzureField.VNET_CONTROLPLANE_SUBNET_NAME, required: true, requiresBackendData: true, label: 'CONTROL PLANE SUBNET' },
        { name: AzureField.VNET_CUSTOM_NAME, required: true, doNotAutoRestore: true, label: 'VNET NAME' },
        { name: AzureField.VNET_EXISTING_NAME, required: true, requiresBackendData: true, label: 'VNET NAME' },
        { name: AzureField.VNET_EXISTING_OR_CUSTOM, required: true, primaryTrigger: true, defaultValue: VnetOptionType.EXISTING },
        { name: AzureField.VNET_PRIVATE_CLUSTER, isBoolean: true, label: 'PRIVATE AZURE CLUSTER' },
        { name: AzureField.VNET_PRIVATE_IP, label: 'PRIVATE IP' },
        { name: AzureField.VNET_RESOURCE_GROUP, required: true, requiresBackendData: true, label: 'RESOURCE GROUP' },
        // special hidden field used to capture existing subnet cidr when user selects existing subnet
        { name:  AzureField.VNET_CONTROLPLANE_SUBNET_CIDR, doNotAutoRestore: true },
        // defaultCidrFields
        { name: AzureField.VNET_CUSTOM_CIDR, required: true, doNotAutoRestore: true, label: 'VNET CIDR BLOCK',
            validators: [SimpleValidator.NO_WHITE_SPACE, SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
        { name: AzureField.VNET_CONTROLPLANE_NEWSUBNET_CIDR, required: true, doNotAutoRestore: true,
            validators: [SimpleValidator.NO_WHITE_SPACE, SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
    ]
};

export const AzureVnetStepMapping: StepMapping = {
    fieldMappings: [
        ...AzureVnetStandaloneStepMapping.fieldMappings,
        { name: AzureField.VNET_WORKER_NEWSUBNET_NAME, required: true, doNotAutoRestore: true, label: 'WORKER NODE SUBNET NAME' },
        { name: AzureField.VNET_WORKER_NEWSUBNET_CIDR, required: true, doNotAutoRestore: true },
        { name: AzureField.VNET_WORKER_SUBNET_NAME, required: true, requiresBackendData: true, label: 'WORKER NODE SUBNET' },
        // defaultCidrFields
        { name: AzureField.VNET_WORKER_NEWSUBNET_CIDR, required: true, doNotAutoRestore: true, label: 'WORKER NODE SUBNET CIDR',
            validators: [SimpleValidator.NO_WHITE_SPACE, SimpleValidator.IS_VALID_IP_NETWORK_SEGMENT] },
    ]
}

// About AzureVnetStep:
// Most of the fields require backend data to be populated (or are set based on cascading onChange events), so most of the fields have
// doNotAutoRestore or requiresBackendData set TRUE; we rely on the event handlers to populate the fields based on stored data AFTER the
// backend data arrives.
// The choice of whether to do existing or custom is the first trigger field
