// Angular module
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { DockerWizardComponent } from './docker-wizard.component';

// App imports
export const routes: Routes = [
    {
        path: '',
        component: DockerWizardComponent
    }
];

/**
 * @module DockerWizardRoutingModule
 * @description
 * This is routing module for the wizard module.
 */
@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class DockerWizardRoutingModule {}
