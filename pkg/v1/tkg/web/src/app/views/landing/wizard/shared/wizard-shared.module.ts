// Angular imports
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
// App imports
import { AuditLoggingComponent } from './components/widgets/audit-logging/audit-logging.component';
import { AwsOsImageStepComponent } from '../../aws-wizard/os-image-step/aws-os-image-step.component';
import { AzureOsImageStepComponent } from '../../azure-wizard/os-image-step/azure-os-image-step.component';
import { CodemirrorModule } from '@ctrl/ngx-codemirror';
import { DeleteDataPopupComponent } from './components/delete-data-popup.component';
import { FieldMapUtilities } from './field-mapping/FieldMapUtilities';
import { MetadataStepComponent } from './components/steps/metadata-step/metadata-step.component';
import { SharedCeipStepComponent } from './components/steps/ceip-step/ceip-step.component';
import { SharedIdentityStepComponent } from './components/steps/identity-step/identity-step.component';
import { SharedLoadBalancerStepComponent } from './components/steps/load-balancer/load-balancer-step.component';
import { SharedModule } from '../../../../shared/shared.module';
import { SharedNetworkStepComponent } from './components/steps/network-step/network-step.component';
import { SSLThumbprintModalComponent } from './components/modals/ssl-thumbprint-modal/ssl-thumbprint-modal.component';
import { StepControllerComponent } from './step-controller/step-controller.component';
import { StepFormNotificationComponent } from './step-form-notification/step-form-notification.component';
import { StepWrapperComponent } from './step-wrapper/step-wrapper.component';
import { StepWrapperSetComponent } from './step-wrapper/step-wrapper-set.component';
import { TreeSelectComponent } from './tree-select/tree-select.component';
import { ValidationService } from './validation/validation.service';
import { VsphereNetworkStepComponent } from '../../vsphere-wizard/vsphere-network-step/vsphere-network-step.component';
import { VsphereOsImageStepComponent } from '../../vsphere-wizard/vsphere-os-image-step/vsphere-os-image-step.component';
import { VsphereLoadBalancerStepComponent } from '../../vsphere-wizard/load-balancer/vsphere-load-balancer-step.component';

@NgModule({
    declarations: [
        AuditLoggingComponent,
        AwsOsImageStepComponent,
        AzureOsImageStepComponent,
        DeleteDataPopupComponent,
        MetadataStepComponent,
        SSLThumbprintModalComponent,
        SharedCeipStepComponent,
        SharedIdentityStepComponent,
        SharedLoadBalancerStepComponent,
        SharedNetworkStepComponent,
        StepControllerComponent,
        StepFormNotificationComponent,
        StepWrapperComponent,
        StepWrapperSetComponent,
        TreeSelectComponent,
        VsphereLoadBalancerStepComponent,
        VsphereNetworkStepComponent,
        VsphereOsImageStepComponent
    ],
    imports: [
        CodemirrorModule,
        CommonModule,
        SharedModule
    ],
    exports: [
        AuditLoggingComponent,
        AwsOsImageStepComponent,
        AzureOsImageStepComponent,
        DeleteDataPopupComponent,
        MetadataStepComponent,
        SSLThumbprintModalComponent,
        SharedCeipStepComponent,
        SharedIdentityStepComponent,
        SharedLoadBalancerStepComponent,
        SharedNetworkStepComponent,
        StepControllerComponent,
        StepFormNotificationComponent,
        StepWrapperComponent,
        StepWrapperSetComponent,
        TreeSelectComponent,
        VsphereLoadBalancerStepComponent,
        VsphereNetworkStepComponent,
        VsphereOsImageStepComponent
    ],
    providers: [
        ValidationService
    ]
})
export class WizardSharedModule { }
