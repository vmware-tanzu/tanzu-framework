// Angular modules
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

// App imports
import { VSphereWizardComponent } from './vsphere-wizard.component';

export const routes: Routes = [
    {
        path: '',
        component: VSphereWizardComponent
    }
];

/**
 * @module VSphereWizardRoutingModule
 * @description
 * This is routing module for the wizard module.
 */
@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class VSphereWizardRoutingModule {}
