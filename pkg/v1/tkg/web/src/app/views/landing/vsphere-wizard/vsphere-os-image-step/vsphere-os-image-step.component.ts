// Angular imports
import { Component, OnInit } from '@angular/core';

// Library imports
import { VSphereVirtualMachine } from 'tanzu-mgmt-plugin-api-lib';

// App imports
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { OsImageProviderInputs, SharedOsImageStepDirective } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-vsphere-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class VsphereOsImageStepComponent extends SharedOsImageStepDirective<VSphereVirtualMachine> {
    private tkrVersionString: string;

    constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
        this.tkrVersion.subscribe(value => { this.tkrVersionString = value; });
    }

    // NOTE: there is an implicit assumption here that the tkrVersion Observable will have delivered a value before
    // setProviderInputs() is called (so that the usage below will be valid)
    protected supplyProviderInputs(): OsImageProviderInputs {
        const noImageAlertMessage = 'Your ' + this.clusterTypeDescriptor + ' cluster will be deployed with Tanzu Kubernetes release (TKr)' +
            ' ' + this.tkrVersionString +
            '. We are unable to detect a VM template that belongs to this Tanzu Kubernetes release. You must install ' +
            'a VM template that belongs to Tanzu Kubernetes release ' +  this.tkrVersionString + ' to continue with deployment of' +
            ' the ' + this.clusterTypeDescriptor + ' cluster. You may click the refresh icon to reload the OS image list once the ' +
            'appropriate VM template has been installed.';
        const osImageTooltipContent = 'Select a VM template for a base OS image that you have already imported ' +
            ' into vSphere. If no compatible template is present, import one into vSphere and click the Refresh button.';
        const nonTemplateAlertMessage = 'Your selected OS image must be converted to a VM template. ' +
            'You may click the refresh icon to reload the OS image list once this has been done.'
        return {
            event: TkgEventType.VSPHERE_GET_OS_IMAGES,
            noImageAlertMessage: noImageAlertMessage,
            osImageTooltipContent: osImageTooltipContent,
            nonTemplateAlertMessage: nonTemplateAlertMessage,
        };
    }
}
