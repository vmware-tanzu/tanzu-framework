import { Component, OnInit } from '@angular/core';
import { OsImageProviderInputs, SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AWSVirtualMachine } from '../../../../swagger/models';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import ServiceBroker from '../../../../shared/service/service-broker';
import { APIClient } from '../../../../swagger';
import { Observable } from 'rxjs';
import Broker from '../../../../shared/service/broker';

@Component({
    selector: 'app-aws-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AwsOsImageStepComponent extends SharedOsImageStepComponent<AWSVirtualMachine> implements OnInit {
    // aws globals
    region: string;

    constructor(protected fieldMapUtilities: FieldMapUtilities, protected serviceBroker: ServiceBroker, private apiClient: APIClient) {
        super(fieldMapUtilities, serviceBroker);
    }

    ngOnInit() {
        super.onInit();

        Broker.messenger.getSubject(TkgEventType.AWS_REGION_CHANGED)
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
