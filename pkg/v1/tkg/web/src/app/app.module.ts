// Angular imports
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

// Third-party imports
import { LogMonitorModule } from 'ngx-log-monitor';
import { CodemirrorModule } from '@ctrl/ngx-codemirror';

// Module imports
import { APIClientModule } from './swagger/index';
import { AppRoutingModule } from './app-routing.module';
import { SharedModule } from './shared/shared.module';

// Component imports
import { AppComponent } from './app.component';
import { HeaderBarModule } from './shared/components/header-bar/header-bar.module';

// Service imports
import { AppDataService } from './shared/service/app-data.service';
import { BrandingService } from './shared/service/branding.service';
import { WebsocketService } from './shared/service/websocket.service';
import { VSphereWizardFormService } from './shared/service/vsphere-wizard-form.service';

@NgModule({
    declarations: [
        AppComponent
    ],
    imports: [
        BrowserModule,
        AppRoutingModule,
        LogMonitorModule,
        BrowserAnimationsModule,
        HeaderBarModule,
        APIClientModule.forRoot({
            domain: '',
            httpOptions: {
                headers: {
                    'Content-Type': 'application/json'
                }
            }
        }),
        SharedModule,
        CodemirrorModule
    ],
    providers: [
        AppDataService,
        BrandingService,
        WebsocketService
    ],
    bootstrap: [AppComponent]
})
export class AppModule { }
