// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import { AWSVirtualMachine } from '../../../../swagger/models';
import AppServices from '../../../../shared/service/appServices';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { OsImageProviderInputs, SharedOsImageStepDirective } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-aws-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AwsOsImageStepComponent extends SharedOsImageStepDirective<AWSVirtualMachine> implements OnInit {
    // aws globals
    region: string;

    constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    ngOnInit() {
        super.ngOnInit();

        AppServices.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
            .subscribe(event => {
                this.region = event.payload;
            });
    }

    protected supplyProviderInputs(): OsImageProviderInputs {
        return {
            event: TkgEventType.AWS_GET_OS_IMAGES,
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your AWS account. If no compatible OS image is present, import one into ' +
                'AWS and click the Refresh button'
        };
    }
}
