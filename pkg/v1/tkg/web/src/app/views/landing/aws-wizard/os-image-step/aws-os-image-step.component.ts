import { Component, OnInit } from '@angular/core';
import { OsImageProviderInputs, SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { AwsWizardFormService } from '../../../../shared/service/aws-wizard-form.service';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AWSVirtualMachine } from '../../../../swagger/models';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';

@Component({
    selector: 'app-aws-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AwsOsImageStepComponent extends SharedOsImageStepComponent<AWSVirtualMachine> implements OnInit {
    constructor(private awsWizardFormService: AwsWizardFormService, protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    ngOnInit() {
        super.onInit();
    }

    protected supplyProviderInputs(): OsImageProviderInputs<AWSVirtualMachine> {
        return {
            event: TkgEventType.AWS_GET_OS_IMAGES,
            osImageService: this.awsWizardFormService,
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your AWS account. If no compatible OS image is present, import one into ' +
                'AWS and click the Refresh button'
        };
    }
}
