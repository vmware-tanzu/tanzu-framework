import { ElementFinder } from 'protractor';
import { Stepper } from './step-expand.po';

const stepper = new Stepper();

const aws_step_list = [ stepper.getProviderStep(), stepper.getVpcStep(),
    stepper.getNodeSettingsStep(), stepper.getMetadataStep(),
    stepper.getNetworkStep(), stepper.getIdentityStep(),
    stepper.getTmcStep(), stepper.getCeipStep()
];

const azure_step_list = [ stepper.getProviderStep(), stepper.getVnetStep(),
    stepper.getNodeSettingsStep(), stepper.getMetadataStep(),
    stepper.getNetworkStep(), stepper.getIdentityStep(),
    stepper.getTmcStep(), stepper.getCeipStep()
];

const vsphere_step_list = [ stepper.getProviderStep(),
    stepper.getNodeSettingsStep(), stepper.getNsxLbStep(),
    stepper.getMetadataStep(), stepper.getResourceStep(),
    stepper.getNetworkStep(), stepper.getIdentityStep(),
    stepper.getOsImageStep(), stepper.getTmcStep(),
    stepper.getCeipStep()
]

export default {
    stepper,
    aws_step_list,
    azure_step_list,
    vsphere_step_list
}
