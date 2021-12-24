import { Component, OnInit } from '@angular/core';
import { OsImageProviderInputs, SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { AzureWizardFormService } from '../../../../shared/service/azure-wizard-form.service';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AzureVirtualMachine } from '../../../../swagger/models';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepComponent<AzureVirtualMachine> implements OnInit {
    constructor(private azureWizardFormService: AzureWizardFormService, protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    ngOnInit() {
        super.onInit();
    }

    protected supplyProviderInputs(): OsImageProviderInputs<AzureVirtualMachine> {
        return {
            event: TkgEventType.AZURE_GET_OS_IMAGES,
            noImageAlertMessage: '',
            osImageService: this.azureWizardFormService,
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your Azure account. If no compatible OS image is present, import one into ' +
                'Azure and click the Refresh button',
        };
    }
}
