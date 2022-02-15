import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';

export const VsphereResourceStepMapping: StepMapping = {
    fieldMappings: [
        { name: VsphereField.RESOURCE_VMFOLDER, required: true, requiresBackendData: true, label: 'VM FOLDER' },
        { name: VsphereField.RESOURCE_DATASTORE, required: true, requiresBackendData: true, label: 'DATASTORE' },
        { name: VsphereField.RESOURCE_POOL, required: true, requiresBackendData: true, label: 'CLUSTERS, HOSTS, AND RESOURCE POOLS' },
    ]
}
// About VsphereResourceStep:
// All these values depend on backend data being available, so we don't auto-restore the value on initialization. Rather, we rely on
// the event handlers that handle the backend data arriving to set the field values from stored data at that time.
