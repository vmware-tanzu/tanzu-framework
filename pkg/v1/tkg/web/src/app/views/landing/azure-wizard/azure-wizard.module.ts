// Angular modules
import { NgModule } from '@angular/core';

// App imports
import { AzureWizardRoutingModule } from './azure-wizard-routing.moduel';
import { ValidationService } from '../wizard/shared/validation/validation.service';
import { AzureWizardComponent } from './azure-wizard.component';
import { AzureProviderStepComponent } from './provider-step/azure-provider-step.component';
import { CommonModule } from '@angular/common';
import { SharedModule } from 'src/app/shared/shared.module';
import { LandingModule } from '../landing.module';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { VnetStepComponent } from './vnet-step/vnet-step.component';
import { WizardSharedModule } from './../wizard/shared/wizard-shared.module';

@NgModule({
    declarations: [
        AzureWizardComponent,
        AzureProviderStepComponent,
        VnetStepComponent,
        NodeSettingStepComponent,
    ],
    imports: [
        CommonModule,
        AzureWizardRoutingModule,
        SharedModule,
        LandingModule,
        WizardSharedModule
    ],
    providers: [
        ValidationService
    ]
})
export class AzureWizardModule { }
