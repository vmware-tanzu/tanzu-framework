import { NewVpcCommon } from './aws-new-vpc-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.aws_step_list;
const testname = 'TKGm AWS(new vpc) UI Accessibility Test';
const wizardldap = new NewVpcCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardldap.getFlowTestingDescription = () => {
    return "AWS vTaas flow (New Vpc)"
};
wizardldap.executeAll(true);
vtaasExe.executeVtaas();
