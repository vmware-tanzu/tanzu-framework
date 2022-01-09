import { Component, Input } from '@angular/core';

export enum NotificationTypes {
    INFO = 'info',
    SUCCESS = 'success',
    WARNING = 'warning',
    ERROR = 'error'
}

export interface Notification {
    notificationType: string;
    message: string;
}

@Component({
    selector: 'app-alert-notification',
    templateUrl: './alert-notification.component.html'
})
export class AlertNotificationComponent {
    @Input() notification: Notification;

    isInfo(): boolean {
        return this.notification.notificationType === NotificationTypes.INFO;
    }

    isSuccess(): boolean {
        return this.notification.notificationType === NotificationTypes.SUCCESS;
    }

    isWarning(): boolean {
        return this.notification.notificationType === NotificationTypes.WARNING;
    }

    isError(): boolean {
        return this.notification.notificationType === NotificationTypes.ERROR;
    }
}
