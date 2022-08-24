import { SimpleValidator } from '../../wizard/shared/constants/validation.constants';
import { VsphereField } from '../vsphere-wizard.constants';

export const KUBE_VIP = 'Kube-vip';
export const NSX_ADVANCED_LOAD_BALANCER = "NSX Advanced Load Balancer";

export const VsphereNodeSettingFieldMappings = [
    { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_PROVIDER, required: true, defaultValue: KUBE_VIP,
        label: 'CONTROL PLANE ENDPOINT PROVIDER' },
    { name: VsphereField.NODESETTING_CONTROL_PLANE_ENDPOINT_IP, required: true, validators: [SimpleValidator.IS_VALID_FQDN_OR_IP],
        label: 'CONTROL PLANE ENDPOINT' },
];
// About VsphereNodeSettingStandaloneStep:
// Extends common NodeSettingStep
