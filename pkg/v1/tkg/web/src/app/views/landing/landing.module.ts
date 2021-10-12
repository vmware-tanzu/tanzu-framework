// Angular modules
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

// Third party modules
import { LogMonitorModule } from 'ngx-log-monitor';

// App imports
import { LandingComponent } from './landing.component';
import { LandingRoutingModule } from './landing-routing.module';
import { StartComponent } from './start/start.component';
import { ConfirmComponent } from './confirm/confirm.component';
import { DeployProgressComponent } from './deploy-progress/deploy-progress.component';
import { WcpRedirectComponent } from './wcp-redirect/wcp-redirect.component';
import { IncompatibleComponent } from './incompatible/incompatible.component';
import { SharedModule } from '../../shared/shared.module';
import { PreviewConfigComponent } from '../../shared/components/preview-config/preview-config.component';
import { VmwCopyToClipboardButtonComponent } from '../../shared/components/copy-to-clipboard-button/copy-to-clipboard-button.component';
import { ErrorNotificationComponent } from "../../shared/components/error-notification/error-notification.component";
import { ErrorImportFileComponent } from "../../shared/components/error-import-file/error-import-file.component";
import { ImportFileSuccessNotificationComponent } from "../../shared/components/import-file-success-notification/import-file-success-notification.component";
@NgModule({
    declarations: [
        LandingComponent,
        StartComponent,
        ConfirmComponent,
        DeployProgressComponent,
        WcpRedirectComponent,
        IncompatibleComponent,
        VmwCopyToClipboardButtonComponent,
        PreviewConfigComponent,
        ErrorNotificationComponent,
        ErrorImportFileComponent,
        ImportFileSuccessNotificationComponent
    ],
    imports: [
        CommonModule,
        LandingRoutingModule,
        LogMonitorModule,
        SharedModule
    ],
    exports: [
        ConfirmComponent,
        ErrorNotificationComponent,
        ErrorImportFileComponent,
        ImportFileSuccessNotificationComponent
    ]
})

export class LandingModule {}
