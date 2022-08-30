import { DockerLdapCommon } from './docker-ldap-common';

const wizardldap = new DockerLdapCommon();
wizardldap.executeAll(false);
