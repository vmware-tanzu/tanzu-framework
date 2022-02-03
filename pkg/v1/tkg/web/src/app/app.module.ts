// Angular imports
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

// Third-party imports
import { LogMonitorModule } from 'ngx-log-monitor';
import { CodemirrorModule } from '@ctrl/ngx-codemirror';

// Module imports
import { APIClientModule } from 'tanzu-management-cluster-ng-api';
import { AppRoutingModule } from './app-routing.module';
import { SharedModule } from './shared/shared.module';

// Component imports
import { AppComponent } from './app.component';
import { HeaderBarModule } from './shared/components/header-bar/header-bar.module';
import { ThemeToggleComponent } from './shared/components/theme-toggle/theme-toggle.component';

// Service imports
import { BrandingService } from './shared/service/branding.service';
import { WebsocketService } from './shared/service/websocket.service';

@NgModule({
    declarations: [
        AppComponent,
        ThemeToggleComponent
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
        BrandingService,
        WebsocketService
    ],
    bootstrap: [AppComponent]
})
export class AppModule { }
