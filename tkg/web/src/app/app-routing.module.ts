// Angular imports
import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

// App imports

const routes: Routes = [
    {
        path: '',
        redirectTo: 'ui',
        pathMatch: 'full'
    },
    {
        path: 'ui',
        loadChildren: () => import('./views/landing/landing.module').then(m => m.LandingModule)
    }
];

@NgModule({
    imports: [RouterModule.forRoot(routes, {useHash: true})],
    exports: [RouterModule]
})
export class AppRoutingModule { }
