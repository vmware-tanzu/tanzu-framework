import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereField } from '../vsphere-wizard.constants';
import { IpFamilyEnum } from '../../../../shared/constants/app.constants';

export const VsphereProviderStepFieldMapping: StepMapping = {
    fieldMappings: [
        { name: VsphereField.PROVIDER_CONNECTION_INSECURE, isBoolean: true, defaultValue: false, label: 'VSPHERE INSECURE CONNECTION' },
        { name: VsphereField.PROVIDER_DATA_CENTER, required: true, label: 'DATACENTER', requiresBackendData: true },
        { name: VsphereField.PROVIDER_IP_FAMILY, defaultValue: IpFamilyEnum.IPv4, featureFlag: 'vsphereIPv6', label: 'IP FAMILY' },
        { name: VsphereField.PROVIDER_SSH_KEY, required: true, label: 'SSH PUBLIC KEY' },
        { name: VsphereField.PROVIDER_SSH_KEY_FILE, doNotAutoSave: true },
        { name: VsphereField.PROVIDER_THUMBPRINT, label: 'SSL THUMBPRINT' },
        { name: VsphereField.PROVIDER_USER_NAME, required: true, label: 'USERNAME' },
        { name: VsphereField.PROVIDER_USER_PASSWORD, required: true, mask: true, label: 'PASSWORD' },
        { name: VsphereField.PROVIDER_VCENTER_ADDRESS, required: true, label: 'VCENTER SERVER' },
    ]
}
// About VsphereProviderStep:
// The user is expect to fill in all the fields EXCEPT data center, and then connect. Once the connection is made, the backend data on
// existing data centers is retrieved, and the handler for that event sets the data center field value from stored data. Therefore, we
// set requiresBackendData TRUE for the data center field (because the handler takes care of it and the value should NOT be set when the
// field is first added to the form).
