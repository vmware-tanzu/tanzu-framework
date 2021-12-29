import { AwsField, VpcType } from '../aws-wizard.constants';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';

export const AwsVpcStepMapping: StepMapping = {
    fieldMappings: [
        { name: AwsField.VPC_TYPE, required: true, defaultValue: VpcType.EXISTING },
        { name: AwsField.VPC_NEW_CIDR },
        { name: AwsField.VPC_EXISTING_CIDR },
        { name: AwsField.VPC_EXISTING_ID },
        { name: AwsField.VPC_NON_INTERNET_FACING, defaultValue: false, isBoolean: true },
    ]
}
