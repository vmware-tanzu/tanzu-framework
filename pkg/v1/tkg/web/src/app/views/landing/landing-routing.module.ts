// Angular modules
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

// App imports
import { LandingComponent } from './landing.component';
import { StartComponent } from './start/start.component';
import { DeployProgressComponent } from './deploy-progress/deploy-progress.component';
import { WcpRedirectComponent } from './wcp-redirect/wcp-redirect.component';
import { IncompatibleComponent } from './incompatible/incompatible.component';

export const routes: Routes = [
    {
        path: '',
        component: LandingComponent,
        children: [
            {
                path: '',
                component: StartComponent
            },
            {
                path: 'wizard',
                loadChildren: () => import('./vsphere-wizard/vsphere-wizard.module').then(m => m.WizardModule)
            },
            {
                path: 'aws/wizard',
                loadChildren: () => import('./aws-wizard/aws-wizard.module').then(m => m.AwsWizardModule)
            },
            {
                path: 'azure/wizard',
                loadChildren: () => import('./azure-wizard/azure-wizard.module').then(m => m.AzureWizardModule)
            },
            {
                path: 'docker/wizard',
                loadChildren: () => import('./docker-wizard/docker-wizard.module').then(m => m.DockerWizardModule)
            },
            {
                path: 'deploy-progress',
                component: DeployProgressComponent
            },
            {
                path: 'vsphere-with-kubernetes',
                component: WcpRedirectComponent
            },
            {
                path: 'incompatible',
                component: IncompatibleComponent
            }
        ]
    }
];

/**
 * @module LandingRoutingModule
 * @description
 * This is routing module for the landing module.
 */
@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class LandingRoutingModule {}
