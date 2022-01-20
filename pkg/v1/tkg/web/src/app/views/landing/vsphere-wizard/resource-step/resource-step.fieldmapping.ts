import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';

export const VsphereResourceStepMapping: StepMapping = {
    fieldMappings: [
        { name: VsphereField.RESOURCE_POOL, required: true },
        { name: VsphereField.RESOURCE_DATASTORE, required: true },
        { name: VsphereField.RESOURCE_VMFOLDER, required: true },
    ]
}
