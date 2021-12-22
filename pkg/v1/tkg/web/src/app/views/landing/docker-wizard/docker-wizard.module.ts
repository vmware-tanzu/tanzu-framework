// Angular modules
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

// App imports
import { DockerWizardComponent } from './docker-wizard.component';
import { DockerWizardRoutingModule } from './docker-wizard-routing.module';
import { LandingModule } from '../landing.module';
import { SharedModule } from 'src/app/shared/shared.module';
import { WizardSharedModule } from '../wizard/shared/wizard-shared.module';
import { DaemonValidationStepComponent } from './daemon-validation-step/daemon-validation-step.component';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';

@NgModule({
    declarations: [DockerWizardComponent, DaemonValidationStepComponent, NodeSettingStepComponent],
    imports: [
        CommonModule,
        DockerWizardRoutingModule,
        SharedModule,
        WizardSharedModule,
        LandingModule
    ]
})
export class DockerWizardModule { }
