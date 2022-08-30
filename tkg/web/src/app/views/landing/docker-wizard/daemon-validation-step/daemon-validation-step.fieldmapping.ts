import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';

export const DaemonStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'isConnected', validators: [SimpleValidator.IS_TRUE], isBoolean: true, doNotAutoRestore: true,
            label: 'DOCKER DAEMON CONNECTED' }
    ]
}
// About DaemonStep:
// When building the form initially, we do NOT care whether the user previously connected (we want to check again NOW), so we set
// doNotAutoRestore to TRUE. (However, we DO want to store the value for later display, which is why we DON'T use neverStore.)
