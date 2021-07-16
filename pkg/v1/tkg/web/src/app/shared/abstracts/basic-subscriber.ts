import { OnDestroy } from "@angular/core";
import { Subject } from 'rxjs';
import Broker from "../service/broker";
import { TkgEventType } from "../service/Messenger";

/**
 * Base class tha should be extended by any class that want to have a flag indicating if the component
 * is still alive. This is useful to handle Observable subscriptions in order to unsubscribe. Instead
 * of doing unsubscribe() you can use operator .takeWhile(() => this.isAlive).
 */
export abstract class BasicSubscriber implements OnDestroy {

    protected isAlive = true;
    protected unsubscribe: Subject<void> = new Subject();

    ngOnDestroy() {
        this.isAlive = false;
        this.unsubscribe.next();
        this.unsubscribe.complete();
    }

    showContextualHelp(title: string, keywords: Array<string>) {
        Broker.messenger.publish({
            type: TkgEventType.OPEN_CONTEXTUAL_HELP,
            payload: {
                title,
                keywords
            }
        })
    }

}
