import { StepMapping } from '../../../field-mapping/FieldMapping';

export const IdentityStepMapping: StepMapping = {
    fieldMappings: [
        { name: 'identityType', defaultValue: 'oidc' },
        { name: 'idmSettings', isBoolean: true, defaultValue: true },
        { name: 'issuerURL' },
        { name: 'clientId' },
        { name: 'clientSecret' },
        { name: 'scopes' },
        { name: 'oidcUsernameClaim' },
        { name: 'oidcGroupsClaim' },
        { name: 'endpointIp' },
        { name: 'endpointPort' },
        { name: 'bindPW' },
        { name: 'groupSearchFilter' },
        { name: 'userSearchFilter' },
        { name: 'userSearchUsername' },
        { name: 'bindDN' },
        { name: 'userSearchBaseDN' },
        { name: 'groupSearchBaseDN' },
        { name: 'groupSearchUserAttr' },
        { name: 'groupSearchGroupAttr' },
        { name: 'groupSearchNameAttr' },
        { name: 'ldapRootCAData' },
        { name: 'testUserName' },
        { name: 'testGroupName' },
    ]
}
