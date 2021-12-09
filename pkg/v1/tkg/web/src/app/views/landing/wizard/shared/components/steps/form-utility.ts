import { WizardBaseDirective } from '../../wizard-base/wizard-base';
import { IdentityManagementType, WizardForm } from '../../constants/wizard.constants';

interface I18nDataForHtml {
    title: string,
    description: string,
}
export interface FormDataForHTML {
    name: string,
    description: string,
    title: string,
    i18n: I18nDataForHtml,
}

export class FormUtility {
    static IdentityFormDescription(wizard: WizardBaseDirective): string {
        const identityType = wizard.getFieldValue(WizardForm.IDENTITY, 'identityType');
        const ldapEndpointIp = wizard.getFieldValue(WizardForm.IDENTITY, 'endpointIp');
        const ldapEndpointPort = wizard.getFieldValue(WizardForm.IDENTITY, 'endpointPort');
        const oidcIssuer = wizard.getFieldValue(WizardForm.IDENTITY, 'issuerURL');

        if (identityType === IdentityManagementType.OIDC && oidcIssuer) {
            return 'OIDC configured: ' + oidcIssuer;
        } else if (identityType === IdentityManagementType.LDAP && ldapEndpointIp) {
            return 'LDAP configured: ' + ldapEndpointIp + ':' + ldapEndpointPort;
        }
        return 'Specify identity management';
    }

    static MetadataFormDescription(wizard: WizardBaseDirective): string {
        const clusterLocation = wizard.getFieldValue(WizardForm.METADATA, 'clusterLocation');
        return clusterLocation ? 'Location: ' + clusterLocation : 'Specify metadata for the ' + wizard.clusterTypeDescriptor + ' cluster';

    }

    static NetworkFormDescription(wizard: WizardBaseDirective): string {
        const serviceCidr = wizard.getFieldValue(WizardForm.NETWORK, "clusterServiceCidr");
        const podCidr = wizard.getFieldValue(WizardForm.NETWORK, "clusterPodCidr");
        if (serviceCidr && podCidr) {
            return `Cluster service CIDR: ${serviceCidr} Cluster POD CIDR: ${podCidr}`;
        }
        return "Specify how TKG networking is provided and global network settings";
    }

    static OsImageFormDescription(wizard: WizardBaseDirective): string {
        if (wizard.getFieldValue(WizardForm.OSIMAGE, 'osImage') && wizard.getFieldValue(WizardForm.OSIMAGE, 'osImage').name) {
            return 'OS Image: ' + wizard.getFieldValue(WizardForm.OSIMAGE, 'osImage').name;
        }
        return 'Specify the OS Image';
    }

    static titleCase(target): string {
        return target.replace(/(^|\s)\S/g, function(t) { return t.toUpperCase() });
    }

    static formOverrideDescription(formData: FormDataForHTML, description: string): FormDataForHTML {
        formData.description = description;
        return formData;
    }
}

