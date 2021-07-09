// Angular modules
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

// App imports
import { AwsWizardComponent } from './aws-wizard.component';

export const routes: Routes = [
    {
        path: '',
        component: AwsWizardComponent
    }
];

/**
 * @module AwsWizardRoutingModule
 * @description
 * This is routing module for the wizard module.
 */
@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class AwsWizardRoutingModule {}
