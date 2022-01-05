import { FieldMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';

export const VsphereNetworkFieldMappings: FieldMapping[] = [
        { name: VsphereField.NETWORK_NAME, label: 'NETWORK NAME', requiresBackendData: true,
                backingObject: { displayField: 'displayName', valueField: 'name', type: 'VsphereNetwork' } },
]
