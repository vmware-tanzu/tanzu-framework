// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import AppServices from '../../../../shared/service/appServices';
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
    private currentRegion: string;

    ngOnInit() {
        super.ngOnInit();
        // subscribe to the AWS_REGION_CHANGED event so that we can capture the current region. We need it so that if
        // the user wants to refresh the os images available, we have the region id for the backend call to get the os images
        AppServices.messenger.subscribe<string>(TanzuEventType.AWS_REGION_CHANGED, event => {
            this.currentRegion = event.payload;
        });
    }

    protected supplyProviderInputs(): OsImageProviderInputs {
        return {
            createOsImageEventPayload: this.createOsImageEventPayload.bind(this),
            event: TanzuEventType.AWS_GET_OS_IMAGES,
            eventImportFileFailure: TanzuEventType.AWS_CONFIG_FILE_IMPORT_ERROR,
            eventImportFileSuccess: TanzuEventType.AWS_CONFIG_FILE_IMPORTED,
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your AWS account. If no compatible OS image is present, import one into ' +
                'AWS and click the Refresh button'
        };
    }

    // returns a payload that can be used with the AWS_GET_OS_IMAGES event to refresh the os images from the backend
    private createOsImageEventPayload() {
        return { region: this.currentRegion };
    }
}
