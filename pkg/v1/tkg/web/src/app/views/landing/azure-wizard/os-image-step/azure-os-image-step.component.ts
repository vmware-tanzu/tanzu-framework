import { Component, OnInit } from '@angular/core';
import { OsImageProviderInputs, SharedOsImageStepComponent } from '../../wizard/shared/components/steps/os-image-step/os-image-step.component';
import { AzureWizardFormService } from '../../../../shared/service/azure-wizard-form.service';
import { TkgEventType } from '../../../../shared/service/Messenger';
import { AzureVirtualMachine } from '../../../../swagger/models';
import { FieldMapUtilities } from '../../wizard/shared/field-mapping/FieldMapUtilities';
import ServiceBroker from '../../../../shared/service/service-broker';
import { APIClient } from '../../../../swagger';
import { Observable } from 'rxjs';
import Broker from '../../../../shared/service/broker';

@Component({
    selector: 'app-azure-os-image-step',
    templateUrl: '../../wizard/shared/components/steps/os-image-step/os-image-step.component.html',
    styleUrls: ['../../wizard/shared/components/steps/os-image-step/os-image-step.component.scss']
})
export class AzureOsImageStepComponent extends SharedOsImageStepComponent<AzureVirtualMachine> implements OnInit {
    constructor(protected fieldMapUtilities: FieldMapUtilities, protected serviceBroker: ServiceBroker, private apiClient: APIClient) {
        super(fieldMapUtilities, serviceBroker);
    }

    ngOnInit() {
        super.onInit();
    }

    protected supplyProviderInputs(): OsImageProviderInputs<AzureVirtualMachine> {
        return {
            event: TkgEventType.AZURE_GET_OS_IMAGES,
            noImageAlertMessage: '',
            fetcher: this.fetchOsImages.bind(this),
            osImageTooltipContent: 'Select a base OS image that you have already imported ' +
                'into your Azure account. If no compatible OS image is present, import one into ' +
                'Azure and click the Refresh button',
        };
    }

    private fetchOsImages(data: any): Observable<AzureVirtualMachine[]> {
        return this.apiClient.getAzureOSImages();
    }
}
