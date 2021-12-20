import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';
import { IpFamilyEnum } from '../../../../shared/constants/app.constants';

export const VsphereProviderStepFieldMapping: StepMapping = {
    fieldMappings: [
        { name: VsphereField.PROVIDER_CONNECTION_INSECURE, isBoolean: true, defaultValue: false },
        { name: VsphereField.PROVIDER_DATA_CENTER, required: true },
        { name: VsphereField.PROVIDER_IP_FAMILY, defaultValue: IpFamilyEnum.IPv4, required: true },
        { name: VsphereField.PROVIDER_SSH_KEY, required: true },
        { name: VsphereField.PROVIDER_SSH_KEY_FILE, required: true },
        { name: VsphereField.PROVIDER_THUMBPRINT, required: true },
        { name: VsphereField.PROVIDER_USER_NAME, required: true },
        { name: VsphereField.PROVIDER_USER_PASSWORD, required: true },
        { name: VsphereField.PROVIDER_VCENTER_ADDRESS, required: true },
    ]
}
