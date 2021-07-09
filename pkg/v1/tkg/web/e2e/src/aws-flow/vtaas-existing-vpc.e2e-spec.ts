import { ExistingVpcCommon } from './aws-existing-vpc-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.aws_step_list;
const testname = 'TKGm AWS(existing vpc) UI Accessibility Test';
const wizardoidc = new ExistingVpcCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardoidc.getFlowTestingDescription = () => {
    return "AWS vTaas flow (Existing Vpc)"
};
wizardoidc.executeAll(true);
vtaasExe.executeVtaas();
