// App imports
import { AwsField, VpcType } from '../aws-wizard.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const AwsVpcStepMapping: StepMapping = {
    fieldMappings: [
        { name: AwsField.VPC_TYPE, primaryTrigger: true },
        { name: AwsField.VPC_NEW_CIDR, label: 'VPC CIDR' },
        { name: AwsField.VPC_EXISTING_CIDR, label: 'VPC CIDR', doNotAutoRestore: true },
        { name: AwsField.VPC_EXISTING_ID, label: 'VPC ID', requiresBackendData: true },
        { name: AwsField.VPC_NON_INTERNET_FACING, defaultValue: false, isBoolean: true, label: 'THIS IS NOT AN INTERNET FACING VPC' },
    ]
}
// About AwsVpcStep:
// The first trigger field is whether the user wants and existing or new VPC.
// The VPC_EXISTING_ID field requires data from the back end, so it is requiresBackendData to prevent the value from being set before the
// backend data arrives. We rely on the handlers of the backend-data-arrived event to set the field value based on stored data.
// VPC_EXISTING_CIDR field is auto-set based on the VPC_EXISTING_ID field
