// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import { AWSVirtualMachine } from '../../../../swagger/models';
import {
    OsImageProviderInputs,
    SharedOsImageStepDirective
} from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TanzuEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-aws-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AwsOsImageStepComponent extends SharedOsImageStepDirective<AWSVirtualMachine> implements OnInit {
    protected supplyProviderInputs(): OsImageProviderInputs {
        return {
            event: TanzuEventType.AWS_GET_OS_IMAGES,
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your AWS account. If no compatible OS image is present, import one into ' +
                'AWS and click the Refresh button'
        };
    }

    protected supplyImportFileSuccessEvent(): TanzuEventType {
        return TanzuEventType.AWS_CONFIG_FILE_IMPORTED;
    }

    protected supplyImportFileFailureEvent(): TanzuEventType {
        return TanzuEventType.AWS_CONFIG_FILE_IMPORT_ERROR;
    }
}
