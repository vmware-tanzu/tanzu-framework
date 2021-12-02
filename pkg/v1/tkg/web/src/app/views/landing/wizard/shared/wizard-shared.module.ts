import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { SharedModule } from '../../../../shared/shared.module';
import { ValidationService } from './validation/validation.service';
import { StepFormNotificationComponent } from './step-form-notification/step-form-notification.component';
import { StepControllerComponent } from './step-controller/step-controller.component';
import { SharedCeipStepComponent } from './components/steps/ceip-step/ceip-step.component';
import { SharedNetworkStepComponent } from './components/steps/network-step/network-step.component';
import { SharedLoadBalancerStepComponent } from './components/steps/load-balancer/load-balancer-step.component';
import { MetadataStepComponent } from './components/steps/metadata-step/metadata-step.component';
import { CodemirrorModule } from '@ctrl/ngx-codemirror';
import { DeleteDataPopupComponent } from './components/delete-data-popup.component';
import { SSLThumbprintModalComponent } from './components/modals/ssl-thumbprint-modal/ssl-thumbprint-modal.component';
import { SharedIdentityStepComponent } from './components/steps/identity-step/identity-step.component';
import { TreeSelectComponent } from './tree-select/tree-select.component';
import { AuditLoggingComponent } from './components/widgets/audit-logging/audit-logging.component';
import { SharedOsImageStepComponent } from './components/steps/os-image-step/os-image-step.component';

@NgModule({
    declarations: [
        StepFormNotificationComponent,
        StepControllerComponent,
        SharedCeipStepComponent,
        SharedNetworkStepComponent,
        SharedLoadBalancerStepComponent,
        MetadataStepComponent,
        DeleteDataPopupComponent,
        SSLThumbprintModalComponent,
        SharedIdentityStepComponent,
        TreeSelectComponent,
        AuditLoggingComponent,
        SharedOsImageStepComponent
    ],
    imports: [
        CommonModule,
        SharedModule,
        CodemirrorModule
    ],
    exports: [
        StepFormNotificationComponent,
        StepControllerComponent,
        SharedCeipStepComponent,
        SharedNetworkStepComponent,
        SharedLoadBalancerStepComponent,
        MetadataStepComponent,
        DeleteDataPopupComponent,
        SSLThumbprintModalComponent,
        SharedIdentityStepComponent,
        TreeSelectComponent,
        AuditLoggingComponent,
        SharedOsImageStepComponent
    ],
    providers: [
        ValidationService
    ]
})
export class WizardSharedModule { }
