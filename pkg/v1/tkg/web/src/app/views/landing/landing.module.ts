// Angular modules
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

// Third party modules
import { LogMonitorModule } from 'ngx-log-monitor';

// App imports
import { AlertNotificationComponent } from '../../shared/components/alert-notification/alert-notification.component';
import { LandingComponent } from './landing.component';
import { LandingRoutingModule } from './landing-routing.module';
import { StartComponent } from './start/start.component';
import { ClusterClassInfoComponent } from './wizard/shared/components/widgets/cluster-class-info/cluster-class-info.component';
import { ConfirmationComponent } from './confirm/confirmation.component';
import { DeployProgressComponent } from './deploy-progress/deploy-progress.component';
import { WcpRedirectComponent } from './wcp-redirect/wcp-redirect.component';
import { IncompatibleComponent } from './incompatible/incompatible.component';
import { SharedModule } from '../../shared/shared.module';
import { PreviewConfigComponent } from '../../shared/components/preview-config/preview-config.component';
import { VmwCopyToClipboardButtonComponent } from '../../shared/components/copy-to-clipboard-button/copy-to-clipboard-button.component';
import { ErrorNotificationComponent } from "../../shared/components/error-notification/error-notification.component";

@NgModule({
    declarations: [
        AlertNotificationComponent,
        ConfirmationComponent,
        DeployProgressComponent,
        ErrorNotificationComponent,
        IncompatibleComponent,
        LandingComponent,
        PreviewConfigComponent,
        StartComponent,
        VmwCopyToClipboardButtonComponent,
        WcpRedirectComponent
    ],
    imports: [
        CommonModule,
        LandingRoutingModule,
        LogMonitorModule,
        SharedModule
    ],
    exports: [
        AlertNotificationComponent,
        ClusterClassInfoComponent,
        ConfirmComponent,
        ConfirmationComponent,
        ErrorNotificationComponent
    ]
})

export class LandingModule {}
