import { Component } from '@angular/core';
import { SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { VSphereWizardFormService } from '../../../../shared/service/vsphere-wizard-form.service';
import { AwsWizardFormService } from '../../../../shared/service/aws-wizard-form.service';
import { AzureWizardFormService } from '../../../../shared/service/azure-wizard-form.service';
import Broker from '../../../../shared/service/broker';
import { BehaviorSubject } from 'rxjs';
import { Observable } from 'rxjs/internal/Observable';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepComponent {
    constructor(private azureWizardFormService: AzureWizardFormService, protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    ngOnInit() {
        super.onInit();
    }

    protected setProviderInputs() {
        this.wizardFormService = this.azureWizardFormService;
        this.eventType = TkgEventType.AZURE_GET_OS_IMAGES;
        this.osImageTooltipContent = 'Select a base OS image that you have already imported ' +
        'into your Azure account. If no compatible OS image is present, import one into ' +
        'Azure and click the Refresh button';
    }
}
