import { DockerLdapCommon } from './docker-ldap-common';
import { VtaasCommon } from '../common/vtaas-common';
import vtaas_list from '../common/vtaas_list';

const steps = vtaas_list.stepper;
const lists = vtaas_list.aws_step_list;
const testname = 'TKGm Docker(Ldap) UI Accessibility Test';
const wizardldap = new DockerLdapCommon();
const vtaasExe = new VtaasCommon(steps, lists, testname);

wizardldap.getFlowTestingDescription = () => {
    return "Docker vTaas flow (Ldap)"
};
wizardldap.executeAll(true);
vtaasExe.executeVtaas();
