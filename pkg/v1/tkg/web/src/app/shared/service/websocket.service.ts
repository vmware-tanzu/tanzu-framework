// Angular imports
import { Injectable } from '@angular/core';
import { Observable, Subject, Observer } from 'rxjs';

// App imports
import { PROVIDERS } from '../constants/app.constants';
import { environment } from '../../../environments/environment';

@Injectable({
    providedIn: 'root'
})
export class WebsocketService {
    private ws: WebSocket;
    private subject: Subject<any>;

    constructor() {
    }

    connect(): Observable<any> {
        if (!this.subject) {
            this.subject = this.create();
        }
        return this.subject.asObservable();
    }

    create() {
        if (!window) {
            throw new Error("The service should only be used a web application.");
        }

        const { protocol, host } = window.location;
        const url = (protocol === "https" ? "wss" : "ws") + `://${host}/ws`;
        this.ws = new WebSocket(url);

        const observable = Observable.create(
            (obs: Observer<MessageEvent>) => {
                this.ws.onmessage = obs.next.bind(obs);
                this.ws.onerror = obs.error.bind(obs);
                this.ws.onclose = obs.complete.bind(obs);

                return this.ws.close.bind(this.ws);
            });

        const observer = {
            next: (data: Object) => {
                if (this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(JSON.stringify(data));
                }
            }
        };

        return Subject.create(observer, observable);
    }

    sendMessage(message: string) {
        const payload = { operation: message };
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(payload));
        } else {
            console.log('ws connection not yet open');
        }
    }

    setOnClose(handler) {
        this.subject = null;
        this.ws.onclose = handler;
    }
}
