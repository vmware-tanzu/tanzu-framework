import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { WizardSharedModule } from './../wizard/shared/wizard-shared.module';
import { SharedModule } from '../../../shared/shared.module';
import { LandingModule } from '../landing.module';
import { AwsWizardRoutingModule } from './aws-wizard-routing.module';

import { AwsWizardComponent } from './aws-wizard.component';
import { NodeSettingStepComponent } from './node-setting-step/node-setting-step.component';
import { AwsProviderStepComponent } from './provider-step/aws-provider-step.component';
import { VpcStepComponent } from './vpc-step/vpc-step.component';

import { ValidationService } from '../wizard/shared/validation/validation.service';

@NgModule({
    declarations: [
        AwsWizardComponent,
        AwsProviderStepComponent,
        VpcStepComponent,
        NodeSettingStepComponent
    ],
    imports: [
        CommonModule,
        AwsWizardRoutingModule,
        SharedModule,
        LandingModule,
        WizardSharedModule
    ],
    providers: [
        ValidationService
    ]
})
export class AwsWizardModule { }
export enum AwsStep {
    PROVIDER = 'provider',
    VPC = 'vpc',
    NODESETTING = 'nodeSetting',
    NETWORK = 'network',
    METADATA = 'metadata',
    IDENTITY = 'identity',
    OSIMAGE = 'osImage'
}
export enum AwsForm {
    PROVIDER = 'awsProviderForm',
    VPC = 'vpcForm',
    NODESETTING = 'awsNodeSettingForm',
    NETWORK = 'networkForm',
    METADATA = 'metadataForm',
    IDENTITY = 'identityForm',
    OSIMAGE = 'osImageForm'
}
