// Angular modules
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

// App imports
import { AzureWizardComponent } from './azure-wizard.component';

export const routes: Routes = [
    {
        path: '',
        component: AzureWizardComponent
    }
];

/**
 * @module AzureWizardRoutingModule
 * @description
 * This is routing module for the wizard module.
 */
@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class AzureWizardRoutingModule {}
