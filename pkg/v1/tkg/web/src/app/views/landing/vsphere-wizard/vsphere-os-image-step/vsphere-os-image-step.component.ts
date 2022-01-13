// Angular imports
import { Component } from '@angular/core';
// App imports
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { OsImageProviderInputs, SharedOsImageStepDirective } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TanzuEventType } from '../../../../shared/service/Messenger';
import { StepMapping } from '../../wizard/shared/field-mapping/FieldMapping';
import { VsphereOsImageStepMapping } from './vsphere-os-image-step.fieldmapping';
import { VSphereVirtualMachine } from '../../../../swagger/models';

@Component({
    selector: 'app-vsphere-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class VsphereOsImageStepComponent extends SharedOsImageStepDirective<VSphereVirtualMachine> {
    private tkrVersionString: string;

    constructor() {
        super();
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
            event: TanzuEventType.VSPHERE_GET_OS_IMAGES,
            noImageAlertMessage: noImageAlertMessage,
            osImageTooltipContent: osImageTooltipContent,
            nonTemplateAlertMessage: nonTemplateAlertMessage,
        };
    }

    protected getImageFromStoredValue(osImageValue: string): VSphereVirtualMachine {
        // NOTE: we are switching to use the MOID as the stored value. However, older config files will have the image name.
        // we therefore find the first image that matches the saved value on either the MOID or the name
        return this.osImages.find(image => image.moid === osImageValue || image.name === osImageValue);
    }

    protected supplyStepMapping(): StepMapping {
        return VsphereOsImageStepMapping;
    }

    protected supplyImportFileSuccessEvent(): TanzuEventType {
        return TanzuEventType.VSPHERE_CONFIG_FILE_IMPORTED;
    }

    protected supplyImportFileFailureEvent(): TanzuEventType {
        return TanzuEventType.VSPHERE_CONFIG_FILE_IMPORT_ERROR;
    }
}
