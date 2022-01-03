// Angular imports
import { Component, OnInit } from '@angular/core';
// App imports
import { AzureVirtualMachine } from '../../../../swagger/models';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import { OsImageProviderInputs, SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import ServiceBroker from '../../../../shared/service/service-broker';
import { TkgEventType } from '../../../../shared/service/Messenger';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepComponent<AzureVirtualMachine> implements OnInit {
    constructor(protected fieldMapUtilities: FieldMapUtilities, protected serviceBroker: ServiceBroker) {
        super(fieldMapUtilities, serviceBroker);
    }

    ngOnInit() {
        super.onInit();
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
