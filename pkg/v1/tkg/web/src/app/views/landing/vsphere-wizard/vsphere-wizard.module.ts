// Angular modules
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

// App imports
import { LandingModule } from '../landing.module';
import { VSphereWizardRoutingModule } from './vsphere-wizard-routing.module';
import { SharedModule } from '../../../shared/shared.module';
import { WizardSharedModule } from '../wizard/shared/wizard-shared.module';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { VSphereWizardComponent } from './vsphere-wizard.component';
import { VSphereProviderStepComponent } from './provider-step/vsphere-provider-step.component';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { ResourceStepComponent } from './resource-step/resource-step.component'
@NgModule({
    declarations: [
        VSphereWizardComponent,
        VSphereProviderStepComponent,
        NodeSettingStepComponent,
        ResourceStepComponent
    ],
    imports: [
        CommonModule,
        VSphereWizardRoutingModule,
        SharedModule,
        LandingModule,
        WizardSharedModule
    ],
    exports: [
    ],
    providers: [
        ValidationService
    ]
})
export class WizardModule { }
