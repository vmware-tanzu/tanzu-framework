// Angular imports
import { Component } from '@angular/core';

// Library imports
import { AzureVirtualMachine } from 'tanzu-mgmt-plugin-api-lib';

// App imports
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { OsImageProviderInputs, SharedOsImageStepDirective } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepDirective<AzureVirtualMachine> {
    constructor(protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    protected supplyProviderInputs(): OsImageProviderInputs {
        return {
            event: TkgEventType.AZURE_GET_OS_IMAGES,
            noImageAlertMessage: '',
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your Azure account. If no compatible OS image is present, import one into ' +
                'Azure and click the Refresh button',
        };
    }
}
