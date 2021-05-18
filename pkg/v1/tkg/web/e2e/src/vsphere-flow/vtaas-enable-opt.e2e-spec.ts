import { EnableOptionsCommon } from './vsphere-enable-opt-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.vsphere_step_list;
const testname = 'TKGm Vsphere(avi & ldap) UI Accessibility Test';
const wizardldap = new EnableOptionsCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardldap.getFlowTestingDescription = () => {
    return "Vsphere vTaas flow (LDAP)"
};
wizardldap.executeAll(true);
vtaasExe.executeVtaas();
