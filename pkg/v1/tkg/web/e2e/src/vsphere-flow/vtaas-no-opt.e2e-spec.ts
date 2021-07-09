import { DisableOptionsCommon } from './vsphere-no-opt-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.vsphere_step_list;
const testname = 'TKGm Vsphere(oidc) UI Accessibility Test';
const wizardoidc = new DisableOptionsCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardoidc.getFlowTestingDescription = () => {
    return "Vsphere vTaas flow (OIDC)"
};
wizardoidc.executeAll(true);
vtaasExe.executeVtaas();
