import { Component, OnInit } from '@angular/core';
import { SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { VSphereWizardFormService } from '../../../../shared/service/vsphere-wizard-form.service';
import { AwsWizardFormService } from '../../../../shared/service/aws-wizard-form.service';
import { AzureWizardFormService } from '../../../../shared/service/azure-wizard-form.service';
import Broker from '../../../../shared/service/broker';
import { BehaviorSubject } from 'rxjs';
import { Observable } from 'rxjs/internal/Observable';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-aws-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AwsOsImageStepComponent extends SharedOsImageStepComponent {
    constructor(private awsWizardFormService: AwsWizardFormService) {
        super();
    }

    ngOnInit() {
        super.onInit();
    }

    protected setProviderInputs() {
        this.wizardFormService = this.awsWizardFormService;
        this.eventType = TkgEventType.AWS_GET_OS_IMAGES;
        this.enableNonTemplateAlert = false;
        this.osImageTooltipContent = 'Select a base OS image that you have already imported ' +
        'into your AWS account. If no compatible OS image is present, import one into ' +
        'AWS and click the Refresh button';
    }
}
