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
import { StepWrapperComponent } from './step-wrapper/step-wrapper.component';
import { VsphereOsImageStepComponent } from '../../vsphere-wizard/vsphere-os-image-step/vsphere-os-image-step.component';
import { StepWrapperSetComponent } from './step-wrapper/step-wrapper-set.component';
import { AwsOsImageStepComponent } from '../../aws-wizard/os-image-step/aws-os-image-step.component';
import { AzureOsImageStepComponent } from '../../azure-wizard/os-image-step/azure-os-image-step.component';
import { FieldMapUtilities } from './field-mapping/FieldMapUtilities';
import { VsphereNetworkStepComponent } from '../../vsphere-wizard/vsphere-network-step/vsphere-network-step.component';

@NgModule({
    declarations: [
        StepFormNotificationComponent,
        StepControllerComponent,
        StepWrapperSetComponent,
        StepWrapperComponent,
        SharedCeipStepComponent,
        SharedNetworkStepComponent,
        SharedLoadBalancerStepComponent,
        MetadataStepComponent,
        DeleteDataPopupComponent,
        SSLThumbprintModalComponent,
        SharedIdentityStepComponent,
        TreeSelectComponent,
        AuditLoggingComponent,
        VsphereOsImageStepComponent,
        VsphereNetworkStepComponent,
        AwsOsImageStepComponent,
        AzureOsImageStepComponent
    ],
    imports: [
        CommonModule,
        SharedModule,
        CodemirrorModule
    ],
    exports: [
        StepFormNotificationComponent,
        StepControllerComponent,
        StepWrapperComponent,
        StepWrapperSetComponent,
        SharedCeipStepComponent,
        SharedNetworkStepComponent,
        SharedLoadBalancerStepComponent,
        MetadataStepComponent,
        DeleteDataPopupComponent,
        SSLThumbprintModalComponent,
        SharedIdentityStepComponent,
        TreeSelectComponent,
        AuditLoggingComponent,
        VsphereOsImageStepComponent,
        VsphereNetworkStepComponent,
        AwsOsImageStepComponent,
        AzureOsImageStepComponent
    ],
    providers: [
        FieldMapUtilities,
        ValidationService
    ]
})
export class WizardSharedModule { }
