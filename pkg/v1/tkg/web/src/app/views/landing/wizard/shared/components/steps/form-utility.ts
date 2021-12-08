import { WizardBaseDirective } from '../../wizard-base/wizard-base';
import { IdentityManagementType, WizardForm, WizardStep } from '../../constants/wizard.constants';

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

    static CommonStepDescription(step: string, wizard: WizardBaseDirective): string {
        if (step === WizardStep.NETWORK) {
            return FormUtility.NetworkFormDescription(wizard);
        } else if (step === WizardStep.METADATA) {
            return FormUtility.MetadataFormDescription(wizard);
        } else if (step === WizardStep.IDENTITY) {
            return FormUtility.IdentityFormDescription(wizard);
        } else if (step === WizardStep.OSIMAGE) {
            return FormUtility.OsImageFormDescription(wizard);
        }
        console.log('WARNING: Unrecognized step passed to CommonStepDescription(): ' + step);
        return 'Step ' + step;
    }
}
