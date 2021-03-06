import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureClouds, AzureField } from '../azure-wizard.constants';

export const AzureProviderStepMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.PROVIDER_AZURECLOUD, required: true, defaultValue: AzureClouds[0], label: 'AZURE ENVIRONMENT',
            backingObject: {displayField: 'displayName', valueField: 'name'} },
        { name: AzureField.PROVIDER_CLIENT, required: true, label: 'CLIENT ID' },
        { name: AzureField.PROVIDER_CLIENTSECRET, required: true, mask: true, label: 'CLIENT SECRET' },
        { name: AzureField.PROVIDER_REGION, required: true, requiresBackendData: true, label: 'REGION' },
        { name: AzureField.PROVIDER_RESOURCEGROUPEXISTING, required: true, requiresBackendData: true, label: 'EXISTING RESOURCE GROUP' },
        { name: AzureField.PROVIDER_RESOURCEGROUPCUSTOM, doNotAutoRestore: true, label: 'RESOURCE GROUP NAME' },
        { name: AzureField.PROVIDER_RESOURCEGROUPOPTION, required: true, label: 'RESOURCE GROUP' },
        { name: AzureField.PROVIDER_SSHPUBLICKEY, required: true, label: 'SSH PUBLIC KEY' },
        { name: AzureField.PROVIDER_SUBSCRIPTION, required: true, label: 'SUBSCRIPTION ID' },
        { name: AzureField.PROVIDER_TENANT, required: true, label: 'TENANT ID' },
    ]
}

// About AzureProviderStep:
// We expect the user to fill out:
// TENANTID, CLIENTID, CLIENTSECRET, SUBSCRIPTIONID and AZURECLOUD and then click CONNECT.
// After that the REGION listbox should be populated and the user can select a resource group (or create one).
// We therefore do not want to restore a selected value to the REGION listbox until the user connects and the backend populates the backing
// data array. Likewise with the existing group resource id -- it should not be restored until the list of existing resources has been
// received from the backend.
// We also do not want to prematurely restore the value to the custom resource group (when creating the form initially), because that
// previously-custom group may have become an existing group, and we won't be able to detect that situation until we get the list of
// groups, and in addition, it should only be set if the user wants to create a new group
