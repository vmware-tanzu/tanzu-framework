import { ExistingVnetCommon } from './azure-existing-vnet-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.azure_step_list;
const testname = 'TKGm Azure(existing vnet) UI Accessibility Test';
const wizardoidc = new ExistingVnetCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardoidc.getFlowTestingDescription = () => {
    return "Azure vTaas flow (Existing Vnet)"
};
wizardoidc.executeAll(true);
vtaasExe.executeVtaas();
