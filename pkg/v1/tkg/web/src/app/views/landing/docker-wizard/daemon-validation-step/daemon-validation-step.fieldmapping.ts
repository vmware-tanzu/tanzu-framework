import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const DaemonStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'isConnected', validators: [SimpleValidator.IS_TRUE] }
    ]
}
