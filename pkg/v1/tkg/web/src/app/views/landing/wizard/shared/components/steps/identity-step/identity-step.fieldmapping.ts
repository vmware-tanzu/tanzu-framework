import { StepMapping } from '../../../field-mapping/FieldMapping';
import { IdentityManagementType } from '../../../constants/wizard.constants';

export enum IdentityField {
    BIND_DN = 'bindDN',
    BIND_PW = 'bindPW',
    CLIENT_ID = 'clientId',
    CLIENT_SECRET = 'clientSecret',
    ENDPOINT_IP = 'endpointIp',
    ENDPOINT_PORT = 'endpointPort',
    GROUP_SEARCH_BASE_DN = 'groupSearchBaseDN',
    GROUP_SEARCH_FILTER = 'groupSearchFilter',
    GROUP_SEARCH_GROUP_ATTR = 'groupSearchGroupAttr',
    GROUP_SEARCH_NAME_ATTR = 'groupSearchNameAttr',
    GROUP_SEARCH_USER_ATTR = 'groupSearchUserAttr',
    IDENTITY_TYPE = 'identityType',
    IDM_SETTINGS = 'idmSettings',
    ISSUER_URL = 'issuerURL',
    LDAP_ROOT_CA = 'ldapRootCAData',
    OIDC_GROUPS_CLAIM = 'oidcGroupsClaim',
    OIDC_USERNAME_CLAIM = 'oidcUsernameClaim',
    SCOPES = 'scopes',
    TEST_GROUP_NAME = 'testGroupName',
    TEST_USER_NAME = 'testUserName',
    USER_SEARCH_BASE_DN = 'userSearchBaseDN',
    USER_SEARCH_FILTER = 'userSearchFilter',
    USER_SEARCH_USERNAME = 'userSearchUsername',
}

export const IdentityStepMapping: StepMapping = {
    fieldMappings: [
        { name: IdentityField.IDENTITY_TYPE, defaultValue: IdentityManagementType.OIDC },
        { name: IdentityField.IDM_SETTINGS, isBoolean: true, defaultValue: true },
        { name: IdentityField.ISSUER_URL },
        { name: IdentityField.CLIENT_ID },
        { name: IdentityField.CLIENT_SECRET },
        { name: IdentityField.SCOPES },
        { name: IdentityField.OIDC_GROUPS_CLAIM },
        { name: IdentityField.OIDC_USERNAME_CLAIM },
        { name: IdentityField.ENDPOINT_IP },
        { name: IdentityField.ENDPOINT_PORT },
        { name: IdentityField.BIND_PW },
        { name: IdentityField.GROUP_SEARCH_FILTER },
        { name: IdentityField.USER_SEARCH_FILTER },
        { name: IdentityField.USER_SEARCH_USERNAME },
        { name: IdentityField.BIND_DN },
        { name: IdentityField.USER_SEARCH_BASE_DN },
        { name: IdentityField.GROUP_SEARCH_BASE_DN },
        { name: IdentityField.GROUP_SEARCH_USER_ATTR },
        { name: IdentityField.GROUP_SEARCH_GROUP_ATTR },
        { name: IdentityField.GROUP_SEARCH_NAME_ATTR },
        { name: IdentityField.LDAP_ROOT_CA },
        { name: IdentityField.TEST_USER_NAME },
        { name: IdentityField.TEST_GROUP_NAME },
    ]
}
