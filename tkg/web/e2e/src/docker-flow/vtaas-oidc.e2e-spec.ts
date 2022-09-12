import { DockerOidcCommon } from './docker-oidc-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.aws_step_list;
const testname = 'TKGm Docker(Oidc) UI Accessibility Test';
const wizardoidc = new DockerOidcCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardoidc.getFlowTestingDescription = () => {
    return "Docker vTaas flow (Oidc)"
};
wizardoidc.executeAll(true);
vtaasExe.executeVtaas();
