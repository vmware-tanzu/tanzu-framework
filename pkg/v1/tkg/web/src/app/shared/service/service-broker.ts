import { TkgEventType } from './Messenger';
import Broker from './broker';
import { Observable, ReplaySubject } from 'rxjs';
import { StepFormDirective } from '../../views/landing/wizard/shared/step-form/step-form';
import { takeUntil } from 'rxjs/operators';

interface ServiceBrokerEntry<OBJ> {
    fetcher: (data: any) => Observable<OBJ[]>,
    staticError?: string,
    dataStream: ReplaySubject<OBJ[]>,
    errorStream: ReplaySubject<string>,
}

export default class ServiceBroker {
    private entries: Map<TkgEventType, ServiceBrokerEntry<any>> = new Map<TkgEventType, ServiceBrokerEntry<any>>();

    // convenience wrapper to publish triggering events
    trigger(eventTypes: TkgEventType[], payload?: any) {
        eventTypes.forEach(eventType => { Broker.messenger.publish({type: eventType, payload: payload}) });
    }

    clear<OBJ>(eventType: TkgEventType) {
        const serviceBrokerEntry: ServiceBrokerEntry<OBJ> = this.getEntry<OBJ>(eventType);
        if (serviceBrokerEntry) {
            serviceBrokerEntry.dataStream.next([]);
        }
    }

    // register() is called by those providing services. This is typically done by wizards setting up how
    // to respond to data-request events, i.e. linking data-request events to API calls to the backend
    register<OBJ>(eventType: TkgEventType, fetcher: (data: any) => Observable<OBJ[]>, staticError?: string) {
        if (this.entries[eventType]) {
            console.warn('service broker detects duplicate registration of event ' + TkgEventType[eventType] + '; ignoring');
            return;
        }
        this.entries[eventType] = {
            fetcher: fetcher,
            staticError: staticError,
            errorStream: new ReplaySubject<string>(),
            dataStream: new ReplaySubject<OBJ[]>(),
        };
        // we subscribe to the messenger to ensure that whenever the target event is broadcast, we go fetch the data
        Broker.messenger.getSubject(eventType)
            .subscribe((event) => this.fetchData<OBJ>(eventType, event.payload));
    }

    // subscribe() is called by those consuming data services. This is typically a step that relies on whatever data
    // is returned from the backend (for example, giving the user a choice of networks, regions, datacenters, etc)
    subscribe<OBJ>(eventType: TkgEventType, onDataReceived: (data: OBJ[]) => void, onError: (error: string) => void): boolean {
        const serviceBrokerEntry: ServiceBrokerEntry<OBJ> = this.getEntry<OBJ>(eventType);
        if (!serviceBrokerEntry) {
            return false;
        }
        serviceBrokerEntry.dataStream.subscribe(onDataReceived);
        serviceBrokerEntry.errorStream.subscribe(onError);
        return true;
    }

    // This default error handler sets (or clears) the step's errorNotification field.
    private defaultStepErrorHandler(step: StepFormDirective): (error: string) => void {
        return (error: string) => {
            step.errorNotification = (error) ? error : '';
        };
    }

    // This is a convenience method for steps wanting to call with an error handler that appends to existing error messages.
    // When setting the error, it ADDS to any existing error, rather than overwriting it.
    // This is useful if the step expects to make several service calls at once.
    public appendingStepErrorHandler(step: StepFormDirective): (error: string) => void {
        return (error: string) => {
            if (!error) {
                step.errorNotification = '';
            } else if (! step.errorNotification.endsWith(error)) {  // don't append same error message twice in a row
                if (step.errorNotification) {
                    step.errorNotification = step.errorNotification + ' ';
                }
                step.errorNotification = step.errorNotification + error;
            }
        };
    }

    // convenience method to allow steps to register with a default error handler (namely, setting their errorNotification field),
    // as well as unsubscribing to the data stream with the ngOnDestroy event
    stepSubscribe<OBJ>(step: StepFormDirective, eventType: TkgEventType,
                       onDataReceived: (data: OBJ[]) => void, onError?: (error: string) => void): boolean {
        const serviceBrokerEntry: ServiceBrokerEntry<OBJ> = this.getEntry<OBJ>(eventType);
        if (!serviceBrokerEntry) {
            console.error('Event ' + eventType + ' was not registered with the service broker before ' + step.formName +
                ' tried to subscribe to it.');
            return false;
        }
        if (!onError) {
            onError = this.defaultStepErrorHandler(step);
        }
        serviceBrokerEntry.dataStream
            .pipe(takeUntil(step.unsubscribeOnDestroy))
            .subscribe(onDataReceived);
        serviceBrokerEntry.errorStream.subscribe(onError);
        return true;
    }

    // TODO: This method should currently only be called from outside by TESTS
    public simulateError(eventType: TkgEventType, errMsg: string) {
        this.getEntry(eventType).errorStream.next(errMsg);
    }

    // TODO: This method should currently only be called from outside by TESTS
    public simulateData(eventType: TkgEventType, data: any) {
        this.getEntry(eventType).dataStream.next(data);
    }

    private getEntry<OBJ>(eventType: TkgEventType): ServiceBrokerEntry<OBJ> {
        const result = this.entries[eventType];
        if (result) {
            return result;
        }
        console.error('ServiceBroker tried to get entry for event ' + eventType + ' but no such event has been registered');
        return null;
    }

    private fetchData<OBJ>(eventType: TkgEventType, fetcherPayload: any) {
        const entry = this.getEntry<OBJ>(eventType);
        entry.fetcher(fetcherPayload).subscribe(
            (data => {
                // we received data, so broadcast it to anyone listening (and clear any previous errors)
                entry.dataStream.next(data);
                entry.errorStream.next(null);
            }),
            (err => {
                // we received an error, so broadcast it to anyone listening (and clear any previous data)
                const errMsg = err.error && err.error.message ? err.error.message : null;
                const error = errMsg || err.message || JSON.stringify(err);
                let message = (entry.staticError ? entry.staticError + ' ' : '') + (error ? error : '');
                if (!message) {
                    message = 'Service error encountered';
                }
                entry.errorStream.next(message);
                entry.dataStream.next([]);
            })
        )
    }
}
