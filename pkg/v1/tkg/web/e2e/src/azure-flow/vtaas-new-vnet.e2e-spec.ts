import { NewVnetCommon } from './azure-new-vnet-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.azure_step_list;
const testname = 'TKGm Azure(new vnet) UI Accessibility Test';
const wizardldap = new NewVnetCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardldap.getFlowTestingDescription = () => {
    return "Azure vTaas flow (New Vnet)"
};
wizardldap.executeAll(true);
vtaasExe.executeVtaas();
