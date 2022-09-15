import { AwsField, CredentialType } from '../aws-wizard.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const AwsProviderStepMapping: StepMapping = {
    fieldMappings: [
        { name: AwsField.PROVIDER_AUTH_TYPE, required: true, defaultValue: CredentialType.PROFILE, primaryTrigger: true,
            label: 'AWS CREDENTIAL TYPE' },
        { name: AwsField.PROVIDER_ACCESS_KEY, mask: true, label: 'ACCESS KEY ID (OPTIONAL)', doNotAutoRestore: true },
        { name: AwsField.PROVIDER_SECRET_ACCESS_KEY, mask: true, label: 'SECRET ACCESS KEY (OPTIONAL)', doNotAutoRestore: true },
        { name: AwsField.PROVIDER_SESSION_TOKEN, label: 'SESSION TOKEN (OPTIONAL)', doNotAutoRestore: true },
        { name: AwsField.PROVIDER_PROFILE_NAME, label: 'AWS CREDENTIAL PROFILE', doNotAutoRestore: true },
        { name: AwsField.PROVIDER_REGION, required: true, label: 'REGION', doNotAutoRestore: true },
    ]
}
// About AwsProviderStep:
// The first thing the user must do is select what kind of credentials they are using (AwsField.PROVIDER_AUTH_TYPE), and this choic
// triggers the UI to display the corresponding fields. Therefore, we set primaryTrigger to TRUE for AwsField.PROVIDER_AUTH_TYPE.
// All the other fields will then be restored based on the handler of AwsField.PROVIDER_AUTH_TYPE changes, so we set their
// configuration with doNotAutoRestore TRUE, since we don't want to build the form originally with any values set for those fields.
