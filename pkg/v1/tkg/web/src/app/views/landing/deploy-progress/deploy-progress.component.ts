// Angular imports
import { Component, OnInit } from '@angular/core';
// Third party imports
import { BehaviorSubject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { LogMessage as NgxLogMessage } from 'ngx-log-monitor';
// App imports
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import AppServices from 'src/app/shared/service/appServices';
import { BasicSubscriber } from '../../../shared/abstracts/basic-subscriber';
import { EditionData } from '../../../shared/service/branding.service';
import { FormMetaDataStore } from '../wizard/shared/FormMetaDataStore';
import { TanzuEvent, TanzuEventType } from "../../../shared/service/Messenger";
import { WebsocketService } from '../../../shared/service/websocket.service';

@Component({
    selector: 'tkg-kickstart-ui-deploy-progress',
    templateUrl: './deploy-progress.component.html',
    styleUrls: ['./deploy-progress.component.scss']
})
export class DeployProgressComponent extends BasicSubscriber implements OnInit {

    providerType: string = '';
    cli: string = '';
    pageTitle: string = '';
    clusterTypeDescriptor: string;
    messages: any[] = [];
    msgs$ = new BehaviorSubject<NgxLogMessage>(null);
    curStatus: any = {
        msg: '',
        status: '',
        curPhase: '',
        finishedCount: 0,
        totalCount: 0,
    };

    APP_ROUTES: Routes = APP_ROUTES;
    phases: Array<string> = [];
    currentPhaseIdx: number;

    constructor(private websocketService: WebsocketService) {
        super();
        AppServices.messenger.subscribe(TanzuEventType.CLI_CHANGED, event => { this.cli = event.payload; }, this.unsubscribe);
    }

    ngOnInit(): void {
        this.initWebSocket();

        AppServices.messenger.subscribe<EditionData>(TanzuEventType.BRANDING_CHANGED, data => {
                this.pageTitle = data.payload.branding.title;
                this.clusterTypeDescriptor = data.payload.clusterTypeDescriptor;
            }, this.unsubscribe);

        AppServices.appDataService.getProviderType()
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((provider) => {
                if (provider && provider.includes('vsphere')) {
                    this.providerType = 'vSphere';
                } else if (provider && provider.includes('aws')) {
                    this.providerType = 'AWS';
                } else if (provider && provider.includes('azure')) {
                    this.providerType = 'Azure';
                } else if (provider && provider.includes('docker')) {
                    this.providerType = 'Docker';
                }
            });
    }

    initWebSocket() {
        this.websocketService.connect()
            .subscribe(evt => {
                const processed = this.processData(JSON.parse(evt.data));
                if (processed) {
                    this.msgs$.next(processed as NgxLogMessage);
                    this.messages.push(processed);
                }
            });

        setTimeout(_ => {
            this.websocketService.sendMessage('logs');
        }, 100);

        this.websocketService.setOnClose(_ => {
            if (this.curStatus.status !== 'successful' && this.curStatus.status !== 'failed') {
                setTimeout(() => {
                    this.initWebSocket();
                }, 5000);
            }
        });
    }

    /**
     * @method convert log type
     * @param {string} logType
     * @return string
     * 'ERROR' -> 'ERR'
     * 'FATAL' -> 'ERR'
     * 'INFO' -> 'INFO'
     * 'WARN' -> 'WARN'
     * 'UNKNOWN' -> null
     */
    convertLogType(logType: string): string {
        if (logType === 'ERROR') {
            return 'ERR';
        } else if (logType === 'FATAL') {
            return 'ERR';
        } else if (logType === 'UNKNOWN') {
            return null;
        } else {
            return logType;
        }
    }

    /**
     * @method process websocket data
     *  if data is a line of log, push to logs array
     *  if data is status update, update deployment status
     * @param {object} data websocket entry from backend
     */
    processData(data) {
        if (data.type === 'log') {
            this.curStatus.curPhase = data.data.currentPhase || this.curStatus.curPhase;
            return {
                message: data.data.message.slice(21),
                type: this.convertLogType(data.data.logType),
                timestamp: data.data.message.slice(1, 20)
            };
        } else {
            this.curStatus.msg = data.data.message;
            this.curStatus.status = data.data.status;

            this.phases = data.data.totalPhases || [];
            if (data.data.currentPhase && this.phases.length) {
                this.curStatus.curPhase = data.data.currentPhase;
                this.currentPhaseIdx = this.phases.indexOf(this.curStatus.curPhase);
            }

            if (this.curStatus.status === 'successful') {
                this.curStatus.finishedCount = this.curStatus.totalCount;
                this.currentPhaseIdx = this.phases.length;
                FormMetaDataStore.deleteAllSavedData();
            } else if (this.curStatus.status !== 'failed') {
                this.curStatus.finishedCount = Math.max(0, data.data.totalPhases.indexOf(this.curStatus.curPhase));
            }

            this.curStatus.totalCount = data.data.totalPhases ? data.data.totalPhases.length : 0;
            return null;
        }
    }

    /**
     * @method getStepCurrentState
     * @param idx - the index of each step in the ngFor expression
     * helper method determines which state should be displayed for each
     * step of the timeline component
     */
    getStepCurrentState(idx) {
        if (idx === this.currentPhaseIdx && this.curStatus.status === 'failed') {
            return 'error';
        } else if (idx < this.currentPhaseIdx || this.curStatus.status === 'successful') {
            return 'success'
        } else if (idx === this.currentPhaseIdx) {
            return 'current';
        } else {
            return 'not-started'
        }
    }

    /**
     * @method getStatusDescription
     * generates page description text depending on edition and status
     * @return {string}
     */
    getStatusDescription(): string {
        if (this.curStatus.status === 'running') {
            return `Deployment of the ${this.pageTitle} ${this.clusterTypeDescriptor} cluster to ${this.providerType} is in progress.`;
        } else if (this.curStatus.status === 'successful') {
            return `Deployment of the ${this.pageTitle} ${this.clusterTypeDescriptor} cluster to ${this.providerType} is successful.`;
        } else if (this.curStatus.status === 'failed') {
            return `Deployment of the ${this.pageTitle} ${this.clusterTypeDescriptor} cluster to ${this.providerType} has failed.`;
        }
    }
}
