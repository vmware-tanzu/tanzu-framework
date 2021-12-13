import { Component } from '@angular/core';
import { SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { VSphereWizardFormService } from '../../../../shared/service/vsphere-wizard-form.service';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-vsphere-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class VsphereOsImageStepComponent extends SharedOsImageStepComponent {
    private tkrVersionString: string;

    constructor(private vSphereWizardFormService: VSphereWizardFormService) {
        super();
        this.tkrVersion.subscribe(value => { this.tkrVersionString = value; });
    }

    // NOTE: there is an inherent assumption here that the tkrVersion Observable will have delivered a value before
    // setProviderInputs() is called (so that the usage below will be valid)
    protected setProviderInputs() {
        this.wizardFormService = this.vSphereWizardFormService;
        this.eventType = TkgEventType.VSPHERE_GET_OS_IMAGES;
        this.enableNonTemplateAlert = true;
        this.noImageAlertMessage = 'Your ' + this.clusterTypeDescriptor + ' cluster will be deployed with Tanzu Kubernetes release (TKr)' +
            ' ' + this.tkrVersionString +
            '. We are unable to detect a VM template that belongs to this Tanzu Kubernetes release. You must install ' +
            'a VM template that belongs to Tanzu Kubernetes release ' +  this.tkrVersionString + ' to continue with deployment of' +
            ' the ' + this.clusterTypeDescriptor + ' cluster. You may click the refresh icon to reload the OS image list once the ' +
            'appropriate VM template has been installed.';
        this.osImageTooltipContent = 'Select a VM template for a base OS image that you have already imported ' +
        ' into vSphere. If no compatible template is present, import one into vSphere and click the Refresh button.';
    }
}
