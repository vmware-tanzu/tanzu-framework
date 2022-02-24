// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import { AzureVirtualMachine } from '../../../../swagger/models';
import { OsImageProviderInputs, SharedOsImageStepDirective } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TanzuEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepDirective<AzureVirtualMachine> implements OnInit {
    protected supplyProviderInputs(): OsImageProviderInputs {
        return {
            event: TanzuEventType.AZURE_GET_OS_IMAGES,
            noImageAlertMessage: '',
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your Azure account. If no compatible OS image is present, import one into ' +
                'Azure and click the Refresh button',
        };
    }

    protected supplyImportFileSuccessEvent(): TanzuEventType {
        return TanzuEventType.AZURE_CONFIG_FILE_IMPORTED;
    }

    protected supplyImportFileFailureEvent(): TanzuEventType {
        return TanzuEventType.AZURE_CONFIG_FILE_IMPORT_ERROR;
    }
}
