// Angular modules
import { Component, OnInit } from '@angular/core';

// App imports
import { SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { AzureWizardFormService } from '../../../../shared/service/azure-wizard-form.service';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepComponent implements OnInit {
    constructor(private azureWizardFormService: AzureWizardFormService) {
        super();
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
