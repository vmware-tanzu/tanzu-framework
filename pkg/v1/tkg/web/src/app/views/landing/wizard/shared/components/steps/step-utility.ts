import { WizardBaseDirective } from '../../wizard-base/wizard-base';
import { IdentityManagementType, WizardForm, WizardStep } from '../../constants/wizard.constants';

export class StepUtility {
    static IdentityStepDescription(wizard: WizardBaseDirective): string {
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

    static MetadataStepDescription(wizard: WizardBaseDirective): string {
        const clusterLocation = wizard.getFieldValue(WizardForm.METADATA, 'clusterLocation');
        return clusterLocation ? 'Location: ' + clusterLocation : 'Specify metadata for the ' + wizard.clusterTypeDescriptor + ' cluster';

    }
    
    static NetworkStepDescription(wizard: WizardBaseDirective): string {
        const serviceCidr = wizard.getFieldValue(WizardForm.NETWORK, "clusterServiceCidr");
        const podCidr = wizard.getFieldValue(WizardForm.NETWORK, "clusterPodCidr");
        if (serviceCidr && podCidr) {
            return `Cluster service CIDR: ${serviceCidr} Cluster POD CIDR: ${podCidr}`;
        }
        return "Specify how TKG networking is provided and global network settings";
    }
    
    static OsImageStepDescription(wizard: WizardBaseDirective): string {
        if (wizard.getFieldValue(WizardForm.OSIMAGE, 'osImage') && wizard.getFieldValue(WizardForm.OSIMAGE, 'osImage').name) {
            return 'OS Image: ' + wizard.getFieldValue(WizardForm.OSIMAGE, 'osImage').name;
        }         
        return 'Specify the OS Image';
    }
    
    static CommonStepDescription(step: string, wizard: WizardBaseDirective): string {
        if (step === WizardStep.NETWORK) {
            return StepUtility.NetworkStepDescription(wizard);
        } else if (step === WizardStep.METADATA) {
            return StepUtility.MetadataStepDescription(wizard);
        } else if (step === WizardStep.IDENTITY) {
            return StepUtility.IdentityStepDescription(wizard);
        } else if (step === WizardStep.OSIMAGE) {
            return StepUtility.OsImageStepDescription(wizard);
        }
        console.log('WARNING: Unrecognized step passed to CommonStepDescription(): ' + step);
        return 'Step ' + step;
    }
}
