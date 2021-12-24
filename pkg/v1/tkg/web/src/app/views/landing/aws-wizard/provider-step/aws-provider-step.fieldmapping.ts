import { AwsField, CredentialType } from '../aws-wizard.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const AwsProviderStepMapping: StepMapping = {
    fieldMappings: [
        { name: AwsField.PROVIDER_ACCESS_KEY },
        { name: AwsField.PROVIDER_AUTH_TYPE, required: true, defaultValue: CredentialType.PROFILE },
        { name: AwsField.PROVIDER_PROFILE_NAME },
        { name: AwsField.PROVIDER_REGION, required: true },
        { name: AwsField.PROVIDER_SECRET_ACCESS_KEY },
        { name: AwsField.PROVIDER_SESSION_TOKEN },
    ]
}
