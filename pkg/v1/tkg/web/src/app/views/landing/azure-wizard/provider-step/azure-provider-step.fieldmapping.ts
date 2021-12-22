import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { AzureField, ResourceGroupOption } from '../azure-wizard.constants';

export const AzureProviderStepMapping: StepMapping = {
    fieldMappings: [
        { name: AzureField.PROVIDER_AZURECLOUD, required: true },
        { name: AzureField.PROVIDER_CLIENT, required: true },
        { name: AzureField.PROVIDER_CLIENTSECRET, required: true },
        { name: AzureField.PROVIDER_REGION, required: true },
        { name: AzureField.PROVIDER_RESOURCEGROUPEXISTING, required: true },
        { name: AzureField.PROVIDER_RESOURCEGROUPCUSTOM },
        { name: AzureField.PROVIDER_RESOURCEGROUPOPTION, required: true },
        { name: AzureField.PROVIDER_SSHPUBLICKEY, required: true },
        { name: AzureField.PROVIDER_SUBSCRIPTION, required: true },
        { name: AzureField.PROVIDER_TENANT, required: true },
    ]
}
